package v1

type OpenstackVMError struct {
	// Human-friendly description of the error.
	Message string `json:"message,omitempty"`
}
