package sumup

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	clientpkg "github.com/sumup/sumup-go/client"
)

func eventPayload(eventType, objectURL string) []byte {
	return fmt.Appendf(nil, `{"id":"evt_123","type":%q,"created_at":"2026-04-11T10:00:00Z","object":{"id":"obj_123","type":"resource","url":%q}}`, eventType, objectURL)
}

func TestClient_ParseEventNotification(t *testing.T) {
	t.Parallel()

	c := NewClient()

	for eventType, typ := range map[string]EventNotification{
		EventTypeMemberCreated: &MemberCreatedEvent{},
		EventTypeMemberUpdated: &MemberUpdatedEvent{},
		EventTypeMemberDeleted: &MemberDeletedEvent{},
		EventTypeReaderCreated: &ReaderCreatedEvent{},
		EventTypeReaderDeleted: &ReaderDeletedEvent{},
		"future.event":         &UnknownEvent{},
	} {
		t.Run(eventType, func(t *testing.T) {
			t.Parallel()

			body := eventPayload(eventType, "https://api.sumup.com/object")
			signature := eventSignature(testEventSecret, time.Now(), body)
			event, err := c.ParseEventNotification(testEventSecret, body, signature)
			if err != nil {
				t.Fatal(err)
			}
			if reflect.TypeOf(event) != reflect.TypeOf(typ) {
				t.Fatalf("got %T, want %T", event, typ)
			}
			if event.EventID() != "evt_123" || event.EventType() != eventType {
				t.Fatalf("lost metadata: %v", event)
			}
		})
	}

	t.Run("verification precedes parsing", func(t *testing.T) {
		t.Parallel()

		if _, err := c.ParseEventNotification(testEventSecret, []byte("{"), "t=123,v1=bad"); !errors.Is(err, ErrEventSignatureExpired) {
			t.Fatalf("verification did not precede parsing: %v", err)
		}
	})
}

