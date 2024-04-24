/*
Copyright 2024.

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

// OpenstackRoomSpec defines the desired state of OpenstackRoom
type OpenstackRoomSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of OpenstackRoom. Edit openstackroom_types.go to remove/update
	//Foo string `json:"foo,omitempty"`
	Region string `json:"region,omitempty"`
	Image  string `json:"image,omitempty"`
	Flavor string `json:"flavor,omitempty"`
}

// OpenstackRoomStatus defines the observed state of OpenstackRoom
type OpenstackRoomStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase       string      `json:"phase,omitempty"`
	Message     string      `json:"message,omitempty"`
	LastUpdated metav1.Time `json:"lastUpdated,omitempty"`
	//Conditions []Condition `json:"conditions,omitempty"`
}

type Condition struct {
	Type               string      `json:"type"`
	Status             string      `json:"status"`
	Reason             string      `json:"reason,omitempty"`
	Message            string      `json:"message,omitempty"`
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// OpenstackRoom is the Schema for the openstackrooms API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Region",type=string,JSONPath=`.spec.region`
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
// +kubebuilder:printcolumn:name="Flavor",type=string,JSONPath=`.spec.flavor`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type OpenstackRoom struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenstackRoomSpec   `json:"spec,omitempty"`
	Status OpenstackRoomStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OpenstackRoomList contains a list of OpenstackRoom
type OpenstackRoomList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenstackRoom `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OpenstackRoom{}, &OpenstackRoomList{})
}
