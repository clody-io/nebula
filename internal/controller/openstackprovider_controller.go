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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// OpenstackProviderReconciler reconciles a OpenstackProvider object
type OpenstackProviderReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=openstackvms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=openstackvms/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=openstackvms/finalizers,verbs=update
// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=openstackproviders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=openstackproviders/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=openstackproviders/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OpenstackProvider object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *OpenstackProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// TODO(user): your logic here

	logger.Info("Hello Vm is created")

	openstackVM := &virtualenvv1.OpenstackVM{}
	if err := r.Get(ctx, req.NamespacedName, openstackVM); err != nil {
		if apierrors.IsNotFound(err) {

			openstackVM.Name = req.Name
			openstackVM.Namespace = req.Namespace

			return r.updateVirtualEnvInfraStatus(ctx, openstackVM, "deleted")
		}
		return ctrl.Result{}, err
	}

	return r.updateVirtualEnvInfraStatus(ctx, openstackVM, "created")
}

func (r *OpenstackProviderReconciler) updateVirtualEnvInfraStatus(ctx context.Context, vm *virtualenvv1.OpenstackVM, event string) (ctrl.Result, error) {
	// OpenstackVM의 OwnerReference를 통해 VirtualEnvInfra 찾기
	virtualEnvInfra := &virtualenvv1.VirtualEnvInfra{}
	// 2. OwnerReference로 A 리소스를 찾기
	owner := metav1.GetControllerOf(vm)
	if owner == nil || owner.Kind != "VirtualEnvInfra" {
		// A 리소스가 아닌 경우 무시
		return reconcile.Result{}, nil
	}
	// 3. A 리소스 가져오기
	if err := r.Get(ctx, client.ObjectKey{Name: owner.Name, Namespace: vm.Namespace}, virtualEnvInfra); err != nil {
		return reconcile.Result{}, err
	}

	var vmStatus = []virtualenvv1.VirtualMachineStatus{}

	switch event {
	case "deleted":
		vmStatus = append(
			vmStatus, virtualenvv1.VirtualMachineStatus{
				Status: "Deleted",
			})
		virtualEnvInfra.Status.VirtualMachineStatus = vmStatus
		return reconcile.Result{}, r.Status().Update(ctx, virtualEnvInfra)

	case "created":
		vmStatus = append(
			vmStatus, virtualenvv1.VirtualMachineStatus{
				Status:        "created",
				ConnectionURL: "http://172.16.156.253:8187",
			})
		virtualEnvInfra.Status.VirtualMachineStatus = vmStatus

		return reconcile.Result{}, r.Status().Update(ctx, virtualEnvInfra)
	}

	return ctrl.Result{}, r.Update(ctx, virtualEnvInfra)
}

// SetupWithManager sets up the controller with the Manager.
func (r *OpenstackProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&virtualenvv1.OpenstackVM{}).
		Complete(r)
}
