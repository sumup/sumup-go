package sumup

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// EventSignatureHeader carries t=<unix timestamp>,v1=<hex HMAC>.
	EventSignatureHeader = "X-SumUp-Webhook-Signature"
	// EventSignatureVersion is the accepted signature scheme.
	EventSignatureVersion = "v1"
	// eventTolerance is the fixed maximum accepted clock skew.
	eventTolerance = 5 * time.Minute
)

var (
	// ErrEventSecretMissing indicates an empty signing secret.
	ErrEventSecretMissing = errors.New("missing event signing secret")
	// ErrEventTimestampInvalid indicates a missing or malformed timestamp.
	ErrEventTimestampInvalid = errors.New("invalid event timestamp")
	// ErrEventSignatureInvalid indicates a missing, malformed, or mismatched signature.
	ErrEventSignatureInvalid = errors.New("invalid event signature")
	// ErrEventSignatureExpired indicates a timestamp outside the allowed clock skew.
	ErrEventSignatureExpired = errors.New("event timestamp outside allowed tolerance")
	// ErrEventPayloadInvalid indicates an invalid notification envelope.
	ErrEventPayloadInvalid = errors.New("invalid event payload")
	// ErrEventAlreadyRegistered indicates a duplicate typed callback registration.
	ErrEventAlreadyRegistered = errors.New("event callback already registered")
)

// EventCallback handles notifications without a dedicated typed callback.
// The context is the caller's context. Events retain the handler's client for fetching resources.
// Callbacks may run concurrently and should return errors to allow delivery retries.
type EventCallback func(context.Context, EventNotification) error

// EventCallbackError distinguishes callback failures from invalid requests.
// HTTP receivers should return a 5xx response so failed processing can be retried.
type EventCallbackError struct{ Err error }

func (e *EventCallbackError) Error() string { return "event callback: " + e.Err.Error() }
func (e *EventCallbackError) Unwrap() error { return e.Err }

// EventsHandler verifies, parses, and dispatches notifications. Register typed
// callbacks using the generated On methods. Known events without a callback and
// unknown event types go to the required fallback callback.
//
// Handling and registration are safe concurrently. Callbacks run synchronously
// without the registration lock held; they must synchronize their own shared state.
// An EventsHandler must not be copied after first use.
type EventsHandler struct {
	client    *Client
	secret    string
	fallback  EventCallback
	mu        sync.RWMutex
	callbacks map[string]EventCallback
}

// String describes registered callbacks without exposing the signing secret.
func (h *EventsHandler) String() string {
	return fmt.Sprintf("EventsHandler{events:%v}", h.registeredEventTypes())
}

// GoString also masks secrets when the handler is formatted with %#v.
func (h *EventsHandler) GoString() string { return h.String() }

// NewEventsHandler creates a verified event receiver. Client, signing secret,
// and fallback are required. No environment variables are read implicitly.
func NewEventsHandler(client *Client, secret string, fallback EventCallback) (*EventsHandler, error) {
	if client == nil || client.c == nil {
		return nil, fmt.Errorf("create events handler: missing client")
	}
	if secret == "" {
		return nil, ErrEventSecretMissing
	}
	if fallback == nil {
		return nil, fmt.Errorf("create events handler: missing fallback callback")
	}
	return &EventsHandler{client: client, secret: secret, fallback: fallback, callbacks: make(map[string]EventCallback)}, nil
}

// EventsHandler creates a verified event receiver bound to this API client.
func (c *Client) EventsHandler(secret string, fallback EventCallback) (*EventsHandler, error) {
	return NewEventsHandler(c, secret, fallback)
}

func registerEvent[T EventNotification](h *EventsHandler, eventType string, callback func(context.Context, T) error) error {
	if callback == nil {
		return fmt.Errorf("register %s: nil callback", eventType)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, exists := h.callbacks[eventType]; exists {
		return fmt.Errorf("%w: %s", ErrEventAlreadyRegistered, eventType)
	}
	h.callbacks[eventType] = func(ctx context.Context, event EventNotification) error {
		return callback(ctx, event.(T))
	}
	return nil
}

func (h *EventsHandler) registeredEventTypes() []string {
	h.mu.RLock()
	types := slices.Collect(maps.Keys(h.callbacks))
	h.mu.RUnlock()
	slices.Sort(types)
	return types
}

// Handle verifies and dispatches an exact raw body. Never reserialize the body
// before calling this method. The caller owns reading and limiting the request body.
// Context cancellation propagates to the callback.
func (h *EventsHandler) Handle(ctx context.Context, payload []byte, signature string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	event, err := h.Parse(payload, signature)
	if err != nil {
		return err
	}
	h.mu.RLock()
	callback, ok := h.callbacks[event.EventType()]
	h.mu.RUnlock()
	if !ok {
		callback = h.fallback
	}
	if err := callback(ctx, event); err != nil {
		return &EventCallbackError{Err: err}
	}
	return nil
}

// Parse verifies and parses an event without invoking callbacks.
func (h *EventsHandler) Parse(payload []byte, signature string) (EventNotification, error) {
	return h.client.ParseEventNotification(h.secret, payload, signature)
}

// VerifyEventSignature verifies HMAC-SHA256 over v1:timestamp:body using the
// signing secret and a fixed five-minute tolerance. Verification uses the exact body
// bytes and constant-time digest comparison. Pass the complete signature header
// in t=<unix timestamp>,v1=<hex HMAC> format. An empty secret is always rejected.
func VerifyEventSignature(secret string, payload []byte, signature string) error {
	return verifyEventSignature(secret, payload, signature, time.Now())
}

func verifyEventSignature(secret string, payload []byte, signature string, now time.Time) error {
	if secret == "" {
		return ErrEventSecretMissing
	}
	stamp, signature, ok := strings.Cut(strings.TrimSpace(signature), ",")
	if !ok {
		return ErrEventSignatureInvalid
	}
	timestamp, ok := strings.CutPrefix(stamp, "t=")
	if !ok {
		return ErrEventTimestampInvalid
	}
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil || ts < 0 || strings.HasPrefix(timestamp, "+") {
		return ErrEventTimestampInvalid
	}
	if ts < now.Add(-eventTolerance).Unix() || ts > now.Add(eventTolerance).Unix() {
		return ErrEventSignatureExpired
	}
	version, digest, ok := strings.Cut(signature, "=")
	if !ok || version != EventSignatureVersion || len(digest) != sha256.Size*2 {
		return ErrEventSignatureInvalid
	}
	provided, err := hex.DecodeString(digest)
	if err != nil {
		return ErrEventSignatureInvalid
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = io.WriteString(mac, EventSignatureVersion+":"+timestamp+":")
	_, _ = mac.Write(payload)
	if !hmac.Equal(mac.Sum(nil), provided) {
		return ErrEventSignatureInvalid
	}
	return nil
}
