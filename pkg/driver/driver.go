package main

import (
	"context"
	"flag"
	"fmt"
	snapshotterapiv1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	snapshotterv1 "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned/typed/volumesnapshot/v1"
	"github.com/sambatv/k8s-snapshots/pkg/apis/snapshotrule/v1alpha1"
	"github.com/sambatv/k8s-snapshots/pkg/generated/clientset/versioned"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"sort"
	"strings"
	"time"
)

type SnapshotDriver struct {
	config            *rest.Config
	clientset         *kubernetes.Clientset
	versionedSnapshot *versioned.Clientset
	snapshotV1Client  *snapshotterv1.SnapshotV1Client

	snapshotClassName string
	update            bool
}

func NewSnapshotDriver(config *rest.Config) SnapshotDriver {
	var err error

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Sprintf("Cannot create clientset: %s", err.Error()))
	}

	versionedSnapshot, err := versioned.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	snapshotV1Client, err := snapshotterv1.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return SnapshotDriver{
		config:            config,
		clientset:         clientset,
		versionedSnapshot: versionedSnapshot,
		snapshotV1Client:  snapshotV1Client,
		snapshotClassName: "csi-aws-vsc",
	}
}

func (d SnapshotDriver) listSnapshots(namespace string) ([]snapshotterapiv1.VolumeSnapshot, error) {
	snapshotList, err := d.snapshotV1Client.VolumeSnapshots(namespace).List(
		context.Background(),
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, err
	}

	return snapshotList.Items, err
}

func (d SnapshotDriver) listPvc(namespace string) ([]v1.PersistentVolumeClaim, error) {
	persistentVolumeClient := d.clientset.CoreV1().PersistentVolumeClaims(namespace)

	claimList, err := persistentVolumeClient.List(
		context.Background(),
		metav1.ListOptions{},
	)

	if err == nil {
		return claimList.Items, nil
	}
	return nil, err
}

func (d SnapshotDriver) listSnapshotRules() ([]v1alpha1.SnapshotRule, error) {
	ctx := context.Background()
	ruleList, err := d.versionedSnapshot.K8ssnapshotsV1alpha1().SnapshotRules("").List(
		ctx,
		metav1.ListOptions{},
	)

	if err != nil {
		return nil, err
	}
	return ruleList.Items, nil
}

func (d SnapshotDriver) getPvcForRule(rule v1alpha1.SnapshotRule) ([]v1.PersistentVolumeClaim, error) {
	var matchLabels []string
	for key, value := range rule.Spec.Selector.MatchLabels {
		matchLabels = append(matchLabels, fmt.Sprintf("%s=%s", key, value))
	}

	if len(matchLabels) == 0 {
		log.Printf("Empty labels, cannot operate")
		return nil, fmt.Errorf("cannot fetch pvc, empty labels")
	}

	pvcList, err := d.clientset.CoreV1().PersistentVolumeClaims(rule.Namespace).List(
		context.Background(),
		metav1.ListOptions{
			LabelSelector: strings.Join(matchLabels, ", "),
		},
	)
	if err != nil {
		return nil, err
	}

	return pvcList.Items, err
}

func (d SnapshotDriver) createSnapshot(rule v1alpha1.SnapshotRule, pvc v1.PersistentVolumeClaim) error {
	currentTime := time.Now()
	snapshotName := fmt.Sprintf("%s-%s", pvc.Name, currentTime.Format("2006-01-02"))

	snapshotClassName := rule.Spec.SnapshotClassName
	if snapshotClassName != "" {
		snapshotClassName = d.snapshotClassName
	}

	snapshot := &snapshotterapiv1.VolumeSnapshot{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: rule.Namespace,
			Name:      snapshotName,
			Labels:    nil,
			Annotations: map[string]string{
				"pvc": pvc.Name,
			},
			OwnerReferences: nil,
			ResourceVersion: "",
		},
		Spec: snapshotterapiv1.VolumeSnapshotSpec{
			Source: snapshotterapiv1.VolumeSnapshotSource{
				PersistentVolumeClaimName: &pvc.Name,
				VolumeSnapshotContentName: nil,
			},
			VolumeSnapshotClassName: &snapshotClassName,
		},
		Status: nil,
	}

	ctx := context.Background()
	client := d.snapshotV1Client.VolumeSnapshots(rule.Namespace)

	volumeSnapshot, err := client.Get(
		ctx,
		snapshotName,
		metav1.GetOptions{},
	)

	if err == nil {
		log.Printf("#%s %s already exists, skipping...", volumeSnapshot.Namespace, volumeSnapshot.Name)
		return nil
	}

	snapshot.Annotations["date"] = currentTime.Format("2006-01-02")
	create, err := client.Create(ctx, snapshot, metav1.CreateOptions{})

	if err != nil {
		log.Printf("#%s %s - cannot create snapshot: %s", snapshot.Namespace, snapshot.Name, err.Error())
		return err
	}

	log.Printf("#%s %s - snapshot created ssuccessfully", create.Namespace, create.Name)
	return nil
}

