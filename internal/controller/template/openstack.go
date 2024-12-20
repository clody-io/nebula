package template

import (
	virtualenvv1 "github.com/clody-io/nebula/api/v1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"strings"
)

func filterNebulaAnnotations(annotations map[string]string) map[string]string {
	nebulaAnnotations := make(map[string]string)

	for key, value := range annotations {
		if strings.HasPrefix(key, "nebula.clody.kubernetes.io/") {
			nebulaAnnotations[key] = value
		}
	}

	return nebulaAnnotations
}

func GenerateOpenstackVM(logCtx *log.Entry, virtualEnvInfra *virtualenvv1.VirtualEnvInfra) (*virtualenvv1.OpenstackVM, error) {

	annoatations := virtualEnvInfra.GetAnnotations()
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

	openstackVM := &virtualenvv1.OpenstackVM{
		ObjectMeta: metav1.ObjectMeta{
			Name:            virtualEnvInfra.Name + "-openstack-vm",
			Namespace:       virtualEnvInfra.Namespace,
			OwnerReferences: ownerRef,
			Annotations:     filterNebulaAnnotations(annoatations),
		},

		TypeMeta: metav1.TypeMeta{
			Kind:       "OpenstackVM",
			APIVersion: "virtual-env.clody.io/v1",
		},
	}

	return openstackVM, nil
}
