k8s-snapshots
====

K8s-snapshots is a small snapshot manager for k8s clusters written in Golang.

Requirements
====
Application requires to have CRD from external-snapshotter repository and Volume driver with Snapshot support.

Basics
====
Application lists all own resources to define rules for snapshotting.
Now only daily backups are supported with 7-day retention.

Example
====
Currently, the application supports only own CRD for snapshots' management, example below:

```yaml
apiVersion: "k8ssnapshots.io/v1alpha1"
kind: SnapshotRule
metadata:
  name: zookeeper
  namespace: kafka
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: zookeeper
```
