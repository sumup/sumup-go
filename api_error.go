package sumup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// APIError preserves the HTTP response for an unsuccessful API request.
// Use errors.As to inspect a documented body such as *Problem.
type APIError struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	Err        error
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("HTTP %d: %v", e.StatusCode, e.Err)
	}
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, http.StatusText(e.StatusCode))
}

// Unwrap returns the typed API body or the response-reading error.
func (e *APIError) Unwrap() error { return e.Err }

func readAPIError(response *http.Response, body error) error {
	raw, err := io.ReadAll(response.Body)
	result := &APIError{StatusCode: response.StatusCode, Headers: response.Header.Clone(), Body: raw}
	if err != nil {
		result.Err = fmt.Errorf("read error response: %w", err)
	} else if body != nil && json.Unmarshal(raw, body) == nil {
		result.Err = body
	}
	return result
}
