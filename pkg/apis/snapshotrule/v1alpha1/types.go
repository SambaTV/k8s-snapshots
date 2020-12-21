// +kubebuilder:object:generate=true
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	GroupName string = "k8ssnapshots.io"
	Kind      string = "SnapshotRule"
	Version   string = "v1alpha1"
	Plural    string = "snapshotrules"
	Singular  string = "snapshotrule"
	Name             = Plural + "." + GroupName
)

type Selector struct {
	MatchLabels map[string]string `json:"matchLabels"`
}

type SnapshotRuleSpec struct {
	Selector Selector `json:"selector"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TestResource describes a TestResource custom resource.
type SnapshotRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec SnapshotRuleSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TestResourceList is a list of TestResource resources.
type SnapshotRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []SnapshotRule `json:"items"`
}
