package controller

import (
	"context"
	"fmt"

	virtualenvv1 "github.com/clody-io/nebula/api/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func CreateOrUpdateOpenstackVM(ctx context.Context, logCtx *log.Entry, c client.Client, obj *virtualenvv1.OpenstackVM, f controllerutil.MutateFn) (controllerutil.OperationResult, error) {
	key := client.ObjectKeyFromObject(obj)
	if err := c.Get(ctx, key, obj); err != nil {
		if !errors.IsNotFound(err) {
			return controllerutil.OperationResultNone, err
		}
		if err := mutate(f, key, obj); err != nil {
			return controllerutil.OperationResultNone, err
		}
		if err := c.Create(ctx, obj); err != nil {
			return controllerutil.OperationResultNone, err
		}
		return controllerutil.OperationResultCreated, nil
	}

	normalizedLive := obj.DeepCopy()

	// Mutate the live object to match the desired state.
	if err := mutate(f, key, obj); err != nil {
		return controllerutil.OperationResultNone, err
	}

	equality := conversion.EqualitiesOrDie(
		func(a, b resource.Quantity) bool {
			// Ignore formatting, only care that numeric value stayed the same.
			// TODO: if we decide it's important, it should be safe to start comparing the format.
			//
			// Uninitialized quantities are equivalent to 0 quantities.
			return a.Cmp(b) == 0
		},
		func(a, b metav1.MicroTime) bool {
			return a.UTC() == b.UTC()
		},
		func(a, b metav1.Time) bool {
			return a.UTC() == b.UTC()
		},
		func(a, b labels.Selector) bool {
			return a.String() == b.String()
		},
		func(a, b fields.Selector) bool {
			return a.String() == b.String()
		},
	)

	if equality.DeepEqual(normalizedLive, obj) {
		return controllerutil.OperationResultNone, nil
	}

	patch := client.MergeFrom(normalizedLive)

	if err := c.Patch(ctx, obj, patch); err != nil {
		return controllerutil.OperationResultNone, err
	}
	return controllerutil.OperationResultUpdated, nil
}

// mutate wraps a MutateFn and applies validation to its result
func mutate(f controllerutil.MutateFn, key client.ObjectKey, obj client.Object) error {
	if err := f(); err != nil {
		return fmt.Errorf("error while wrapping using MutateFn: %w", err)
	}
	if newKey := client.ObjectKeyFromObject(obj); key != newKey {
		return fmt.Errorf("MutateFn cannot mutate object name and/or object namespace")
	}
	return nil
}
