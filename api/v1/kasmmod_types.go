/*
Copyright 2023.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	Statesucess = "Successfully Deploy"
	Statefails  = "Failed to Deploy"
)

// KasmmodSpec defines the desired state of Kasmmod
type KasmmodSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Kasmmod. Edit kasmmod_types.go to remove/update
	Size           int32  `json:"size"`
	Port           int32  `json:"port,omitempty"`
	Image          string `json:"image,omitempty"`
	Serviceaccount string `json:"serviceaccount,omitempty"`
	Sessionid      string `json:"sessionid,omitempty"`
}

// KasmmodStatus defines the observed state of Kasmmod
type KasmmodStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Nodes []string `json:"nodes"`
	State string   `json:"state"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Kasmmod is the Schema for the kasmmods API
type Kasmmod struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              KasmmodSpec   `json:"spec,omitempty"`
	Status            KasmmodStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KasmmodList contains a list of Kasmmod
type KasmmodList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kasmmod `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kasmmod{}, &KasmmodList{})
}
