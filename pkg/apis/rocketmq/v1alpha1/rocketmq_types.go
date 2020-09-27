/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RocketmqSpec defines the desired state of Rocketmq
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type RocketmqSpec struct {
	// Broker defines broker info
	Broker BrokerSpec `json:"broker,omitempty"`
	// NameService defines name Service info
	NameService NameServiceSpec `json:"nameService,omitempty"`
}

// RocketmqStatus defines the observed state of Rocketmq
// +k8s:openapi-gen=true
type RocketmqStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Broker string `json:"broker"`
	NameService string `json:"nameService"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Rocketmq is the Schema for the Rocketmq API
// +k8s:openapi-gen=true
type Rocketmq struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RocketmqSpec `json:"spec,omitempty"`
	Status RocketmqStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RocketmqList contains a list of Rocketmq
type RocketmqList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items []Rocketmq `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Rocketmq{}, &RocketmqList{})
}
