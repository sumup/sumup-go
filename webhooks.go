package sumup

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// WebhookSignatureHeader is the header carrying the versioned webhook signature.
	WebhookSignatureHeader = "X-SumUp-Webhook-Signature"
	// WebhookTimestampHeader is the header carrying the Unix timestamp used for signing.
	WebhookTimestampHeader = "X-SumUp-Webhook-Timestamp"
	// WebhookSignatureVersion is the current webhook signature scheme version.
	WebhookSignatureVersion = "v1"
	// DefaultWebhookMaxSkew is the default maximum allowed clock skew for webhook verification.
	DefaultWebhookMaxSkew = 5 * time.Minute
)

var (
	// ErrWebhookTimestampInvalid indicates that the webhook timestamp header cannot be parsed.
	ErrWebhookTimestampInvalid = errors.New("invalid webhook timestamp")
	// ErrWebhookSignatureInvalid indicates that the webhook signature is malformed or does not match the payload.
	ErrWebhookSignatureInvalid = errors.New("invalid webhook signature")
	// ErrWebhookSignatureExpired indicates that the webhook timestamp is outside the allowed time skew window.
	ErrWebhookSignatureExpired = errors.New("webhook timestamp outside allowed time skew")
)

// WebhookOption configures webhook verification behavior.
type WebhookOption func(h *webhookHandler)

// WithSecret sets the webhook signing secret used during signature verification.
func WithSecret(secret string) WebhookOption {
	return func(h *webhookHandler) {
		h.secret = secret
	}
}

// WithMaxSkew sets the maximum allowed difference between the webhook timestamp and the local clock.
func WithMaxSkew(skew time.Duration) WebhookOption {
	return func(h *webhookHandler) {
		h.maxSkew = skew
	}
}

// WithClient attaches a SumUp client that parsed webhook events can later use to fetch their referenced objects.
func WithClient(c *Client) WebhookOption {
	return func(h *webhookHandler) {
		h.client = c
	}
}

type webhookHandler struct {
	secret  string
	client  *Client
	maxSkew time.Duration
}

// NewWebhookHandler creates a webhook handler configured from the provided options.
//
// By default it reads the signing secret from the `SUMUP_WEBHOOK_SECRET` environment variable
// and uses [DefaultWebhookMaxSkew] as the timestamp max skew.
func NewWebhookHandler(opts ...WebhookOption) *webhookHandler {
	wh := &webhookHandler{
		secret:  os.Getenv("SUMUP_WEBHOOK_SECRET"),
		maxSkew: DefaultWebhookMaxSkew,
	}

	for _, opt := range opts {
		opt(wh)
	}

	return wh
}

// WebhookHandler creates a webhook handler bound to this client.
//
// The returned handler uses the client's API transport when webhook events call
// methods such as [WebhookEvent.FetchObject]. Additional [WebhookOption] values
// can be used to override the default signing secret or timestamp max skew.
func (c *Client) WebhookHandler(opts ...WebhookOption) *webhookHandler {
	return NewWebhookHandler(append([]WebhookOption{
		WithClient(c),
	}, opts...)...)
}

// ParseRequest reads, verifies, and parses a webhook HTTP request into the most specific known event type.
func (wh *webhookHandler) ParseRequest(req *http.Request) (webhookEvent, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	return wh.Parse(req.Header, body)
}

// Parse verifies webhook headers and parses the payload into the most specific known event type.
func (wh *webhookHandler) Parse(header http.Header, body []byte) (webhookEvent, error) {
	if err := wh.Verify(header, body); err != nil {
		return nil, err
	}

	return wh.parseVerified(body)
}

// Verify verifies a webhook request.
func (wh *webhookHandler) Verify(header http.Header, body []byte) error {
	signature := header.Get(WebhookSignatureHeader)
	if signature == "" {
		return fmt.Errorf("%w: missing signature", ErrWebhookSignatureInvalid)
	}

	timestampValue := header.Get(WebhookTimestampHeader)
	if timestampValue == "" {
		return fmt.Errorf("%w: missing timestamp", ErrWebhookTimestampInvalid)
	}

	timestampRaw, err := strconv.ParseInt(timestampValue, 10, 64)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrWebhookTimestampInvalid, err)
	}

	timestamp := time.Unix(timestampRaw, 0).UTC()

	age := time.Since(timestamp)
	if age < 0 {
		age = -age
	}

	if age > wh.maxSkew {
		return ErrWebhookSignatureExpired
	}

	version, digest, found := strings.Cut(signature, "=")
	if !found || version == "" || digest == "" {
		return ErrWebhookSignatureInvalid
	}

	if version != WebhookSignatureVersion {
		return ErrWebhookSignatureInvalid
	}

	provided, err := hex.DecodeString(digest)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrWebhookSignatureInvalid, err)
	}

	mac := hmac.New(sha256.New, []byte(wh.secret))
	_, _ = mac.Write(signedWebhookContent(timestamp, body))
	if !hmac.Equal(mac.Sum(nil), provided) {
		return ErrWebhookSignatureInvalid
	}

	return nil
}

func signedWebhookContent(timestamp time.Time, body []byte) []byte {
	return []byte(WebhookSignatureVersion + ":" + strconv.FormatInt(timestamp.Unix(), 10) + ":" + string(body))
}
