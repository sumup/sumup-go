package sumup

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	clientpkg "github.com/sumup/sumup-go/client"
)

func TestWebhookHandler_parseVerified(t *testing.T) {
	t.Parallel()

	client := &Client{}
	handler := NewWebhookHandler(WithClient(client))
	createdAt := "2026-04-11T10:00:00Z"

	tests := []struct {
		name string
		body string
		want any
	}{
		{
			name: "checkout created",
			body: `{"id":"evt_123","type":"checkout.created","created_at":"` + createdAt + `","object":{"id":"chk_123","type":"checkout","url":"https://api.sumup.com/v0.1/checkouts/chk_123"}}`,
			want: &CheckoutCreatedWebhookEvent{},
		},
		{
			name: "checkout processed",
			body: `{"id":"evt_123","type":"checkout.processed","created_at":"` + createdAt + `","object":{"id":"chk_123","type":"checkout","url":"https://api.sumup.com/v0.1/checkouts/chk_123"}}`,
			want: &CheckoutProcessedWebhookEvent{},
		},
		{
			name: "checkout failed",
			body: `{"id":"evt_123","type":"checkout.failed","created_at":"` + createdAt + `","object":{"id":"chk_123","type":"checkout","url":"https://api.sumup.com/v0.1/checkouts/chk_123"}}`,
			want: &CheckoutFailedWebhookEvent{},
		},
		{
			name: "checkout terminated",
			body: `{"id":"evt_123","type":"checkout.terminated","created_at":"` + createdAt + `","object":{"id":"chk_123","type":"checkout","url":"https://api.sumup.com/v0.1/checkouts/chk_123"}}`,
			want: &CheckoutTerminatedWebhookEvent{},
		},
		{
			name: "member created",
			body: `{"id":"evt_123","type":"member.created","created_at":"` + createdAt + `","object":{"id":"mem_123","type":"member","url":"https://api.sumup.com/v0.1/merchants/M123/members/mem_123"}}`,
			want: &MemberCreatedWebhookEvent{},
		},
		{
			name: "member removed",
			body: `{"id":"evt_123","type":"member.removed","created_at":"` + createdAt + `","object":{"id":"mem_123","type":"member","url":"https://api.sumup.com/v0.1/merchants/M123/members/mem_123"}}`,
			want: &MemberRemovedWebhookEvent{},
		},
		{
			name: "unknown event",
			body: `{"id":"evt_123","type":"something.else","created_at":"` + createdAt + `","object":{"id":"obj_123","type":"other","url":"https://api.sumup.com/v0.1/other/obj_123"}}`,
			want: &UnknownEventNotification{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := handler.parseVerified([]byte(tt.body))
			if err != nil {
				t.Fatalf("parseVerified() error = %v", err)
			}

			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Fatalf("parseVerified() type = %T, want %T", got, tt.want)
			}

			assertWebhookEventCommonFields(t, got, client)
		})
	}
}

func TestWebhookHandler_parseVerified_errors(t *testing.T) {
	t.Parallel()

	handler := NewWebhookHandler()

	t.Run("invalid json", func(t *testing.T) {
		t.Parallel()

		_, err := handler.parseVerified([]byte("{"))
		if err == nil {
			t.Fatal("parseVerified() error = nil, want error")
		}
	})

	t.Run("invalid event payload", func(t *testing.T) {
		t.Parallel()

		payload := []byte(`{"id":"evt_123","type":"checkout.created","created_at":"not-a-time","object":{"id":"chk_123","type":"checkout","url":"https://api.sumup.com/v0.1/checkouts/chk_123"}}`)

		_, err := handler.parseVerified(payload)
		if err == nil {
			t.Fatal("parseVerified() error = nil, want error")
		}
	})
}

