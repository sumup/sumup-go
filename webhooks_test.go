package sumup

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestNewWebhookHandler(t *testing.T) {
	t.Run("uses environment defaults", func(t *testing.T) {
		t.Setenv("SUMUP_WEBHOOK_SECRET", "env-secret")

		handler := NewWebhookHandler()

		if handler.secret != "env-secret" {
			t.Fatalf("secret = %q, want %q", handler.secret, "env-secret")
		}
		if handler.maxSkew != DefaultWebhookMaxSkew {
			t.Fatalf("maxSkew = %v, want %v", handler.maxSkew, DefaultWebhookMaxSkew)
		}
	})

	t.Run("options override defaults", func(t *testing.T) {
		t.Setenv("SUMUP_WEBHOOK_SECRET", "env-secret")

		client := &Client{}
		handler := NewWebhookHandler(
			WithSecret("opt-secret"),
			WithMaxSkew(time.Minute),
			WithClient(client),
		)

		if handler.secret != "opt-secret" {
			t.Fatalf("secret = %q, want %q", handler.secret, "opt-secret")
		}
		if handler.maxSkew != time.Minute {
			t.Fatalf("maxSkew = %v, want %v", handler.maxSkew, time.Minute)
		}
		if handler.client != client {
			t.Fatal("client was not attached")
		}
	})
}

func TestClientWebhookHandler(t *testing.T) {
	t.Setenv("SUMUP_WEBHOOK_SECRET", "env-secret")

	client := &Client{}
	handler := client.WebhookHandler(WithMaxSkew(time.Second))

	if handler.client != client {
		t.Fatal("client was not attached")
	}
	if handler.secret != "env-secret" {
		t.Fatalf("secret = %q, want %q", handler.secret, "env-secret")
	}
	if handler.maxSkew != time.Second {
		t.Fatalf("maxSkew = %v, want %v", handler.maxSkew, time.Second)
	}
}

