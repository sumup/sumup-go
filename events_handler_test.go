package sumup

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

const testEventSecret = "test-secret"

func ignoreEvent(context.Context, EventNotification) error { return nil }

func testHandler(t *testing.T, fallback EventCallback) *EventsHandler {
	t.Helper()

	h, err := NewClient().EventsHandler(testEventSecret, fallback)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func signedHeader(secret string, timestamp time.Time, body []byte) http.Header {
	stamp := strconv.FormatInt(timestamp.Unix(), 10)
	// Independent concatenated representation, not the verifier's streaming implementation.
	content := append([]byte("v1:"+stamp+":"), body...)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(content)
	header := http.Header{}
	header.Set("X-SumUp-Webhook-Signature", "t="+stamp+",v1="+hex.EncodeToString(mac.Sum(nil)))
	return header
}

func handleSigned(t *testing.T, h *EventsHandler, eventType string) error {
	t.Helper()

	body := eventPayload(eventType, "https://api.sumup.com/object")
	headers := signedHeader(testEventSecret, time.Now(), body)
	return h.Handle(t.Context(), body, headers.Get(EventSignatureHeader))
}

func TestNewEventsHandler(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		client   *Client
		secret   string
		fallback EventCallback
	}{
		{"nil client", nil, testEventSecret, ignoreEvent},
		{"uninitialized client", &Client{}, testEventSecret, ignoreEvent},
		{"empty secret", NewClient(), "", ignoreEvent},
		{"nil fallback", NewClient(), testEventSecret, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewEventsHandler(tc.client, tc.secret, tc.fallback); err == nil {
				t.Fatal("expected configuration error")
			}
		})
	}
}

func TestVerifyEventSignature(t *testing.T) {
	t.Parallel()

	body := []byte(`{"id":"evt_123"}`)
	now := time.Now()
	valid := signedHeader(testEventSecret, now, body)
	stamp := strconv.FormatInt(now.Unix(), 10)
	validSignature := valid.Get(EventSignatureHeader)
	_, digest, _ := strings.Cut(validSignature, ",")
	for _, tc := range []struct {
		name, secret, signature string
		payload                 []byte
		want                    error
	}{
		{"valid", testEventSecret, validSignature, body, nil},
		{"empty secret", "", validSignature, body, ErrEventSecretMissing},
		{"wrong secret", "wrong", validSignature, body, ErrEventSignatureInvalid},
		{"missing signature", testEventSecret, "", body, ErrEventSignatureInvalid},
		{"legacy header", testEventSecret, digest, body, ErrEventSignatureInvalid},
		{"missing timestamp", testEventSecret, "t=," + digest, body, ErrEventTimestampInvalid},
		{"negative timestamp", testEventSecret, "t=-1," + digest, body, ErrEventTimestampInvalid},
		{"signed timestamp", testEventSecret, "t=+1," + digest, body, ErrEventTimestampInvalid},
		{"invalid timestamp", testEventSecret, "t=abc," + digest, body, ErrEventTimestampInvalid},
		{"overflow timestamp", testEventSecret, "t=18446744073709551615," + digest, body, ErrEventTimestampInvalid},
		{"extreme timestamp", testEventSecret, "t=9223372036854775807," + digest, body, ErrEventSignatureExpired},
		{"unsupported version", testEventSecret, "t=" + stamp + ",v2=" + digest[3:], body, ErrEventSignatureInvalid},
		{"short digest", testEventSecret, "t=" + stamp + ",v1=deadbeef", body, ErrEventSignatureInvalid},
		{"invalid hex", testEventSecret, "t=" + stamp + ",v1=" + strings.Repeat("z", 64), body, ErrEventSignatureInvalid},
		{"changed timestamp", testEventSecret, "t=" + strconv.FormatInt(now.Unix()-1, 10) + "," + digest, body, ErrEventSignatureInvalid},
		{"duplicate timestamp", testEventSecret, validSignature + ",t=" + stamp, body, ErrEventSignatureInvalid},
		{"duplicate signature", testEventSecret, validSignature + "," + digest, body, ErrEventSignatureInvalid},
		{"changed body", testEventSecret, validSignature, append(bytes.Clone(body), '\n'), ErrEventSignatureInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := VerifyEventSignature(tc.secret, tc.payload, tc.signature); !errors.Is(err, tc.want) {
				t.Errorf("error = %v, want %v", err, tc.want)
			}
		})
	}

	t.Run("fixed signature fixture", func(t *testing.T) {
		t.Parallel()
		// Shared with the sender fixture; generated independently with OpenSSL.
		const signature = "t=1234567890,v1=02e9076b318aadab2e3d14549950465512b32a100ea122b5bcb815f13d4b3153"
		if err := verifyEventSignature("test-secret", body, signature, time.Unix(1234567890, 0)); err != nil {
			t.Fatal(err)
		}
	})

	for _, skew := range []time.Duration{-eventTolerance - time.Second, -eventTolerance, eventTolerance, eventTolerance + time.Second} {
		t.Run("clock skew "+skew.String(), func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				headers := signedHeader(testEventSecret, time.Now().Add(skew), body)
				err := VerifyEventSignature(testEventSecret, body, headers.Get(EventSignatureHeader))
				var want error
				if skew < -eventTolerance || skew > eventTolerance {
					want = ErrEventSignatureExpired
				}
				if !errors.Is(err, want) {
					t.Errorf("skew %v: error = %v, want %v", skew, err, want)
				}
			})
		})
	}
}

