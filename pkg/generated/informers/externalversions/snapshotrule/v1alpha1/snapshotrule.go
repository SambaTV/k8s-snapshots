/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	snapshotrulev1alpha1 "github.com/sambatv/k8s-snapshots/pkg/apis/snapshotrule/v1alpha1"
	versioned "github.com/sambatv/k8s-snapshots/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/sambatv/k8s-snapshots/pkg/generated/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/sambatv/k8s-snapshots/pkg/generated/listers/snapshotrule/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// SnapshotRuleInformer provides access to a shared informer and lister for
// SnapshotRules.
type SnapshotRuleInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.SnapshotRuleLister
}

type snapshotRuleInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewSnapshotRuleInformer constructs a new informer for SnapshotRule type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewSnapshotRuleInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredSnapshotRuleInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredSnapshotRuleInformer constructs a new informer for SnapshotRule type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredSnapshotRuleInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.K8ssnapshotsV1alpha1().SnapshotRules(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.K8ssnapshotsV1alpha1().SnapshotRules(namespace).Watch(context.TODO(), options)
			},
		},
		&snapshotrulev1alpha1.SnapshotRule{},
		resyncPeriod,
		indexers,
	)
}

func (f *snapshotRuleInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredSnapshotRuleInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *snapshotRuleInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&snapshotrulev1alpha1.SnapshotRule{}, f.defaultInformer)
}

func (f *snapshotRuleInformer) Lister() v1alpha1.SnapshotRuleLister {
	return v1alpha1.NewSnapshotRuleLister(f.Informer().GetIndexer())
}