func TestWebhookEvent_FetchObject(t *testing.T) {
	t.Parallel()

	t.Run("fetches referenced checkout", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s, want %s", r.Method, http.MethodGet)
			}
			if r.URL.Path != "/v0.1/checkouts/chk_123" {
				t.Fatalf("path = %s, want %s", r.URL.Path, "/v0.1/checkouts/chk_123")
			}
			_, _ = w.Write([]byte(`{"id":"chk_123"}`))
		}))
		defer server.Close()

		c := &Client{
			c: clientpkg.New(clientpkg.WithBaseURL(server.URL)),
		}
		event := &WebhookEvent[Checkout]{
			Object: Object{
				URL: "/v0.1/checkouts/chk_123",
			},
			client: c,
		}

		got, err := event.FetchObject(context.Background())
		if err != nil {
			t.Fatalf("FetchObject() error = %v", err)
		}
		if got == nil || got.ID == nil || *got.ID != "chk_123" {
			t.Fatalf("FetchObject() = %+v, want checkout ID %q", got, "chk_123")
		}
	})

	t.Run("returns transport error", func(t *testing.T) {
		t.Parallel()

		c := &Client{
			c: clientpkg.New(clientpkg.WithBaseURL("http://127.0.0.1:1")),
		}
		event := &WebhookEvent[Checkout]{
			Object: Object{
				URL: "/v0.1/checkouts/chk_123",
			},
			client: c,
		}

		_, err := event.FetchObject(context.Background())
		if err == nil {
			t.Fatal("FetchObject() error = nil, want error")
		}
	})

	t.Run("returns decode error", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"id":`))
		}))
		defer server.Close()

		c := &Client{
			c: clientpkg.New(clientpkg.WithBaseURL(server.URL)),
		}
		event := &WebhookEvent[Checkout]{
			Object: Object{
				URL: "/v0.1/checkouts/chk_123",
			},
			client: c,
		}

		_, err := event.FetchObject(context.Background())
		if err == nil {
			t.Fatal("FetchObject() error = nil, want error")
		}
	})
}

func assertWebhookEventCommonFields(t *testing.T, event webhookEvent, client *Client) {
	t.Helper()

	wantCreatedAt := time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)

	switch evt := event.(type) {
	case *CheckoutCreatedWebhookEvent:
		assertWebhookEventFields(t, &evt.WebhookEvent, client, "evt_123", "checkout.created", wantCreatedAt)
	case *CheckoutProcessedWebhookEvent:
		assertWebhookEventFields(t, &evt.WebhookEvent, client, "evt_123", "checkout.processed", wantCreatedAt)
	case *CheckoutFailedWebhookEvent:
		assertWebhookEventFields(t, &evt.WebhookEvent, client, "evt_123", "checkout.failed", wantCreatedAt)
	case *CheckoutTerminatedWebhookEvent:
		assertWebhookEventFields(t, &evt.WebhookEvent, client, "evt_123", "checkout.terminated", wantCreatedAt)
	case *MemberCreatedWebhookEvent:
		assertWebhookEventFields(t, &evt.WebhookEvent, client, "evt_123", "member.created", wantCreatedAt)
	case *MemberRemovedWebhookEvent:
		assertWebhookEventFields(t, &evt.WebhookEvent, client, "evt_123", "member.removed", wantCreatedAt)
	case *UnknownEventNotification:
		webhookEvent := (*WebhookEvent[any])(evt)
		assertWebhookEventFields(t, webhookEvent, client, "evt_123", "something.else", wantCreatedAt)
	default:
		t.Fatalf("unexpected event type %T", event)
	}
}

func assertWebhookEventFields[T any](t *testing.T, event *WebhookEvent[T], client *Client, wantID, wantType string, wantCreatedAt time.Time) {
	t.Helper()

	if event.ID != wantID {
		t.Fatalf("ID = %q, want %q", event.ID, wantID)
	}
	if event.Type != wantType {
		t.Fatalf("Type = %q, want %q", event.Type, wantType)
	}
	if !event.CreatedAt.Equal(wantCreatedAt) {
		t.Fatalf("CreatedAt = %s, want %s", event.CreatedAt, wantCreatedAt)
	}
	if event.client != client {
		t.Fatal("client was not attached")
	}
}
