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
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var virtualenvinfralog = logf.Log.WithName("virtualenvinfra-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *VirtualEnvInfra) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-virtual-env-clody-io-v1-virtualenvinfra,mutating=true,failurePolicy=fail,sideEffects=None,groups=virtual-env.clody.io,resources=virtualenvinfras,verbs=create;update,versions=v1,name=mvirtualenvinfra.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &VirtualEnvInfra{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *VirtualEnvInfra) Default() {
	virtualenvinfralog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-virtual-env-clody-io-v1-virtualenvinfra,mutating=false,failurePolicy=fail,sideEffects=None,groups=virtual-env.clody.io,resources=virtualenvinfras,verbs=create;update,versions=v1,name=vvirtualenvinfra.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &VirtualEnvInfra{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *VirtualEnvInfra) ValidateCreate() (admission.Warnings, error) {
	virtualenvinfralog.Info("validate create", "name", r.Name)

	// 필요한 어노테이션 키들
	requiredAnnotations := []string{
		"nebula.clody.kubernetes.io/user-id",
		"nebula.clody.kubernetes.io/course-id",
	}
	// 누락된 어노테이션을 저장할 리스트
	missingAnnotations := []string{}

	// 어노테이션이 존재하는지 확인
	for _, annotationKey := range requiredAnnotations {
		if _, exists := r.Annotations[annotationKey]; !exists {
			missingAnnotations = append(missingAnnotations, annotationKey)
		}
	}

	// 누락된 어노테이션이 있다면 에러 반환
	if len(missingAnnotations) > 0 {
		return nil, fmt.Errorf("missing required annotations: %v", missingAnnotations)
	}

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *VirtualEnvInfra) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	virtualenvinfralog.Info("validate update", "name", r.Name)

	_, isVEnvInfra := old.(*VirtualEnvInfra)

	if !isVEnvInfra {
		return nil, fmt.Errorf("given resourse is not a virtualEnvInfra resource")
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *VirtualEnvInfra) ValidateDelete() (admission.Warnings, error) {
	virtualenvinfralog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
