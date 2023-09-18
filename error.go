package sumup

import "fmt"

// APIError is custom error type for SumUp API that combinas *all* of our
// current error types into one underlying struct.
type APIError struct {
	// Platform code for the error.
	ErrorCode *string `json:"error_code,omitempty"`
	// Short description of the error.
	Message *string `json:"message,omitempty"`
	// Parameter name (with relative location) to which the error applies. Parameters from embedded resources are displayed using dot notation. For example, `card.name` refers to the `name` parameter embedded in the `card` object.
	Param *string `json:"param,omitempty"`
	// Short description of the error.
	ErrorMessage *string `json:"error_message,omitempty"`
	// HTTP status code for the error.
	StatusCode *string `json:"status_code,omitempty"`
	// Details of the error.
	Details           *string   `json:"details,omitempty"`
	FailedConstraints *[]string `json:"failed_constraints,omitempty"`
	// The status code.
	Status *float64 `json:"status,omitempty"`
	// Short title of the error.
	Title *string `json:"title,omitempty"`
}

// Error returns the error message.
func (e *APIError) Error() string {
	var code string
	if e.ErrorCode != nil {
		code = *e.ErrorCode
	}
	if e.StatusCode != nil && code == "" {
		code = *e.StatusCode
	}
	if e.Title != nil && code == "" {
		code = *e.Title
	}

	var message string
	if e.Message != nil {
		message = *e.Message
	}
	if e.ErrorMessage != nil && message == "" {
		message = *e.ErrorMessage
	}
	if e.Details != nil && message == "" {
		message = *e.Details
	}

	return fmt.Sprintf("%s: %s", code, message)
}
