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

package controller

import (
	"context"

	virtualenvv1 "github.com/clody-io/nebula/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

// SetupWithManager sets up the controller with the Manager.
func (r *VirtualEnvReconciler) reconcilePhase(_ context.Context, venv *virtualenvv1.VirtualEnv) {
	preReconcilePhase := venv.Status.GetTypedPhase()

	if venv.Status.Phase == "" {
		venv.Status.SetTypedPhase(virtualenvv1.VirtualEnvPhasePending)
	}

	if venv.Spec.VirtualEnvInfraRef != nil {
		venv.Status.SetTypedPhase(virtualenvv1.VirtualEnvPhaseProvisioning)
	}

	if venv.Status.InfrastructureReady {
		venv.Status.SetTypedPhase(virtualenvv1.VirtualEnvPhaseProvisioned)
	}

	if venv.Status.FailureMessage != nil {
		venv.Status.SetTypedPhase(virtualenvv1.VirtualEnvPhaseFailed)
	}

	if !venv.DeletionTimestamp.IsZero() {
		venv.Status.SetTypedPhase(virtualenvv1.VirtualEnvPhaseDeleting)
	}

	// Only record the event if the status has changed
	if preReconcilePhase != venv.Status.GetTypedPhase() {
		// Failed clusters should get a Warning event
		if venv.Status.GetTypedPhase() == virtualenvv1.VirtualEnvPhaseFailed {
			r.recorder.Eventf(venv, corev1.EventTypeWarning, string(venv.Status.GetTypedPhase()), "Virtual ENV %s is %s: %s", venv.Name, string(venv.Status.GetTypedPhase()), ptr.Deref(venv.Status.FailureMessage, "unknown"))
		} else {
			r.recorder.Eventf(venv, corev1.EventTypeNormal, string(venv.Status.GetTypedPhase()), "Virtual ENV %s is %s", venv.Name, string(venv.Status.GetTypedPhase()))
		}
	}

}
