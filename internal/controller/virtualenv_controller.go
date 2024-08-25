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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	virtualenvv1 "github.com/clody-io/nebula/api/v1"
)

// VirtualEnvReconciler reconciles a VirtualEnv object
type VirtualEnvReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=virtualenvs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=virtualenvs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=virtualenvs/finalizers,verbs=update

// +kubebuilder:rbac:groups=events.k8s.io,resources=events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VirtualEnv object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *VirtualEnvReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Hello Virtual ENV in reconciled")

	// Fetch the Cluster instance.
	venv := &virtualenvv1.VirtualEnv{}
	if err := r.Client.Get(ctx, req.NamespacedName, venv); err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	//// Always reconcile the Status.Phase field.
	r.reconcilePhase(ctx, venv)
	//
	//// Always attempt to Patch the Cluster object and status after each reconciliation.
	//// Patch ObservedGeneration only if the reconciliation completed successfully
	//patchOpts := []patch.Option{}
	//if reterr == nil {
	//	patchOpts = append(patchOpts, patch.WithStatusObservedGeneration{})
	//}
	//if err := patchCluster(ctx, patchHelper, cluster, patchOpts...); err != nil {
	//	reterr = kerrors.NewAggregate([]error{reterr, err})

	venvInfra := virtualenvv1.VirtualEnvInfra{}
	err := r.Get(ctx, types.NamespacedName{Namespace: venv.Spec.VirtualEnvInfraRef.Namespace, Name: venv.Spec.VirtualEnvInfraRef.Name}, &venvInfra)
	if err != nil {
		return ctrl.Result{}, err
	}

	if venvInfra.Status.VirtualMachineStatus != nil && len(venvInfra.Status.VirtualMachineStatus) > 0 {
		venv.Status.InfrastructureReady = true
		venv.Status.SetTypedPhase(virtualenvv1.VirtualEnvPhaseProvisioned)
		venv.Status.VirtualMachineStatus = venvInfra.Status.VirtualMachineStatus
		return ctrl.Result{}, r.Status().Update(ctx, venv)
	}

	// Handle deletion reconciliation loop.
	if !venv.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, venv)
	}

	return ctrl.Result{}, nil
}

func (r *VirtualEnvReconciler) reconcileDelete(ctx context.Context, venv *virtualenvv1.VirtualEnv) (reconcile.Result, error) {
	return ctrl.Result{}, nil
}

// controlPlaneMachineToCluster is a handler.ToRequestsFunc to be used to enqueue requests for reconciliation
// for Cluster to update its status.controlPlaneInitialized field.
func (r *VirtualEnvReconciler) venvInfraToVenv(ctx context.Context, o client.Object) []ctrl.Request {
	m, ok := o.(*virtualenvv1.VirtualEnvInfra)
	var requests []reconcile.Request

	if !ok {
		panic(fmt.Sprintf("Expected a VirtualEnvInfra but got a %T", o))
	}

	// 변경된 ResourceB를 참조하는 ResourceA를 인덱스를 사용해 조회
	venvList := &virtualenvv1.VirtualEnvList{}
	if err := r.List(context.TODO(), venvList, client.MatchingFields{".spec.objectRefName": m.ObjectMeta.GetName()}); err != nil {
		return requests
	}

	// 해당 ResourceA에 대한 reconcile 요청 생성
	for _, venv := range venvList.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: client.ObjectKey{
				Name:      venv.Name,
				Namespace: venv.Namespace,
			},
		})
	}

	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *VirtualEnvReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("virtualenv-controller")
	return ctrl.NewControllerManagedBy(mgr).
		For(&virtualenvv1.VirtualEnv{}).
		Watches(
			&virtualenvv1.VirtualEnvInfra{},
			handler.EnqueueRequestsFromMapFunc(r.venvInfraToVenv),
		).
		Complete(r)
}
