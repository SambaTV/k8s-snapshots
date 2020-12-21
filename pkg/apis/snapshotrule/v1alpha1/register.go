package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// SchemeBuilder initializes a scheme builder
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	// AddToScheme is a global function that registers this API group & version to a scheme
	AddToScheme = SchemeBuilder.AddToScheme
)

// SchemeGroupVersion is group version used to register these objects.
var SchemeGroupVersion = schema.GroupVersion{
	Group:   GroupName,
	Version: Version,
}

//var (
//	SchemeBuilder      runtime.SchemeBuilder
//	localSchemeBuilder = &SchemeBuilder
//	AddToScheme        = localSchemeBuilder.AddToScheme
//)

func init() {
	SchemeBuilder.Register(addKnownTypes)
}

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(
		SchemeGroupVersion,
		&SnapshotRule{},
		&SnapshotRuleList{},
	)
	scheme.AddKnownTypes(
		SchemeGroupVersion,
		&metav1.Status{},
	)
	metav1.AddToGroupVersion(
		scheme,
		SchemeGroupVersion,
	)
	return nil
}
