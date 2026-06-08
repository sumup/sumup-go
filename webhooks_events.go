package sumup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookEventType identifies the kind of SumUp webhook event payload.
type WebhookEventType string

const (
	// WebhookEventTypeCheckoutCreate identifies a `checkout.created` webhook event.
	WebhookEventTypeCheckoutCreate WebhookEventType = "checkout.created"
	// WebhookEventTypeCheckoutProcessed identifies a `checkout.processed` webhook event.
	WebhookEventTypeCheckoutProcessed WebhookEventType = "checkout.processed"
	// WebhookEventTypeCheckoutFailed identifies a `checkout.failed` webhook event.
	WebhookEventTypeCheckoutFailed WebhookEventType = "checkout.failed"
	// WebhookEventTypeCheckoutTerminated identifies a `checkout.terminated` webhook event.
	WebhookEventTypeCheckoutTerminated WebhookEventType = "checkout.terminated"
	// WebhookEventTypeMemberCreate identifies a `member.created` webhook event.
	WebhookEventTypeMemberCreate WebhookEventType = "member.created"
	// WebhookEventTypeMemberRemoved identifies a `member.removed` webhook event.
	WebhookEventTypeMemberRemoved WebhookEventType = "member.removed"
)

// WebhookEvent is the generic envelope for a SumUp webhook payload.
type WebhookEvent[T any] struct {
	// ID is the unique identifier of the webhook event.
	ID string `json:"id"`
	// Type is the event type string.
	Type string `json:"type"`
	// CreatedAt is the UTC timestamp when the event was created.
	CreatedAt time.Time `json:"created_at"`
	// Object references the related SumUp resource.
	Object Object `json:"object"`

	client *Client
}

// FetchObject retrieves the resource referenced by the webhook event.
func (we *WebhookEvent[T]) FetchObject(ctx context.Context) (*T, error) {
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

// Object describes the resource referenced by a webhook event.
type Object struct {
	// ID is the identifier of the related resource.
	ID string `json:"id"`
	// Type is the type name of the related resource.
	Type string `json:"type"`
	// URL is the canonical API URL for the related resource.
	URL string `json:"url"`
}

type webhookEvent interface {
	isWebhookEvent()
}

// UnknownEventNotification represents a webhook event whose type is not recognized by the SDK.
type UnknownEventNotification WebhookEvent[any]

func (UnknownEventNotification) isWebhookEvent() {}

// CheckoutCreatedWebhookEvent represents a `checkout.created` webhook event.
type CheckoutCreatedWebhookEvent struct {
	WebhookEvent[Checkout]
}

func (CheckoutCreatedWebhookEvent) isWebhookEvent() {}

// CheckoutProcessedWebhookEvent represents a `checkout.processed` webhook event.
type CheckoutProcessedWebhookEvent struct {
	WebhookEvent[Checkout]
}

func (CheckoutProcessedWebhookEvent) isWebhookEvent() {}

// CheckoutFailedWebhookEvent represents a `checkout.failed` webhook event.
type CheckoutFailedWebhookEvent struct {
	WebhookEvent[Checkout]
}

func (CheckoutFailedWebhookEvent) isWebhookEvent() {}

// CheckoutTerminatedWebhookEvent represents a `checkout.terminated` webhook event.
type CheckoutTerminatedWebhookEvent struct {
	WebhookEvent[Checkout]
}

func (CheckoutTerminatedWebhookEvent) isWebhookEvent() {}

// MemberCreatedWebhookEvent represents a `member.created` webhook event.
type MemberCreatedWebhookEvent struct {
	WebhookEvent[Member]
}

func (MemberCreatedWebhookEvent) isWebhookEvent() {}

// MemberRemovedWebhookEvent represents a `member.removed` webhook event.
type MemberRemovedWebhookEvent struct {
	WebhookEvent[Member]
}

func (MemberRemovedWebhookEvent) isWebhookEvent() {}

// parseVerified parses already verified webhook payload.
func (wh *webhookHandler) parseVerified(payload []byte) (webhookEvent, error) {
	var result struct {
		Type WebhookEventType `json:"type"`
	}
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}

	switch result.Type {
	case WebhookEventTypeCheckoutCreate:
		var evt CheckoutCreatedWebhookEvent
		if err := json.Unmarshal(payload, &evt); err != nil {
			return nil, err
		}
		evt.client = wh.client
		return &evt, nil
	case WebhookEventTypeCheckoutProcessed:
		var evt CheckoutProcessedWebhookEvent
		if err := json.Unmarshal(payload, &evt); err != nil {
			return nil, err
		}
		evt.client = wh.client
		return &evt, nil
	case WebhookEventTypeCheckoutFailed:
		var evt CheckoutFailedWebhookEvent
		if err := json.Unmarshal(payload, &evt); err != nil {
			return nil, err
		}
		evt.client = wh.client
		return &evt, nil
	case WebhookEventTypeCheckoutTerminated:
		var evt CheckoutTerminatedWebhookEvent
		if err := json.Unmarshal(payload, &evt); err != nil {
			return nil, err
		}
		evt.client = wh.client
		return &evt, nil
	case WebhookEventTypeMemberCreate:
		var evt MemberCreatedWebhookEvent
		if err := json.Unmarshal(payload, &evt); err != nil {
			return nil, err
		}
		evt.client = wh.client
		return &evt, nil
	case WebhookEventTypeMemberRemoved:
		var evt MemberRemovedWebhookEvent
		if err := json.Unmarshal(payload, &evt); err != nil {
			return nil, err
		}
		evt.client = wh.client
		return &evt, nil
	default:
		var evt UnknownEventNotification
		if err := json.Unmarshal(payload, &evt); err != nil {
			return nil, err
		}
		evt.client = wh.client
		return &evt, nil
	}
}
