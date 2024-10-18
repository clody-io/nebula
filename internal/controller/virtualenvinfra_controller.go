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

	virtualenvv1 "github.com/clody-io/nebula/api/v1"
	"github.com/clody-io/nebula/internal/controller/template"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	logger "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	AnnotationUserID    = "virtualenvinfra.clody.io/user-id"
	AnnotationLectureID = "virtualenvinfra.clody.io/lecture-id"
)

// VirtualEnvInfraReconciler reconciles a VirtualEnvInfra object
type VirtualEnvInfraReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	VirtualEnvInfra *virtualenvv1.VirtualEnvInfra
}

// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=virtualenvinfras,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=virtualenvinfras/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=virtualenvinfras/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VirtualEnvInfra object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *VirtualEnvInfraReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logCtx := logger.WithField("VirtualEnvInfra", req.NamespacedName)

	var virtualEnvInfra virtualenvv1.VirtualEnvInfra

	if err := r.Get(ctx, req.NamespacedName, &virtualEnvInfra); err != nil {
		if client.IgnoreNotFound(err) != nil {
			logCtx.WithError(err).Infof("unable to get VirtualEnvInfra: '%v' ", err)
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Do not attempt to further reconcile the ApplicationSet if it is being deleted.
	if virtualEnvInfra.ObjectMeta.DeletionTimestamp != nil {
		venvInfraName := virtualEnvInfra.ObjectMeta.Name
		logCtx.Debugf("DeletionTimestamp is set on %s", venvInfraName)
		if err := r.Update(ctx, &virtualEnvInfra); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	switch virtualEnvInfra.Spec.Provider {
	case "openstack":
		return r.createOrUpdateOpenstackVM(ctx, virtualEnvInfra)

	default:
		logCtx.Warn(fmt.Sprintf("unsupported provider %s", virtualEnvInfra.Spec))
	}

	return ctrl.Result{}, nil
}

func (r *VirtualEnvInfraReconciler) createOrUpdateOpenstackVM(ctx context.Context, virtualEnvInfra virtualenvv1.VirtualEnvInfra) (ctrl.Result, error) {

	namespacedName := types.NamespacedName{
		Namespace: virtualEnvInfra.Namespace,
		Name:      virtualEnvInfra.Name,
	}
	logCtx := logger.WithField("VirtualEnvInfra", namespacedName)

	desiredVM, err := template.GenerateOpenstackVM(logCtx, virtualEnvInfra)

	found := &virtualenvv1.OpenstackVM{
		ObjectMeta: metav1.ObjectMeta{
			Name:      desiredVM.Name,
			Namespace: desiredVM.Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "OpenstackVM",
			APIVersion: "virtual-env.clody.io/v1",
		},
	}

	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to build desired vm for virtual env infra: %w", err)
	}

	_, err = CreateOrUpdateOpenstackVM(ctx, logCtx, r.Client, desiredVM, func() error {
		// Copy only the Application/ObjectMeta fields that are significant, from the generatedApp
		found.Spec = desiredVM.Spec
		return controllerutil.SetControllerReference(&virtualEnvInfra, found, r.Scheme)
	})
	return ctrl.Result{}, err
}

// getStringFromAnnotation searches a given v1.virtualenvinfra for a specific annotationKey and either returns the annotation's value.
func (r *VirtualEnvInfraReconciler) getStringFromAnnotation(virtualEnvInfra virtualenvv1.VirtualEnvInfra, annotationKey string) string {
	klog.V(4).Infof("getStringFromAnnotation (%s/%s, %v)", virtualEnvInfra.Namespace, virtualEnvInfra.Name, annotationKey)
	if annotationValue, ok := virtualEnvInfra.Annotations[annotationKey]; ok {
		// if there is an annotation, return string about annotation val.
		// annotationValue can be empty, it is working as designed
		klog.V(4).Infof("Found Annotation: %v = %v", annotationKey, annotationValue)
		return annotationValue
	}
	//if there is no annotation, it returns an empty string ("").
	return ""
}

func (r *VirtualEnvInfraReconciler) getCurrentVm(ctx context.Context, virtualEnvInfra virtualenvv1.VirtualEnvInfra) (*virtualenvv1.OpenstackVM, error) {
	logCtx := logger.WithField("VirtualEnvInfra", types.NamespacedName{Name: virtualEnvInfra.Name, Namespace: virtualEnvInfra.Namespace})

	logCtx.Infof("Get currnent OpenstackVM")

	current := &virtualenvv1.OpenstackVM{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: virtualEnvInfra.Namespace,
		Name:      virtualEnvInfra.Name + "-openstack-vm",
	}, current)
	if client.IgnoreNotFound(err) != nil {
		return nil, fmt.Errorf("error retrieving openstackVM: %w", err)
	}
	return current, nil
}