func TestWebhookHandler_Verify(t *testing.T) {
	t.Parallel()

	secret := "wh_sec_test"
	body := []byte(`{"id":"evt_123","type":"checkout.created"}`)
	now := time.Now().UTC()

	tests := []struct {
		name    string
		header  func() http.Header
		wantErr error
	}{
		{
			name: "happy path",
			header: func() http.Header {
				return signedHeader(secret, now, body)
			},
		},
		{
			name: "missing signature",
			header: func() http.Header {
				header := http.Header{}
				header.Set(WebhookTimestampHeader, strconv.FormatInt(now.Unix(), 10))
				return header
			},
			wantErr: ErrWebhookSignatureInvalid,
		},
		{
			name: "missing timestamp",
			header: func() http.Header {
				header := http.Header{}
				header.Set(WebhookSignatureHeader, signWebhook(secret, now, body))
				return header
			},
			wantErr: ErrWebhookTimestampInvalid,
		},
		{
			name: "invalid timestamp",
			header: func() http.Header {
				header := http.Header{}
				header.Set(WebhookTimestampHeader, "not-a-timestamp")
				header.Set(WebhookSignatureHeader, signWebhook(secret, now, body))
				return header
			},
			wantErr: ErrWebhookTimestampInvalid,
		},
		{
			name: "expired timestamp in past",
			header: func() http.Header {
				timestamp := now.Add(-(DefaultWebhookMaxSkew + time.Second))
				return signedHeader(secret, timestamp, body)
			},
			wantErr: ErrWebhookSignatureExpired,
		},
		{
			name: "expired timestamp in future",
			header: func() http.Header {
				timestamp := now.Add(DefaultWebhookMaxSkew + time.Second)
				return signedHeader(secret, timestamp, body)
			},
			wantErr: ErrWebhookSignatureExpired,
		},
		{
			name: "malformed signature",
			header: func() http.Header {
				header := http.Header{}
				header.Set(WebhookTimestampHeader, strconv.FormatInt(now.Unix(), 10))
				header.Set(WebhookSignatureHeader, "v1")
				return header
			},
			wantErr: ErrWebhookSignatureInvalid,
		},
		{
			name: "unsupported signature version",
			header: func() http.Header {
				header := signedHeader(secret, now, body)
				header.Set(WebhookSignatureHeader, "v2="+header.Get(WebhookSignatureHeader)[3:])
				return header
			},
			wantErr: ErrWebhookSignatureInvalid,
		},
		{
			name: "invalid signature digest encoding",
			header: func() http.Header {
				header := http.Header{}
				header.Set(WebhookTimestampHeader, strconv.FormatInt(now.Unix(), 10))
				header.Set(WebhookSignatureHeader, "v1=not-hex")
				return header
			},
			wantErr: ErrWebhookSignatureInvalid,
		},
		{
			name: "signature mismatch",
			header: func() http.Header {
				header := http.Header{}
				header.Set(WebhookTimestampHeader, strconv.FormatInt(now.Unix(), 10))
				header.Set(WebhookSignatureHeader, "v1=deadbeef")
				return header
			},
			wantErr: ErrWebhookSignatureInvalid,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := NewWebhookHandler(WithSecret(secret), WithMaxSkew(DefaultWebhookMaxSkew))
			err := handler.Verify(tt.header(), body)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Verify() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestWebhookHandler_Parse(t *testing.T) {
	t.Parallel()

	secret := "wh_sec_test"
	client := &Client{}
	body := []byte(`{"id":"evt_123","type":"checkout.created","created_at":"2026-04-11T10:00:00Z","object":{"id":"chk_123","type":"checkout","url":"/v0.1/checkouts/chk_123"}}`)
	now := time.Now().UTC()

	t.Run("verifies and parses event", func(t *testing.T) {
		t.Parallel()

		handler := NewWebhookHandler(WithSecret(secret), WithClient(client))

		got, err := handler.Parse(signedHeader(secret, now, body), body)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		event, ok := got.(*CheckoutCreatedWebhookEvent)
		if !ok {
			t.Fatalf("Parse() type = %T, want %T", got, &CheckoutCreatedWebhookEvent{})
		}
		if event.client != client {
			t.Fatal("client was not attached to parsed event")
		}
	})

	t.Run("returns verify error", func(t *testing.T) {
		t.Parallel()

		handler := NewWebhookHandler(WithSecret(secret))

		_, err := handler.Parse(http.Header{}, body)
		if !errors.Is(err, ErrWebhookSignatureInvalid) {
			t.Fatalf("Parse() error = %v, want %v", err, ErrWebhookSignatureInvalid)
		}
	})
}

func TestWebhookHandler_ParseRequest(t *testing.T) {
	t.Parallel()

	secret := "wh_sec_test"
	body := []byte(`{"id":"evt_123","type":"member.created","created_at":"2026-04-11T10:00:00Z","object":{"id":"mem_123","type":"member","url":"/v0.1/members/mem_123"}}`)
	now := time.Now().UTC()

	t.Run("reads verifies and parses request", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header = signedHeader(secret, now, body)

		handler := NewWebhookHandler(WithSecret(secret))

		got, err := handler.ParseRequest(req)
		if err != nil {
			t.Fatalf("ParseRequest() error = %v", err)
		}
		if _, ok := got.(*MemberCreatedWebhookEvent); !ok {
			t.Fatalf("ParseRequest() type = %T, want %T", got, &MemberCreatedWebhookEvent{})
		}
	})

	t.Run("returns body read error", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Body = io.NopCloser(errReader{err: errors.New("read body")})

		handler := NewWebhookHandler(WithSecret(secret))

		_, err := handler.ParseRequest(req)
		if err == nil || err.Error() != "read body" {
			t.Fatalf("ParseRequest() error = %v, want %q", err, "read body")
		}
	})
}

func signedHeader(secret string, timestamp time.Time, body []byte) http.Header {
	header := http.Header{}
	header.Set(WebhookTimestampHeader, strconv.FormatInt(timestamp.Unix(), 10))
	header.Set(WebhookSignatureHeader, signWebhook(secret, timestamp, body))
	return header
}

func signWebhook(secret string, timestamp time.Time, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(signedWebhookContent(timestamp, body))
	return WebhookSignatureVersion + "=" + hex.EncodeToString(mac.Sum(nil))
}

type errReader struct {
	err error
}

func (r errReader) Read(_ []byte) (int, error) {
	return 0, r.err
}
