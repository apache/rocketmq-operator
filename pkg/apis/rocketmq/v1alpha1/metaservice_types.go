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