func TestClient_ParseEventNotificationWithoutVerification(t *testing.T) {
	t.Parallel()

	c := NewClient()
	for name, payload := range map[string]string{
		"malformed JSON":   "{",
		"null payload":     "null",
		"array payload":    "[]",
		"wrong field type": `{"type":42}`,
		"invalid date":     `{"created_at":"invalid"}`,
		"trailing JSON":    `{} {}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := c.ParseEventNotificationWithoutVerification([]byte(payload)); !errors.Is(err, ErrEventPayloadInvalid) {
				t.Fatalf("error = %v", err)
			}
		})
	}

	t.Run("empty fields use zero values", func(t *testing.T) {
		t.Parallel()

		event, err := c.ParseEventNotificationWithoutVerification([]byte(`{}`))
		if err != nil {
			t.Fatal(err)
		}

		unknown, ok := event.(*UnknownEvent)
		if !ok || unknown.EventID() != "" || unknown.EventType() != "" || !unknown.CreatedAt.IsZero() {
			t.Fatalf("unexpected empty event: %#v", event)
		}
	})
}

func TestTypedEvent_FetchObject(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{EventTypeMemberUpdated, "future.event"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.EscapedPath() != "/base/object/a%2Fb" || r.URL.RawQuery != "expand=payments&x=1&x=2" {
					t.Errorf("request URL changed: %s %s", r.Method, r.URL)
				}
				if r.Header.Get("Authorization") != "Bearer test-key" || r.Header.Get("User-Agent") == "" || r.Header.Get("X-Sumup-Lang") != "go" {
					t.Error("missing client headers")
				}
				_, _ = w.Write([]byte(`{"id":"obj_123"}`))
			}))
			t.Cleanup(server.Close)
			httpClient := &http.Client{Transport: eventRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.User != nil || req.URL.Fragment != "" {
					t.Error("credentials or fragment reached transport")
				}
				return http.DefaultTransport.RoundTrip(req)
			})}
			c := NewClient(clientpkg.WithBaseURL(server.URL+"/base"), clientpkg.WithAPIKey("test-key"), clientpkg.WithClient(httpClient))
			event, err := c.ParseEventNotificationWithoutVerification(eventPayload(kind, "https://user:password@api.sumup.com/object/a%2Fb?expand=payments&x=1&x=2#fragment"))
			if err != nil {
				t.Fatal(err)
			}
			switch e := event.(type) {
			case *MemberUpdatedEvent:
				object, err := e.FetchObject(t.Context())
				if err != nil {
					t.Fatal(err)
				}
				if object.ID != "obj_123" {
					t.Fatalf("object = %+v", object)
				}
			case *UnknownEvent:
				object, err := e.FetchObject(t.Context())
				if err != nil {
					t.Fatal(err)
				}
				if string(*object) != `{"id":"obj_123"}` {
					t.Fatalf("object = %s", *object)
				}
			default:
				t.Fatalf("unexpected event: %T", event)
			}
		})
	}
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"unknown error body", 404, `{"id":"not-a-success"}`, "HTTP 404"},
		{"plain text error", 503, `unavailable`, "HTTP 503"},
		{"rate limited", 429, `{"message":"retry later"}`, "HTTP 429"},
		{"malformed problem", 400, `{"type":`, "HTTP 400"},
		{"invalid JSON", 200, `{`, "decode event object"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)
			c := NewClient(clientpkg.WithBaseURL(server.URL))
			e := TypedEvent[Member]{client: c, Object: EventObject{URL: clientpkg.APIUrl + "/object"}}
			_, err := e.FetchObject(t.Context())
			if tc.status != http.StatusOK {
				var apiError *APIError
				if !errors.As(err, &apiError) || apiError.StatusCode != tc.status || string(apiError.Body) != tc.body || apiError.Headers.Get("Retry-After") != "60" {
					t.Fatalf("lost API error context: %v", err)
				}
				var problem *Problem
				if errors.As(err, &problem) {
					t.Fatalf("unexpected Problem for %s", tc.body)
				}
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v", err)
			}
		})
	}
	t.Run("preserves API problem details", func(t *testing.T) {
		t.Parallel()
		for _, tc := range []struct {
			name   string
			status int
			body   string
		}{
			{"not found", http.StatusNotFound, `{"type":"about:blank","title":"Not Found","detail":"Member was deleted","instance":"/requests/123","status":404}`},
			{"missing status", http.StatusNotFound, `{"type":"about:blank","title":"Not Found","detail":"Member was deleted","instance":"/requests/123"}`},
		} {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/problem+json")
					w.WriteHeader(tc.status)
					_, _ = w.Write([]byte(tc.body))
				}))
				t.Cleanup(server.Close)
				c := NewClient(clientpkg.WithBaseURL(server.URL))
				event := TypedEvent[Member]{client: c, Object: EventObject{URL: clientpkg.APIUrl + "/object"}}
				object, err := event.FetchObject(t.Context())
				var apiError *APIError
				if !errors.As(err, &apiError) || apiError.StatusCode != tc.status || string(apiError.Body) != tc.body || apiError.Headers.Get("Content-Type") != "application/problem+json" {
					t.Fatalf("lost API error context: %v", err)
				}
				var problem *Problem
				if object != nil || !errors.As(err, &problem) {
					t.Fatalf("object = %v, error = %v; expected API problem", object, err)
				}
				if problem.Status == nil || *problem.Status != tc.status || problem.Type != "about:blank" ||
					problem.Title == nil || *problem.Title != "Not Found" ||
					problem.Detail == nil || *problem.Detail != "Member was deleted" ||
					problem.Instance == nil || *problem.Instance != "/requests/123" {
					t.Fatalf("lost problem details: %v", problem)
				}
			})
		}
	})
	t.Run("missing client", func(t *testing.T) {
		t.Parallel()
		if _, err := (TypedEvent[Member]{}).FetchObject(t.Context()); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("transport cause", func(t *testing.T) {
		t.Parallel()
		cause := errors.New("transport failed")
		c := NewClient(clientpkg.WithClient(&http.Client{Transport: eventRoundTripFunc(func(*http.Request) (*http.Response, error) { return nil, cause })}))
		e := TypedEvent[Member]{client: c, Object: EventObject{URL: "https://api.sumup.com/object"}}
		if _, err := e.FetchObject(t.Context()); !errors.Is(err, cause) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("cancellation", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		e := TypedEvent[Member]{client: NewClient(), Object: EventObject{URL: "https://api.sumup.com/object"}}
		if _, err := e.FetchObject(ctx); !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("rejects unsafe URLs before sending credentials", func(t *testing.T) {
		t.Parallel()
		c := NewClient(clientpkg.WithAPIKey("must-not-leak"), clientpkg.WithClient(&http.Client{Transport: eventRoundTripFunc(func(*http.Request) (*http.Response, error) {
			t.Error("unsafe request reached transport")
			return nil, errors.New("unexpected request")
		})}))
		for _, tc := range []struct{ name, url string }{
			{"empty", ""},
			{"malformed", "://bad"},
			{"relative", "/object"},
			{"scheme relative", "//api.sumup.com/object"},
			{"lookalike host", "https://api.sumup.com.evil/object"},
			{"insecure scheme", "http://api.sumup.com/object"},
			{"different port", "https://api.sumup.com:444/object"},
			{"non HTTP", "file:///etc/passwd"},
			{"different host", "https://other.example/object"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				e := TypedEvent[Member]{client: c, Object: EventObject{URL: tc.url}}
				if _, err := e.FetchObject(t.Context()); err == nil {
					t.Fatal("accepted unsafe URL")
				}
			})
		}
	})
	for _, customPolicy := range []bool{false, true} {
		t.Run(fmt.Sprintf("does not follow redirects/custom policy=%t", customPolicy), func(t *testing.T) {
			t.Parallel()
			requests := 0
			const target = "http://api.sumup.com/redirected"
			httpClient := &http.Client{Transport: eventRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				requests++
				if r.URL.String() == target {
					return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"id":"obj_123"}`)), Request: r}, nil
				}
				return &http.Response{StatusCode: http.StatusFound, Header: http.Header{"Location": {target}}, Body: io.NopCloser(strings.NewReader("redirect")), Request: r}, nil
			})}
			policyCalled := false
			policyError := errors.New("custom redirect policy")
			if customPolicy {
				httpClient.CheckRedirect = func(*http.Request, []*http.Request) error {
					policyCalled = true
					return policyError
				}
			}
			c := NewClient(clientpkg.WithClient(httpClient), clientpkg.WithAPIKey("test-key"))
			e := TypedEvent[Member]{client: c, Object: EventObject{URL: clientpkg.APIUrl + "/object"}}
			_, err := e.FetchObject(t.Context())
			if requests != 1 || policyCalled {
				t.Fatalf("redirect escaped fetch: requests=%d custom policy called=%t", requests, policyCalled)
			}
			var apiError *APIError
			if !errors.As(err, &apiError) || apiError.StatusCode != http.StatusFound || apiError.Headers.Get("Location") != target {
				t.Fatalf("lost redirect response: %v", err)
			}

			// Event fetching must not change the shared client's redirect policy.
			requests = 0
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, clientpkg.APIUrl+"/object", nil)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := httpClient.Do(req)
			if customPolicy {
				if !policyCalled || !errors.Is(err, policyError) {
					t.Fatalf("custom redirect policy changed: %v", err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = resp.Body.Close() }()
				if requests != 2 || resp.StatusCode != http.StatusOK {
					t.Fatalf("default redirect policy changed: requests=%d status=%d", requests, resp.StatusCode)
				}
			}
		})
	}
}

type eventRoundTripFunc func(*http.Request) (*http.Response, error)

func (f eventRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
