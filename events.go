package sumup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EventNotificationType identifies the kind of SumUp event notification payload.
type EventNotificationType string

const (
	// EventNotificationTypeCheckoutCreate identifies a `checkout.created` event notification.
	EventNotificationTypeCheckoutCreate EventNotificationType = "checkout.created"
	// EventNotificationTypeCheckoutProcessed identifies a `checkout.processed` event notification.
	EventNotificationTypeCheckoutProcessed EventNotificationType = "checkout.processed"
	// EventNotificationTypeCheckoutFailed identifies a `checkout.failed` event notification.
	EventNotificationTypeCheckoutFailed EventNotificationType = "checkout.failed"
	// EventNotificationTypeCheckoutTerminated identifies a `checkout.terminated` event notification.
	EventNotificationTypeCheckoutTerminated EventNotificationType = "checkout.terminated"
	// EventNotificationTypeMemberCreate identifies a `member.created` event notification.
	EventNotificationTypeMemberCreate EventNotificationType = "member.created"
	// EventNotificationTypeMemberRemoved identifies a `member.removed` event notification.
	EventNotificationTypeMemberRemoved EventNotificationType = "member.removed"
)

// EventNotification is the generic envelope for a SumUp event payload.
type EventNotification[T any] struct {
	// ID is the unique identifier of the event notification.
	ID string `json:"id"`
	// Type is the event type string.
	Type string `json:"type"`
	// CreatedAt is the UTC timestamp when the event was created.
	CreatedAt time.Time `json:"created_at"`
	// Object references the related SumUp resource.
	Object Object `json:"object"`

	client *Client
}

// FetchObject retrieves the resource referenced by the event notification.
func (we *EventNotification[T]) FetchObject(ctx context.Context) (*T, error) {
	resp, err := we.client.c.Call(ctx, http.MethodGet, we.Object.URL)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var v T
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("decode response: %s", err.Error())
	}

	return &v, nil
}

// Object describes the resource referenced by an event notification.
type Object struct {
	// ID is the identifier of the related resource.
	ID string `json:"id"`
	// Type is the type name of the related resource.
	Type string `json:"type"`
	// URL is the canonical API URL for the related resource.
	URL string `json:"url"`
}

type eventNotification interface {
	isEventNotification()
}

// UnknownEventNotification represents an event notification whose type is not recognized by the SDK.
type UnknownEventNotification EventNotification[any]

func (UnknownEventNotification) isEventNotification() {}

// CheckoutCreatedEvent represents a `checkout.created` event notification.
type CheckoutCreatedEvent struct {
	EventNotification[Checkout]
}

func (CheckoutCreatedEvent) isEventNotification() {}

// CheckoutProcessedEvent represents a `checkout.processed` event notification.
type CheckoutProcessedEvent struct {
	EventNotification[Checkout]
}

func (CheckoutProcessedEvent) isEventNotification() {}

// CheckoutFailedEvent represents a `checkout.failed` event notification.
type CheckoutFailedEvent struct {
	EventNotification[Checkout]
}

func (CheckoutFailedEvent) isEventNotification() {}

// CheckoutTerminatedEvent represents a `checkout.terminated` event notification.
type CheckoutTerminatedEvent struct {
	EventNotification[Checkout]
}

func (CheckoutTerminatedEvent) isEventNotification() {}

// MemberCreatedEvent represents a `member.created` event notification.
type MemberCreatedEvent struct {
	EventNotification[Member]
}

func (MemberCreatedEvent) isEventNotification() {}

// MemberRemovedEvent represents a `member.removed` event notification.
type MemberRemovedEvent struct {
	EventNotification[Member]
}

func (MemberRemovedEvent) isEventNotification() {}

// parseVerified parses already verified event payload.
func (wh *eventsHandler) parseVerified(payload io.Reader) (eventNotification, error) {
	var result struct {
		Type EventNotificationType `json:"type"`
	}
	body, err := io.ReadAll(payload)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	switch result.Type {
	case EventNotificationTypeCheckoutCreate:
		var evt CheckoutCreatedEvent
		if err := json.Unmarshal(body, &evt); err != nil {
			return nil, err
		}
		evt.client = wh.client
		return &evt, nil
	case EventNotificationTypeCheckoutProcessed:
		var evt CheckoutProcessedEvent
		if err := json.Unmarshal(body, &evt); err != nil {
			return nil, err
		}
		evt.client = wh.client
		return &evt, nil
	case EventNotificationTypeCheckoutFailed:
		var evt CheckoutFailedEvent
		if err := json.Unmarshal(body, &evt); err != nil {
			return nil, err
		}
		evt.client = wh.client
		return &evt, nil
	case EventNotificationTypeCheckoutTerminated:
		var evt CheckoutTerminatedEvent
		if err := json.Unmarshal(body, &evt); err != nil {
			return nil, err
		}
		evt.client = wh.client
		return &evt, nil
	case EventNotificationTypeMemberCreate:
		var evt MemberCreatedEvent
		if err := json.Unmarshal(body, &evt); err != nil {
			return nil, err
		}
		evt.client = wh.client
		return &evt, nil
	case EventNotificationTypeMemberRemoved:
		var evt MemberRemovedEvent
		if err := json.Unmarshal(body, &evt); err != nil {
			return nil, err
		}
		evt.client = wh.client
		return &evt, nil
	default:
		var evt UnknownEventNotification
		if err := json.Unmarshal(body, &evt); err != nil {
			return nil, err
		}
		evt.client = wh.client
		return &evt, nil
	}
}
