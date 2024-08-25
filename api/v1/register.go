package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// API Group
	Group string = "clody.io"

	// VirtualEnvInfraKind constants
	VirtualEnvInfraKind string = "VirtualEnvInfra"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: "v1"}
)