func VirtualEnvInfraControllerIndexer(rawObj client.Object) []string {
	// grab the job object, extract the owner...
	app := rawObj.(*virtualenvv1.OpenstackVM)
	owner := metav1.GetControllerOf(app)
	if owner == nil {
		return nil
	}
	// ...make sure it's a Virtual Env Infra
	if owner.APIVersion != virtualenvv1.GroupVersion.Version || owner.Kind != "VirtualEnvInfra" {
		return nil
	}

	// ...and if so, return it
	return []string{owner.Name}
}

// SetupWithManager sets up the controller with the Manager.
func (r *VirtualEnvInfraReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(context.TODO(), &virtualenvv1.OpenstackVM{}, ".metadata.controller", VirtualEnvInfraControllerIndexer); err != nil {
		return fmt.Errorf("error setting up with manager: %w", err)
	}
	openstackVmOwnsHandler := getOpenstackVMOwnsHandlerPredicates()
	return ctrl.NewControllerManagedBy(mgr).
		For(&virtualenvv1.VirtualEnvInfra{}).
		Owns(&virtualenvv1.OpenstackVM{}, builder.WithPredicates(openstackVmOwnsHandler)).
		Complete(r)
}

func getOpenstackVMOwnsHandlerPredicates() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			// if we are the owner and there is a create event, we most likely created it and do not need to
			// re-reconcile
			if logger.IsLevelEnabled(logger.DebugLevel) {
				var vmName string
				vm, isVm := e.Object.(*virtualenvv1.OpenstackVM)
				if isVm {
					vmName = vm.Name
				}
				logger.WithField("VirtualMachine", vmName).Debugln("received create event from owning an OpenstackVM")
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			vmOld, isVm := e.ObjectOld.(*virtualenvv1.OpenstackVM)
			if !isVm {
				return false
			}
			logCtx := logger.WithField("VirtualMachine", vmOld.Name)
			logCtx.Debugln("received update event from owning an OpenstackVM")
			vmNew, isVm := e.ObjectNew.(*virtualenvv1.OpenstackVM)
			if !isVm {
				return false
			}
			requeue := shouldRequeueVirtualEnvInfraByOpenstackVM(vmOld, vmNew)
			logCtx.WithField("requeue", requeue).Debugf("requeue: %t caused by VirtualMachine %s\n", requeue, vmNew.Name)
			return requeue
		},
		GenericFunc: func(e event.GenericEvent) bool {
			if logger.IsLevelEnabled(logger.DebugLevel) {
				var vmName string
				vm, isVm := e.Object.(*virtualenvv1.OpenstackVM)
				if isVm {
					vmName = vm.Name
				}
				logger.WithField("OpenstackVM", vmName).Debugln("received generic event from owning an OpenstackVM")
			}
			return true
		},
	}
}

func shouldRequeueVirtualEnvInfraByOpenstackVM(vmOld *virtualenvv1.OpenstackVM, vmNew *virtualenvv1.OpenstackVM) bool {
	if vmOld == nil || vmNew == nil {
		return false
	}

	// the applicationset controller owns the application spec, labels, annotations, and finalizers on the applications
	// reflect.DeepEqual considers nil slices/maps not equal to empty slices/maps
	// https://pkg.go.dev/reflect#DeepEqual
	// ApplicationDestination has an unexported field so we can just use the == for comparison
	if !cmp.Equal(vmOld.Spec, vmNew.Spec, cmpopts.EquateEmpty()) ||
		!cmp.Equal(vmOld.ObjectMeta.GetAnnotations(), vmNew.ObjectMeta.GetAnnotations(), cmpopts.EquateEmpty()) ||
		!cmp.Equal(vmOld.ObjectMeta.GetLabels(), vmNew.ObjectMeta.GetLabels(), cmpopts.EquateEmpty()) ||
		!cmp.Equal(vmOld.ObjectMeta.GetFinalizers(), vmNew.ObjectMeta.GetFinalizers(), cmpopts.EquateEmpty()) {
		return true
	}

	return false
}
