package sumup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	clientpkg "github.com/sumup/sumup-go/client"
)

func eventPayload(eventType, objectURL string) []byte {
	return []byte(fmt.Sprintf(`{"id":"evt_123","type":%q,"created_at":"2026-04-11T10:00:00Z","object":{"id":"obj_123","type":"resource","url":%q}}`, eventType, objectURL))
}

func TestClient_ParseEventNotification(t *testing.T) {
	t.Parallel()
	c := NewClient()
	for _, tc := range []struct {
		eventType string
		want      EventNotification
	}{
		{EventTypeMemberCreated, &MemberCreatedEvent{}},
		{EventTypeMemberUpdated, &MemberUpdatedEvent{}},
		{EventTypeMemberDeleted, &MemberDeletedEvent{}},
		{EventTypeReaderCreated, &ReaderCreatedEvent{}},
		{EventTypeReaderDeleted, &ReaderDeletedEvent{}},
		{"future.event", &UnknownEvent{}},
	} {
		t.Run(tc.eventType, func(t *testing.T) {
			t.Parallel()
			body := eventPayload(tc.eventType, "https://api.sumup.com/object")
			headers := signedHeader(testEventSecret, time.Now(), body)
			event, err := c.ParseEventNotification(testEventSecret, body, headers.Get(EventSignatureHeader))
			if err != nil {
				t.Fatal(err)
			}
			if reflect.TypeOf(event) != reflect.TypeOf(tc.want) {
				t.Fatalf("got %T, want %T", event, tc.want)
			}
			if event.EventID() != "evt_123" || event.EventType() != tc.eventType {
				t.Fatalf("lost metadata: %v", event)
			}
			encoded, err := json.Marshal(event)
			if err != nil {
				t.Fatal(err)
			}
			var original, roundtrip any
			if err := json.Unmarshal(body, &original); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(encoded, &roundtrip); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(original, roundtrip) {
				t.Fatalf("roundtrip changed payload: %s", encoded)
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
	for _, tc := range []struct{ name, payload string }{
		{"malformed JSON", "{"},
		{"null payload", "null"},
		{"invalid date", strings.Replace(string(eventPayload(EventTypeMemberCreated, "https://api.sumup.com/object")), "2026-04-11T10:00:00Z", "invalid", 1)},
		{"trailing JSON", string(eventPayload(EventTypeMemberCreated, "https://api.sumup.com/object")) + ` {}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := c.ParseEventNotificationWithoutVerification([]byte(tc.payload)); !errors.Is(err, ErrEventPayloadInvalid) {
				t.Fatalf("error = %v", err)
			}
		})
	}
	payload := eventPayload("future.event", "https://api.sumup.com/object")
	t.Run("unknown event retains client", func(t *testing.T) {
		t.Parallel()
		event, err := c.ParseEventNotificationWithoutVerification(payload)
		if err != nil {
			t.Fatal(err)
		}
		unknown := event.(*UnknownEvent)
		if unknown.client != c {
			t.Fatal("client missing")
		}
	})
	for _, key := range []string{"id", "type", "created_at", "object"} {
		t.Run("missing "+key, func(t *testing.T) {
			t.Parallel()
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(payload, &fields); err != nil {
				t.Fatal(err)
			}
			delete(fields, key)
			body, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := c.ParseEventNotificationWithoutVerification(body); !errors.Is(err, ErrEventPayloadInvalid) {
				t.Errorf("missing %s: %v", key, err)
			}
		})
	}
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
			c := NewClient(clientpkg.WithBaseURL(server.URL+"/base"), clientpkg.WithAPIKey("test-key"))
			event, err := c.ParseEventNotificationWithoutVerification(eventPayload(kind, clientpkg.APIUrl+"/object/a%2Fb?expand=payments&x=1&x=2"))
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
		{"unknown error body", 404, `{"id":"not-a-success"}`, "unexpected response 404"},
		{"plain text error", 503, `unavailable`, "unexpected response 503"},
		{"malformed problem", 400, `{"type":`, "unexpected response 400"},
		{"invalid JSON", 200, `{`, "decode event object"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(server.Close)
			c := NewClient(clientpkg.WithBaseURL(server.URL))
			e := TypedEvent[Member]{client: c, Object: EventObject{URL: clientpkg.APIUrl + "/object"}}
			if _, err := e.FetchObject(t.Context()); err == nil || !strings.Contains(err.Error(), tc.want) {
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
	t.Run("normalizes URL credentials and fragments", func(t *testing.T) {
		t.Parallel()
		for _, tc := range []struct{ name, url string }{
			{"credentials", "https://user:password@api.sumup.com/object/a%2Fb?x=1&x=2"},
			{"fragment", "https://api.sumup.com/object/a%2Fb?x=1&x=2#fragment"},
			{"both", "https://user:password@api.sumup.com/object/a%2Fb?x=1&x=2#fragment"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.EscapedPath() != "/base/object/a%2Fb" || r.URL.RawQuery != "x=1&x=2" {
						t.Errorf("path or query changed: %s", r.URL)
					}
					if r.URL.User != nil || r.URL.Fragment != "" || r.Header.Get("Authorization") != "Bearer test-key" {
						t.Error("request did not use normalized URL and client authentication")
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
				event := TypedEvent[Member]{client: c, Object: EventObject{URL: tc.url}}
				object, err := event.FetchObject(t.Context())
				if err != nil {
					t.Fatal(err)
				}
				if object.ID != "obj_123" {
					t.Fatalf("object = %+v", object)
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
	t.Run("uses configured redirect policy", func(t *testing.T) {
		t.Parallel()
		target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Error("followed redirect") }))
		t.Cleanup(target.Close)
		source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL, http.StatusFound) }))
		t.Cleanup(source.Close)
		redirectChecked := false
		httpClient := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
			redirectChecked = true
			return http.ErrUseLastResponse
		}}
		c := NewClient(clientpkg.WithBaseURL(source.URL), clientpkg.WithClient(httpClient), clientpkg.WithAPIKey("must-not-leak"))
		e := TypedEvent[Member]{client: c, Object: EventObject{URL: clientpkg.APIUrl + "/object"}}
		if _, err := e.FetchObject(t.Context()); err == nil || !strings.Contains(err.Error(), "302") {
			t.Fatalf("redirect error = %v", err)
		}
		if !redirectChecked {
			t.Fatal("configured redirect policy was not called")
		}
	})
}

type eventRoundTripFunc func(*http.Request) (*http.Response, error)

func (f eventRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
