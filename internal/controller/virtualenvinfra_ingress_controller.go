package controller

import (
	"context"
	"fmt"
	virtualenvv1 "github.com/clody-io/nebula/api/v1"
	logger "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

func (r *VirtualEnvInfraReconciler) createServiceAndIngress(ctx context.Context, virtualEnvInfra *virtualenvv1.VirtualEnvInfra, instance *virtualenvv1.OpenstackVM) error {
	namespacedName := types.NamespacedName{
		Namespace: virtualEnvInfra.Namespace,
		Name:      virtualEnvInfra.Name,
	}

	logCtx := logger.WithField("VirtualEnvInfraController", namespacedName)

	logCtx.Infof("Creating ingress,svc,endpointSlices")

	ownerRef := []metav1.OwnerReference{
		{
			APIVersion:         virtualenvv1.GroupVersion.String(),
			Kind:               "VirtualEnvInfra",
			Name:               virtualEnvInfra.Name,
			UID:                virtualEnvInfra.UID,
			Controller:         ptr.To(true),
			BlockOwnerDeletion: ptr.To(true),
		},
	}
	endpointSlices := &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      virtualEnvInfra.Name + "-svc",
			Namespace: virtualEnvInfra.Namespace,
			Labels: map[string]string{
				"kubernetes.io/service-name":       virtualEnvInfra.Name + "-svc",
				"nebula.clody.kubernetes.io/vm-id": instance.Status.InstanceStatus.InstanceID,
			},
			OwnerReferences: ownerRef, // Add OwnerReferences here
		},
		AddressType: discoveryv1.AddressTypeIPv4,
		Ports: []discoveryv1.EndpointPort{
			{
				Name:        ptr.To("http"),
				Protocol:    ptr.To(corev1.ProtocolTCP),
				Port:        ptr.To(int32(8187)), // Target Port
				AppProtocol: ptr.To("http"),
			},
		},
		Endpoints: []discoveryv1.Endpoint{
			{
				Addresses: []string{instance.Status.FloatingIP}, // IP address
				Conditions: discoveryv1.EndpointConditions{
					Ready: ptr.To(true),
				},
			},
		},
	}

	if err := r.Client.Create(ctx, endpointSlices); err != nil {
		if !errors.IsAlreadyExists(err) {
			logCtx.WithError(err).Errorf("failed to create resource: %s", endpointSlices.GetName())
			return err
		}
		logCtx.Infof("%s already exists", endpointSlices.GetName())
	}

	logCtx.Infof("EndpointSlice is created")

	// Create Service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            virtualEnvInfra.Name + "-svc",
			Namespace:       virtualEnvInfra.Namespace,
			OwnerReferences: ownerRef, // Add OwnerReferences here
			Labels: map[string]string{
				"nebula.clody.kubernetes.io/vm-id": instance.Status.InstanceStatus.InstanceID,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Port:     80,
					Protocol: corev1.ProtocolTCP,
					TargetPort: intstr.IntOrString{
						IntVal: 8187,
					},
				},
			},
			ClusterIP: "None", // Headless service
		},
	}

	if err := r.Create(ctx, service); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create Service: %w", err)
	}

	logCtx.Infof("Service is created")

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      virtualEnvInfra.Name + "-svc",
			Namespace: virtualEnvInfra.Namespace,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class":                "nginx",
				"nginx.ingress.kubernetes.io/rewrite-target": "/",
				"cert-manager.io/cluster-issuer":             "letsencrypt-prod-clody",
			},
			Labels: map[string]string{
				"nebula.clody.kubernetes.io/vm-id": instance.Status.InstanceStatus.InstanceID,
			},
			OwnerReferences: ownerRef, // Add OwnerReferences here
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: ptr.To("nginx"),
			Rules: []networkingv1.IngressRule{
				{
					Host: "dev.clody.io",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/" + fmt.Sprintf("%s", instance.Status.InstanceID),
									PathType: ptr.To(networkingv1.PathTypePrefix),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: virtualEnvInfra.Name + "-svc",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			TLS: []networkingv1.IngressTLS{
				{Hosts: []string{"dev.clody.io"},
					SecretName: virtualEnvInfra.Name + "-svc-tls-secret",
				},
			},
		},
	}

	if err := r.Create(ctx, ingress); err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create Ingress: %w", err)
	}

	logCtx.Infof("Ingress is created")

	return nil
}
