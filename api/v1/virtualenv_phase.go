package v1

type VirtualEnvPhase string

const (
	// VirtualEnvPhasePending is the first state a virtual env is assigned by
	// API server after being created.
	VirtualEnvPhasePending = VirtualEnvPhase("Pending")

	// VirtualEnvPhaseProvisioning is the state when the VirtualEnv has a infrastructure
	// virtual env infra object that can start provisioning the virtual machine.
	VirtualEnvPhaseProvisioning = VirtualEnvPhase("Provisioning")

	// VirtualEnvPhaseProvisioned is the state when its virtual env infra is created and configured
	// and the infrastructure object is ready (if defined).
	VirtualEnvPhaseProvisioned = VirtualEnvPhase("Provisioned")

	VirtualEnvPhaseRunning = VirtualEnvPhase("Running")

	// VirtualEnvPhaseStopped is the state when its virtual env infra is created and configured
	// and the infrastructure object is ready (if defined).
	VirtualEnvPhaseStopped = VirtualEnvPhase("Stopped")
	// VirtualEnvPhaseDeleting is the Virtual Env state when a delete
	// request has been sent to the API Server,
	// but its infrastructure has not yet been fully deleted.
	VirtualEnvPhaseDeleting = VirtualEnvPhase("Deleting")

	// VirtualEnvPhaseFailed is the Virtual Env state when the system
	// might require user intervention.
	VirtualEnvPhaseFailed = VirtualEnvPhase("Failed")

	// VirtualEnvPhaseUnknown is returned if the Virtual Env state cannot be determined.
	VirtualEnvPhaseUnknown = VirtualEnvPhase("Unknown")
)
