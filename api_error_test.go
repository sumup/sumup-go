package sumup

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestAPIError(t *testing.T) {
	t.Parallel()
	t.Run("typed body and response context", func(t *testing.T) {
		t.Parallel()
		raw := `{"type":"about:blank","detail":"Invalid amount"}`
		response := &http.Response{StatusCode: 400, Header: http.Header{"Retry-After": {"10"}}, Body: io.NopCloser(strings.NewReader(raw))}
		err := readAPIError(response, &Problem{})
		var problem *Problem
		if !errors.As(err, &problem) || problem.Detail == nil || *problem.Detail != "Invalid amount" {
			t.Fatalf("typed Problem missing: %v", err)
		}
		var apiError *APIError
		if !errors.As(err, &apiError) || apiError.StatusCode != 400 || string(apiError.Body) != raw || apiError.Headers.Get("Retry-After") != "10" {
			t.Fatalf("response context missing: %v", err)
		}
	})
	t.Run("malformed response", func(t *testing.T) {
		t.Parallel()
		response := &http.Response{StatusCode: 502, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("{invalid"))}
		err := readAPIError(response, &Problem{})
		var apiError *APIError
		if !errors.As(err, &apiError) || apiError.StatusCode != 502 || string(apiError.Body) != "{invalid" {
			t.Fatalf("response context missing: %v", err)
		}
		var problem *Problem
		if errors.As(err, &problem) {
			t.Fatal("invalid body became a Problem")
		}
	})
}
