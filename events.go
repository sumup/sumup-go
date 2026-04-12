package sumup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/sumup/sumup-go/client"
)

// EventNotification is a known or unknown SumUp notification. Use a type switch
// to access a specific event, or register typed callbacks with [EventsHandler].
type EventNotification interface {
	EventID() string
	EventType() string

	isEventNotification()
}

// TypedEvent contains common notification fields and fetches its resource as T.
// Concrete event types are generated from the OpenAPI webhooks.
// Notifications may be retried or arrive out of order: make processing idempotent
// and fetch the latest resource when reconciling local state.
type TypedEvent[T any] struct {
	client *Client

	// ID identifies this notification, including across retries.
	ID string `json:"id"`
	// Type is the wire event name, such as members.updated.
	Type string `json:"type"`
	// CreatedAt is the time when the event was created.
	CreatedAt time.Time `json:"created_at"`
	// Object references the affected API resource.
	Object EventObject `json:"object"`
}

// EventID returns the notification identifier.
func (e TypedEvent[T]) EventID() string { return e.ID }

// EventType returns the wire event name.
func (e TypedEvent[T]) EventType() string { return e.Type }

// EventObject references the API resource affected by a notification.
type EventObject struct {
	// ID identifies the referenced resource.
	ID string `json:"id"`
	// Type is the resource kind, such as member or reader.
	Type string `json:"type"`
	// URL is the absolute API URL for fetching the resource.
	URL string `json:"url"`
}

// UnknownEvent preserves notifications not yet recognized by this SDK.
// FetchObject returns the resource as raw JSON.
type UnknownEvent struct{ TypedEvent[json.RawMessage] }

func (*UnknownEvent) isEventNotification() {}

// FetchObject retrieves the latest resource using the client that parsed the
// notification.
//
// The URL must have the SumUp API origin. Its path and query are resolved against
// the client's configured base URL; URL credentials and fragments are ignored.
// Requests use the client's authentication and redirect policy.
// Deleted resources may return a *Problem API error.
func (e TypedEvent[T]) FetchObject(ctx context.Context) (*T, error) {
	if e.client == nil || e.client.c == nil {
		return nil, fmt.Errorf("fetch event object: event is not bound to a client")
	}

	u, err := url.Parse(e.Object.URL)
	if err != nil {
		return nil, fmt.Errorf("parse event object URL: %w", err)
	}

	if u.Scheme+"://"+u.Host != client.APIUrl {
		return nil, fmt.Errorf("fetch event object: expected a URL on %s", client.APIUrl)
	}

	resp, err := e.client.c.Call(ctx, http.MethodGet, u.EscapedPath(), client.WithQueryValues(u.Query()))
	if err != nil {
		return nil, fmt.Errorf("fetch event object: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		var apiErr Problem
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil && (apiErr.Type != "" || apiErr.Title != nil || apiErr.Detail != nil || apiErr.Status != nil) {
			if apiErr.Status == nil {
				apiErr.Status = &resp.StatusCode
			}
			return nil, fmt.Errorf("fetch event object: %w", &apiErr)
		}
		return nil, fmt.Errorf("fetch event object: unexpected response %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var object T
	if err := json.NewDecoder(resp.Body).Decode(&object); err != nil {
		return nil, fmt.Errorf("decode event object: %w", err)
	}

	return &object, nil
}

// ParseEventNotification verifies the signature and timestamp before parsing the
// exact raw body. Prefer [Client.EventsHandler] when using callback dispatch.
func (c *Client) ParseEventNotification(secret string, payload []byte, signature string) (EventNotification, error) {
	if err := VerifyEventSignature(secret, payload, signature); err != nil {
		return nil, err
	}
	return c.parseEventNotification(payload)
}

// ParseEventNotificationWithoutVerification skips signature verification.
// Use this for fixtures or payloads verified before being queued for later processing.
// Prefer [Client.ParseEventNotification] for incoming requests.
func (c *Client) ParseEventNotificationWithoutVerification(payload []byte) (EventNotification, error) {
	return c.parseEventNotification(payload)
}

func (c *Client) parseEventNotification(payload []byte) (EventNotification, error) {
	var raw TypedEvent[json.RawMessage]
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrEventPayloadInvalid, err)
	}
	if raw.ID == "" || raw.Type == "" || raw.CreatedAt.IsZero() || raw.Object.ID == "" || raw.Object.Type == "" || raw.Object.URL == "" {
		return nil, fmt.Errorf("%w: missing required event fields", ErrEventPayloadInvalid)
	}
	raw.client = c
	return parseKnownEvent(raw), nil
}
