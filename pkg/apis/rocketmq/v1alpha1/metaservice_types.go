package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MetaServiceSpec defines the desired state of MetaService
// +k8s:openapi-gen=true
type MetaServiceSpec struct {
	// Size is the number of the name service Pod
	Size int32 `json:"size"`
	//MetaServiceImage is the name service image
	MetaServiceImage string `json:"metaServiceImage"`
	// ImagePullPolicy defines how the image is pulled.
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`
}

// MetaServiceStatus defines the observed state of MetaService
// +k8s:openapi-gen=true
type MetaServiceStatus struct {
	// MetaServers is the name service ip list
	MetaServers []string `json:"metaServers"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MetaService is the Schema for the metaservices API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type MetaService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MetaServiceSpec   `json:"spec,omitempty"`
	Status MetaServiceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MetaServiceList contains a list of MetaService
type MetaServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MetaService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MetaService{}, &MetaServiceList{})
}
