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
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VirtualEnvSpec defines the desired state of VirtualEnv
type VirtualEnvSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of VirtualEnv. Edit virtualenv_types.go to remove/update
	VirtualEnvInfraRef *corev1.ObjectReference `json:"virtualEnvInfraRef"`

	// label ? annotation ?
	// annotation으로?
	// virtual-env.clody.io/user-id: '833287368871'
	// virtual-env.clody.io/lecture-id: '833287368871'

}

// VirtualEnvStatus defines the observed state of VirtualEnv
type VirtualEnvStatus struct {

	// FailureMessage indicates that there is a fatal problem reconciling the
	// state, and will be set to a descriptive error message.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`

	// Phase represents the current phase of cluster actuation.
	// E.g. Pending, Running, Terminating, Failed etc.
	// +optional
	Phase string `json:"phase,omitempty"`

	// InfrastructureReady is the state of the infrastructure provider.
	// +optional
	InfrastructureReady  bool                   `json:"infrastructureReady"` /// infra가 만들어져 있을때 ( running/stopped )
	VirtualMachineStatus []VirtualMachineStatus `json:"virtualMachineStatus,omitempty"`
}

// ANCHOR_END: ClusterStatus

// SetTypedPhase sets the Phase field to the string representation of ClusterPhase.
func (c *VirtualEnvStatus) SetTypedPhase(p VirtualEnvPhase) {
	c.Phase = string(p)
}

// GetTypedPhase attempts to parse the Phase field and return
// the typed ClusterPhase representation as described in `machine_phase_types.go`.
func (c *VirtualEnvStatus) GetTypedPhase() VirtualEnvPhase {
	switch phase := VirtualEnvPhase(c.Phase); phase {
	case
		VirtualEnvPhasePending,      /// VM 생성 요청만 한 상태
		VirtualEnvPhaseProvisioning, /// VM 생성중
		VirtualEnvPhaseProvisioned,  /// VM 생성 생성은 완료 initial scripts 가 돌고 있는중
		VirtualEnvPhaseRunning,      /// 실제로 사용자가 사용할 수 있는 상태
		VirtualEnvPhaseStopped,      /// 여러가지 이유로 vm이 stop 된상태
		VirtualEnvPhaseDeleting,
		VirtualEnvPhaseFailed:
		return phase
	default:
		return VirtualEnvPhaseUnknown
	}
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// VirtualEnv is the Schema for the virtualenvs API
type VirtualEnv struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualEnvSpec   `json:"spec,omitempty"`
	Status VirtualEnvStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VirtualEnvList contains a list of VirtualEnv
type VirtualEnvList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualEnv `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualEnv{}, &VirtualEnvList{})
}
