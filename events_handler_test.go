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

func TestNewEventHandler(t *testing.T) {
	t.Run("uses environment defaults", func(t *testing.T) {
		t.Setenv("SUMUP_EVENT_SECRET", "env-secret")

		handler := NewEventsHandler()

		if handler.secret != "env-secret" {
			t.Fatalf("secret = %q, want %q", handler.secret, "env-secret")
		}
		if handler.maxSkew != DefaultEventMaxSkew {
			t.Fatalf("maxSkew = %v, want %v", handler.maxSkew, DefaultEventMaxSkew)
		}
	})

	t.Run("options override defaults", func(t *testing.T) {
		t.Setenv("SUMUP_EVENT_SECRET", "env-secret")

		client := &Client{}
		handler := NewEventsHandler(
			WithSecret("opt-secret"),
			WithMaxSkew(time.Minute),
			WithInsecureSkipVerify(),
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
		if !handler.insecureSkipVerify {
			t.Fatal("insecureSkipVerify was not enabled")
		}
	})
}

func TestClientEventHandler(t *testing.T) {
	t.Setenv("SUMUP_EVENT_SECRET", "env-secret")

	client := &Client{}
	handler := client.EventHandler(WithMaxSkew(time.Second))

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

func TestEventHandler_Verify(t *testing.T) {
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
				header.Set(EventTimestampHeader, strconv.FormatInt(now.Unix(), 10))
				return header
			},
			wantErr: ErrEventSignatureInvalid,
		},
		{
			name: "missing timestamp",
			header: func() http.Header {
				header := http.Header{}
				header.Set(EventSignatureHeader, signEvent(secret, now, body))
				return header
			},
			wantErr: ErrEventTimestampInvalid,
		},
		{
			name: "invalid timestamp",
			header: func() http.Header {
				header := http.Header{}
				header.Set(EventTimestampHeader, "not-a-timestamp")
				header.Set(EventSignatureHeader, signEvent(secret, now, body))
				return header
			},
			wantErr: ErrEventTimestampInvalid,
		},
		{
			name: "expired timestamp in past",
			header: func() http.Header {
				timestamp := now.Add(-(DefaultEventMaxSkew + time.Second))
				return signedHeader(secret, timestamp, body)
			},
			wantErr: ErrEventSignatureExpired,
		},
		{
			name: "expired timestamp in future",
			header: func() http.Header {
				timestamp := now.Add(DefaultEventMaxSkew + time.Second)
				return signedHeader(secret, timestamp, body)
			},
			wantErr: ErrEventSignatureExpired,
		},
		{
			name: "malformed signature",
			header: func() http.Header {
				header := http.Header{}
				header.Set(EventTimestampHeader, strconv.FormatInt(now.Unix(), 10))
				header.Set(EventSignatureHeader, "v1")
				return header
			},
			wantErr: ErrEventSignatureInvalid,
		},
		{
			name: "unsupported signature version",
			header: func() http.Header {
				header := signedHeader(secret, now, body)
				header.Set(EventSignatureHeader, "v2="+header.Get(EventSignatureHeader)[3:])
				return header
			},
			wantErr: ErrEventSignatureInvalid,
		},
		{
			name: "invalid signature digest encoding",
			header: func() http.Header {
				header := http.Header{}
				header.Set(EventTimestampHeader, strconv.FormatInt(now.Unix(), 10))
				header.Set(EventSignatureHeader, "v1=not-hex")
				return header
			},
			wantErr: ErrEventSignatureInvalid,
		},
		{
			name: "signature mismatch",
			header: func() http.Header {
				header := http.Header{}
				header.Set(EventTimestampHeader, strconv.FormatInt(now.Unix(), 10))
				header.Set(EventSignatureHeader, "v1=deadbeef")
				return header
			},
			wantErr: ErrEventSignatureInvalid,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := NewEventsHandler(WithSecret(secret), WithMaxSkew(DefaultEventMaxSkew))
			err := handler.Verify(tt.header(), bytes.NewReader(body))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Verify() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestEventHandler_Parse(t *testing.T) {
	t.Parallel()

	secret := "wh_sec_test"
	client := &Client{}
	body := []byte(`{"id":"evt_123","type":"checkout.created","created_at":"2026-04-11T10:00:00Z","object":{"id":"chk_123","type":"checkout","url":"/v0.1/checkouts/chk_123"}}`)
	now := time.Now().UTC()

	t.Run("verifies and parses event", func(t *testing.T) {
		t.Parallel()

		handler := NewEventsHandler(WithSecret(secret), WithClient(client))

		got, err := handler.Parse(signedHeader(secret, now, body), bytes.NewReader(body))
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		event, ok := got.(*CheckoutCreatedEvent)
		if !ok {
			t.Fatalf("Parse() type = %T, want %T", got, &CheckoutCreatedEvent{})
		}
		if event.client != client {
			t.Fatal("client was not attached to parsed event")
		}
	})

	t.Run("returns verify error", func(t *testing.T) {
		t.Parallel()

		handler := NewEventsHandler(WithSecret(secret))

		_, err := handler.Parse(http.Header{}, bytes.NewReader(body))
		if !errors.Is(err, ErrEventSignatureInvalid) {
			t.Fatalf("Parse() error = %v, want %v", err, ErrEventSignatureInvalid)
		}
	})

	t.Run("skips verification when configured", func(t *testing.T) {
		t.Parallel()

		handler := NewEventsHandler(WithSecret(secret), WithInsecureSkipVerify())

		got, err := handler.Parse(http.Header{}, bytes.NewReader(body))
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if _, ok := got.(*CheckoutCreatedEvent); !ok {
			t.Fatalf("Parse() type = %T, want %T", got, &CheckoutCreatedEvent{})
		}
	})
}

func TestEventHandler_ParseRequest(t *testing.T) {
	t.Parallel()

	secret := "wh_sec_test"
	body := []byte(`{"id":"evt_123","type":"member.created","created_at":"2026-04-11T10:00:00Z","object":{"id":"mem_123","type":"member","url":"/v0.1/members/mem_123"}}`)
	now := time.Now().UTC()

	t.Run("reads verifies and parses request", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header = signedHeader(secret, now, body)

		handler := NewEventsHandler(WithSecret(secret))

		got, err := handler.ParseRequest(req)
		if err != nil {
			t.Fatalf("ParseRequest() error = %v", err)
		}
		if _, ok := got.(*MemberCreatedEvent); !ok {
			t.Fatalf("ParseRequest() type = %T, want %T", got, &MemberCreatedEvent{})
		}
	})

	t.Run("returns body read error", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Body = io.NopCloser(errReader{err: errors.New("read body")})
		req.Header = signedHeader(secret, now, nil)

		handler := NewEventsHandler(WithSecret(secret))

		_, err := handler.ParseRequest(req)
		if err == nil || err.Error() != "read body" {
			t.Fatalf("ParseRequest() error = %v, want %q", err, "read body")
		}
	})

	t.Run("skips request verification when configured", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

		handler := NewEventsHandler(WithSecret(secret), WithInsecureSkipVerify())

		got, err := handler.ParseRequest(req)
		if err != nil {
			t.Fatalf("ParseRequest() error = %v", err)
		}
		if _, ok := got.(*MemberCreatedEvent); !ok {
			t.Fatalf("ParseRequest() type = %T, want %T", got, &MemberCreatedEvent{})
		}
	})
}

func signedHeader(secret string, timestamp time.Time, body []byte) http.Header {
	header := http.Header{}
	header.Set(EventTimestampHeader, strconv.FormatInt(timestamp.Unix(), 10))
	header.Set(EventSignatureHeader, signEvent(secret, timestamp, body))
	return header
}

func signEvent(secret string, timestamp time.Time, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(signedEventContent(timestamp, body))
	return EventSignatureVersion + "=" + hex.EncodeToString(mac.Sum(nil))
}

type errReader struct {
	err error
}

func (r errReader) Read(_ []byte) (int, error) {
	return 0, r.err
}