func (d SnapshotDriver) deleteSnapshot(snapshot snapshotterapiv1.VolumeSnapshot) error {
	ctx := context.Background()
	client := d.snapshotV1Client.VolumeSnapshots(snapshot.Namespace)
	err := client.Delete(ctx, snapshot.Name, metav1.DeleteOptions{})
	if err != nil {
		log.Printf("Cannot delete invalid snapshot: %s #%s (error: %s)", snapshot.Name, snapshot.Namespace, err.Error())
		return err
	}
	return nil
}

func (d SnapshotDriver) cleanupSnapshots() {
	ctx := context.Background()

	list, err := d.snapshotV1Client.VolumeSnapshots("").List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Cannot fetch snapshots: %s", err.Error())
		return
	}

	snapshots := map[string]map[string][]snapshotterapiv1.VolumeSnapshot{}

	for _, snapshot := range list.Items {
		if snapshot.Status == nil {
			continue
		}

		if *snapshot.Status.ReadyToUse {
			if _, ok := snapshots[snapshot.Namespace]; !ok {
				snapshots[snapshot.Namespace] = make(map[string][]snapshotterapiv1.VolumeSnapshot)
			}

			snapshots[snapshot.Namespace][snapshot.Annotations["pvc"]] = append(
				snapshots[snapshot.Namespace][snapshot.Annotations["pvc"]],
				snapshot,
			)
		}

		if !*snapshot.Status.ReadyToUse && snapshot.Status.Error != nil && snapshot.Status.Error.Time.Add(time.Minute*30).Before(time.Now()) {
			log.Printf("Snapshot is invalid, deleting...%s (error: %s)", snapshot.Name, *snapshot.Status.Error.Message)
			d.deleteSnapshot(snapshot)
		}
	}

	for namespace, pvcs := range snapshots {
		for pvc, snaps := range pvcs {
			if len(snaps) > 7 {
				sort.Slice(snaps, func(i, j int) bool {
					return snaps[i].Name > snaps[j].Name
				})

				log.Printf("#%s %s - remove old snapshots (%d valid snapshots)", namespace, pvc, len(snaps))
				for _, snap := range snaps[7:] {
					log.Printf("#%s %s removing", namespace, snap.Name)
					d.deleteSnapshot(snap)
				}
			} else {
				log.Printf("#%s %s has %d valid snapshots", namespace, pvc, len(snaps))
			}
		}
	}

}

func main() {
	var kubeconfig *string
	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()

	var config *rest.Config
	var err error
	if *kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		log.Panicf("Cannot initialize config: %s", err.Error())
	}

	driver := NewSnapshotDriver(config)

	log.Printf("Looking for snapshot rules...")
	rules, err := driver.listSnapshotRules()
	if err != nil {
		panic(err)
	}

	if len(rules) == 0 {
		log.Printf("No rules found")
	}

	for _, rule := range rules {
		log.Printf("Processing rule: %s #%s", rule.Name, rule.Namespace)
		pvcForRule, err := driver.getPvcForRule(rule)
		if err != nil {
			log.Printf("Cannot get pvc for rule: %s", err.Error())
			continue
		}
		for _, pvc := range pvcForRule {
			driver.createSnapshot(rule, pvc)
		}
	}

	driver.cleanupSnapshots()
}
