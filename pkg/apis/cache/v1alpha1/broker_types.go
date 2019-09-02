package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BrokerSpec defines the desired state of Broker
// +k8s:openapi-gen=true
type BrokerSpec struct {
	Size int32 `json:"size"`
	// NameServers defines the name service list e.g. 192.168.1.1:9876;192.168.1.2:9876
	NameServers string `json:"nameServers,omitempty"`
	// ReplicationMode is SYNC or ASYNC
	ReplicationMode string `json:"replicationMode,omitempty"`
	// SlavePerGroup is the number of slave brokers in each broker group
	SlavePerGroup int `json:"slavePerGroup"`
	// BaseImage is the broker container image to use for the Pods.
	BrokerImage string `json:"brokerImage"`
	// ImagePullPolicy defines how the image is pulled.
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`
}

// BrokerStatus defines the observed state of Broker
// +k8s:openapi-gen=true
type BrokerStatus struct {
	Nodes []string `json:"nodes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Broker is the Schema for the brokers API
// +k8s:openapi-gen=true
type Broker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BrokerSpec   `json:"spec,omitempty"`
	Status BrokerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BrokerList contains a list of Broker
type BrokerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Broker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Broker{}, &BrokerList{})
}
