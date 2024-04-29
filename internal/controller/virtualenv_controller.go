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
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	virtualenvv1 "clody.io/nebula/api/v1"
)

const virtualEnvFinalizer = "virtual-env.clody.io/finalizer"

// Definitions to manage provider
const (
	gcp       = "GCP"
	aws       = "AWS"
	local     = "local"
	openstack = "Openstack"
)

// Definitions to manage status conditions
const (
	// typeAvailableVirtualEnv represents the status of the Deployment reconciliation
	typeAvailableVirtualEnv = "Available"
	// typeDegradedVirtualEnv represents the status used when the custom resource is deleted and the finalizer operations are yet to occur.
	//typeDegradedVirtualEnv = "Degraded"
)

// VirtualEnvReconciler reconciles a VirtualEnv object
type VirtualEnvReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VirtualEnv object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=virtualenvs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=virtualenvs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=virtualenvs/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
func (r *VirtualEnvReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 1. cr 이 없으므로 미진행
	virtualEnv := &virtualenvv1.VirtualEnv{}
	err := r.Get(ctx, req.NamespacedName, virtualEnv)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// cr 이 발견되지 않음. 생성되지 않은 상태이거나 삭제되어야 함.
			logger.Info("virtualenv resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get virtualenv")
		return ctrl.Result{}, err
	}

	// 2. condition 업데이트
	if virtualEnv.Status.Conditions == nil || len(virtualEnv.Status.Conditions) == 0 {
		meta.SetStatusCondition(&virtualEnv.Status.Conditions, metav1.Condition{
			Type:               typeAvailableVirtualEnv,
			Status:             metav1.ConditionUnknown,
			LastTransitionTime: metav1.Time{},
			Reason:             "Reconciling",
			Message:            "Starting reconciliation",
		})
		if err = r.Status().Update(ctx, virtualEnv); err != nil {
			logger.Error(err, "Failed to update virtualenv status")
			return ctrl.Result{}, err
		}
		// fetch after update
		if err = r.Get(ctx, req.NamespacedName, virtualEnv); err != nil {
			logger.Error(err, "Failed to fetch virtualenv status")
			return ctrl.Result{}, err
		}
	}

	//switch virtualEnv.Spec.Provider {
	//case "Openstack":
	//case "local":
	//
	//}

	// 3. TODO: 필요시 finalizer 추가
	//if !controllerutil.ContainsFinalizer(virtualEnv, virtualEnvFinalizer) {
	//	logger.Info("Adding Finalizer for VirtualEv")
	//	if ok := controllerutil.AddFinalizer(virtualEnv, virtualEnvFinalizer); !ok {
	//		logger.Error(err, "Failed to add finalizer into the custom resource")
	//	}
	//	if err = r.Update(ctx, virtualEnv); err != nil {
	//		logger.Error(err, "Failed to update custom resource to add finalizer")
	//		return ctrl.Result{}, err
	//	}
	//}

	// 4. 삭제 로직
	isVirtualEnvMarkedToBeDeleted := virtualEnv.GetDeletionTimestamp() != nil
	if isVirtualEnvMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(virtualEnv, virtualEnvFinalizer) {
			// TODO: 필요시 finalizer 내부 로직 구현
			logger.Info("Performing Finalizer Operations for VirtualEnv before delete CR")
			//meta.SetStatusCondition(&virtualEnv.Status.Conditions, metav1.Condition{
			//	Type:               typeDegradedVirtualEnv,
			//	Status:             metav1.ConditionUnknown,
			//	LastTransitionTime: metav1.Time{},
			//	Reason:             "Finalizing",
			//	Message:            fmt.Sprintf("Performing finalizer operations for the custom resource: %s", virtualEnv.Name),
			//})
			//if err := r.Status().Update(ctx, virtualEnv); err != nil {
			//	logger.Error(err, "Failed to update virtualenv status")
			//	return ctrl.Result{}, err
			//}
			//
		}
		return ctrl.Result{}, nil
	}

	// 5. 내부 resource 없으면 생성
	if provider := virtualEnv.Spec.Provider; provider == openstack {
		found := &virtualenvv1.OpenstackRoom{}
		err = r.Get(ctx, types.NamespacedName{
			Namespace: virtualEnv.Namespace,
			Name:      virtualEnv.Name,
		}, found)
		if err != nil && apierrors.IsNotFound(err) {
			// 없어서 새로 생성함
			openstackRoom, err := r.OpenstackRoomForVirtualEnv(virtualEnv)
			if err != nil {
				logger.Error(err, "Failed to define new openstackRoom resource for virtualenv")
				meta.SetStatusCondition(&virtualEnv.Status.Conditions, metav1.Condition{
					Type:   typeAvailableVirtualEnv,
					Status: metav1.ConditionFalse, Reason: "Reconciling",
					Message: fmt.Sprintf("Failed to create OpenstackRoom for the custom resource (%s): (%s)", virtualEnv.Name, err)})

				if err := r.Status().Update(ctx, virtualEnv); err != nil {
					logger.Error(err, "Failed to update Virtualenv status")
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, err
			}
			logger.Info("Creating new openstackRoom",
				"OpenstackRoom.Namespace", openstackRoom.Namespace,
				"OpenstackRoom.Name", openstackRoom.Name,
			)
			if err = r.Create(ctx, openstackRoom); err != nil {
				logger.Error(err, "Failed to create new OpenstackRoom",
					"OpenstackRoom.Namespace", openstackRoom.Namespace,
					"OpenstackRoom.Name", openstackRoom.Name,
				)
				return ctrl.Result{}, err
			}
			// 생성완료이고 다음 작업 위해서 requeue 함 1분단위
			// TODO: requeue time 정하기
			return ctrl.Result{RequeueAfter: time.Minute}, nil
		} else if err != nil {
			logger.Error(err, "Failed to get openstackRoom")
			return ctrl.Result{}, err
		}
	} else if provider == local {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

//// doFinalizerOperationsForVirtualEnv sets up the controller with the Manager.
//func (r *VirtualEnvReconciler) doFinalizerOperationsForVirtualEnv(cr *virtualenvv1.VirtualEnv) error {
//}
//

func (r *VirtualEnvReconciler) OpenstackRoomForVirtualEnv(virtualEnv *virtualenvv1.VirtualEnv) (*virtualenvv1.OpenstackRoom, error) {
	ls := labelsForVirtualEnv(virtualEnv.Name)
	openstackRoom := &virtualenvv1.OpenstackRoom{
		ObjectMeta: metav1.ObjectMeta{
			Name: virtualEnv.Name,
			//GenerateName:               "",
			Namespace:         virtualEnv.Namespace,
			CreationTimestamp: metav1.Time{},
			DeletionTimestamp: nil,
			//DeletionGracePeriodSeconds: nil,
			Labels: ls,
			//Annotations:                nil,
			Finalizers:    nil,
			ManagedFields: nil,
		},
		Spec: virtualenvv1.OpenstackRoomSpec{
			Region: "fake-region",
			Image:  "fake-image",
			Flavor: "m1.extra_tiny",
		},
		Status: virtualenvv1.OpenstackRoomStatus{
			Phase:       "Pending",
			Message:     "Pending because nothing is implemented",
			LastUpdated: metav1.Time{},
		},
	}

	if err := ctrl.SetControllerReference(virtualEnv, openstackRoom, r.Scheme); err != nil {
		return nil, err
	}
	return openstackRoom, nil
}

// labelsForVirtualEnv returns the labels for selecting the resources
// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
func labelsForVirtualEnv(name string) map[string]string {
	return map[string]string{"app.kubernetes.io/name": name,
		"app.kubernetes.io/managed-by": "VirtualEnvController",
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *VirtualEnvReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&virtualenvv1.VirtualEnv{}).
		Owns(&virtualenvv1.OpenstackRoom{}).
		Complete(r)
}