func TestEventsHandler_Handle(t *testing.T) {
	t.Parallel()

	t.Run("routes events and permits registration from callbacks", func(t *testing.T) {
		t.Parallel()

		var handled, fallback []string
		h := testHandler(t, func(_ context.Context, event EventNotification) error {
			fallback = append(fallback, event.EventType())
			return nil
		})
		if err := h.OnMemberUpdated(func(ctx context.Context, event *MemberUpdatedEvent) error {
			if ctx != t.Context() || event.client != h.client {
				t.Error("context/client not propagated")
			}
			handled = append(handled, event.Type)
			// Registration within a callback must not deadlock or be permanently locked.
			return h.OnReaderCreated(func(_ context.Context, event *ReaderCreatedEvent) error {
				handled = append(handled, event.Type)
				return nil
			})
		}); err != nil {
			t.Fatal(err)
		}
		for _, eventType := range []string{EventTypeMemberUpdated, EventTypeReaderCreated, EventTypeMemberDeleted, "future.event"} {
			if err := handleSigned(t, h, eventType); err != nil {
				t.Fatal(err)
			}
		}
		if !slices.Equal(handled, []string{EventTypeMemberUpdated, EventTypeReaderCreated}) {
			t.Fatalf("handled = %v", handled)
		}
		if !slices.Equal(fallback, []string{EventTypeMemberDeleted, "future.event"}) {
			t.Fatalf("fallback = %v", fallback)
		}
	})

	t.Run("verification precedes parsing and callbacks", func(t *testing.T) {
		t.Parallel()

		h := testHandler(t, func(context.Context, EventNotification) error { t.Error("unexpected callback"); return nil })
		body := []byte("{")
		if err := h.Handle(t.Context(), body, "t="+strconv.FormatInt(time.Now().Unix(), 10)+",v1=deadbeef"); !errors.Is(err, ErrEventSignatureInvalid) {
			t.Fatal(err)
		}
		headers := signedHeader(testEventSecret, time.Now(), body)
		if err := h.Handle(t.Context(), body, headers.Get(EventSignatureHeader)); !errors.Is(err, ErrEventPayloadInvalid) {
			t.Fatal(err)
		}
	})

	t.Run("cancellation", func(t *testing.T) {
		t.Parallel()

		h := testHandler(t, func(context.Context, EventNotification) error { t.Error("unexpected callback"); return nil })
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		if err := h.Handle(ctx, nil, ""); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	})

	t.Run("fallback error", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("callback failed")
		h := testHandler(t, func(context.Context, EventNotification) error { return cause })
		err := handleSigned(t, h, EventTypeMemberUpdated)
		var callbackError *EventCallbackError
		if !errors.As(err, &callbackError) || !errors.Is(err, cause) {
			t.Fatalf("lost callback error: %v", err)
		}
	})

	t.Run("typed callback error", func(t *testing.T) {
		t.Parallel()

		cause := errors.New("callback failed")
		h := testHandler(t, func(context.Context, EventNotification) error { return cause })

		if err := h.OnMemberUpdated(func(context.Context, *MemberUpdatedEvent) error { return cause }); err != nil {
			t.Fatal(err)
		}

		err := handleSigned(t, h, EventTypeMemberUpdated)
		var callbackError *EventCallbackError
		if !errors.As(err, &callbackError) || !errors.Is(err, cause) {
			t.Fatalf("lost callback error: %v", err)
		}
	})

	t.Run("concurrent registration", func(t *testing.T) {
		t.Parallel()

		var calls atomic.Int64
		h := testHandler(t, func(context.Context, EventNotification) error { calls.Add(1); return nil })
		var wg sync.WaitGroup
		for range 30 {
			wg.Go(func() {
				if err := h.OnReaderCreated(func(context.Context, *ReaderCreatedEvent) error { calls.Add(1); return nil }); err != nil && !errors.Is(err, ErrEventAlreadyRegistered) {
					t.Error(err)
				}
				if err := handleSigned(t, h, EventTypeReaderCreated); err != nil {
					t.Error(err)
				}
				_ = h.registeredEventTypes()
			})
		}
		wg.Wait()
		if calls.Load() != 30 {
			t.Fatalf("calls = %d", calls.Load())
		}
	})
}

func TestEventsHandler_OnMemberUpdated(t *testing.T) {
	t.Parallel()

	t.Run("duplicate callback", func(t *testing.T) {
		t.Parallel()
		h := testHandler(t, ignoreEvent)
		callback := func(context.Context, *MemberUpdatedEvent) error { return nil }
		if err := h.OnMemberUpdated(callback); err != nil {
			t.Fatal(err)
		}
		if err := h.OnMemberUpdated(callback); !errors.Is(err, ErrEventAlreadyRegistered) {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("nil callback", func(t *testing.T) {
		t.Parallel()
		if err := testHandler(t, ignoreEvent).OnMemberUpdated(nil); err == nil {
			t.Fatal("accepted nil callback")
		}
	})
}

func TestEventsHandler_Parse(t *testing.T) {
	t.Parallel()

	t.Run("parses without invoking callbacks", func(t *testing.T) {
		t.Parallel()
		h := testHandler(t, func(context.Context, EventNotification) error { t.Error("parse called callback"); return nil })
		body := eventPayload(EventTypeMemberCreated, "https://api.sumup.com/object")
		headers := signedHeader(testEventSecret, time.Now(), body)
		event, err := h.Parse(body, headers.Get(EventSignatureHeader))
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := event.(*MemberCreatedEvent); !ok {
			t.Fatalf("event = %T", event)
		}
	})

	t.Run("rejects expired signature", func(t *testing.T) {
		t.Parallel()

		h := testHandler(t, ignoreEvent)
		body := eventPayload(EventTypeMemberCreated, "https://api.sumup.com/object")
		headers := signedHeader(testEventSecret, time.Now().Add(-6*time.Minute), body)
		if _, err := h.Parse(body, headers.Get(EventSignatureHeader)); !errors.Is(err, ErrEventSignatureExpired) {
			t.Fatal(err)
		}
	})
}
