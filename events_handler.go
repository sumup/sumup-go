package sumup

import (
	"bytes"
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
	// EventSignatureHeader is the header carrying the versioned event signature.
	EventSignatureHeader = "X-SumUp-Event-Signature"
	// EventTimestampHeader is the header carrying the Unix timestamp used for signing.
	EventTimestampHeader = "X-SumUp-Event-Timestamp"
	// EventSignatureVersion is the current event signature scheme version.
	EventSignatureVersion = "v1"
	// DefaultEventMaxSkew is the default maximum allowed clock skew for event verification.
	DefaultEventMaxSkew = 5 * time.Minute
)

var (
	// ErrEventTimestampInvalid indicates that the event timestamp header cannot be parsed.
	ErrEventTimestampInvalid = errors.New("invalid event timestamp")
	// ErrEventSignatureInvalid indicates that the event signature is malformed or does not match the payload.
	ErrEventSignatureInvalid = errors.New("invalid event signature")
	// ErrEventSignatureExpired indicates that the event timestamp is outside the allowed time skew window.
	ErrEventSignatureExpired = errors.New("event timestamp outside allowed time skew")
)

// EventOption configures event verification behavior.
type EventOption func(h *eventsHandler)

// WithSecret sets the event signing secret used during signature verification.
// By default, the value of `SUMUP_EVENT_SECRET` environment variable is used as the
// events signing secret.
func WithSecret(secret string) EventOption {
	return func(h *eventsHandler) {
		h.secret = secret
	}
}

// WithMaxSkew sets the maximum allowed difference between the event timestamp and the local clock.
// The default maximum allowed skew is 5 minutes.
func WithMaxSkew(skew time.Duration) EventOption {
	return func(h *eventsHandler) {
		h.maxSkew = skew
	}
}

// WithInsecureSkipVerify disables event signature verification during parsing.
func WithInsecureSkipVerify() EventOption {
	return func(h *eventsHandler) {
		h.insecureSkipVerify = true
	}
}

// WithClient attaches a SumUp client that parsed event notifications can later use to fetch their referenced objects.
func WithClient(c *Client) EventOption {
	return func(h *eventsHandler) {
		h.client = c
	}
}

type eventsHandler struct {
	secret             string
	client             *Client
	maxSkew            time.Duration
	insecureSkipVerify bool
}

// NewEventsHandler creates an event handler configured from the provided options.
//
// By default it reads the signing secret from the `SUMUP_EVENT_SECRET` environment variable
// and uses [DefaultEventMaxSkew] as the timestamp max skew.
func NewEventsHandler(opts ...EventOption) *eventsHandler {
	wh := &eventsHandler{
		secret:  os.Getenv("SUMUP_EVENT_SECRET"),
		maxSkew: DefaultEventMaxSkew,
	}

	for _, opt := range opts {
		opt(wh)
	}

	return wh
}

// EventHandler creates an event handler bound to this client.
//
// The returned handler uses the client's API transport when event notifications call
// methods such as [EventNotification.FetchObject]. Additional [EventOption] values
// can be used to override the default signing secret or timestamp max skew.
func (c *Client) EventHandler(opts ...EventOption) *eventsHandler {
	return NewEventsHandler(append([]EventOption{
		WithClient(c),
	}, opts...)...)
}

// ParseRequest reads, verifies, and parses an event HTTP request into the most specific known event type.
func (wh *eventsHandler) ParseRequest(req *http.Request) (eventNotification, error) {
	return wh.Parse(req.Header, req.Body)
}

// Parse verifies event headers and parses the payload into the most specific known event type.
func (wh *eventsHandler) Parse(header http.Header, body io.Reader) (eventNotification, error) {
	if wh.insecureSkipVerify {
		return wh.parseVerified(body)
	}

	var payload bytes.Buffer
	verifierBody := io.TeeReader(body, &payload)

	if err := wh.Verify(header, verifierBody); err != nil {
		return nil, err
	}

	return wh.parseVerified(&payload)
}

// Verify verifies an event request.
func (wh *eventsHandler) Verify(header http.Header, body io.Reader) error {
	signature := header.Get(EventSignatureHeader)
	if signature == "" {
		return fmt.Errorf("%w: missing signature", ErrEventSignatureInvalid)
	}

	timestampValue := header.Get(EventTimestampHeader)
	if timestampValue == "" {
		return fmt.Errorf("%w: missing timestamp", ErrEventTimestampInvalid)
	}

	timestampRaw, err := strconv.ParseInt(timestampValue, 10, 64)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrEventTimestampInvalid, err)
	}

	timestamp := time.Unix(timestampRaw, 0).UTC()

	age := time.Since(timestamp)
	if age < 0 {
		age = -age
	}

	if age > wh.maxSkew {
		return ErrEventSignatureExpired
	}

	version, digest, found := strings.Cut(signature, "=")
	if !found || version == "" || digest == "" {
		return ErrEventSignatureInvalid
	}

	if version != EventSignatureVersion {
		return ErrEventSignatureInvalid
	}

	provided, err := hex.DecodeString(digest)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrEventSignatureInvalid, err)
	}

	mac := hmac.New(sha256.New, []byte(wh.secret))
	if err := writeSignedEventContent(mac, timestamp, body); err != nil {
		return err
	}
	if !hmac.Equal(mac.Sum(nil), provided) {
		return ErrEventSignatureInvalid
	}

	return nil
}

func signedEventContent(timestamp time.Time, body []byte) []byte {
	var buf bytes.Buffer
	_ = writeSignedEventContent(&buf, timestamp, bytes.NewReader(body))
	return buf.Bytes()
}

func writeSignedEventContent(w io.Writer, timestamp time.Time, body io.Reader) error {
	if _, err := io.WriteString(w, EventSignatureVersion+":"+strconv.FormatInt(timestamp.Unix(), 10)+":"); err != nil {
		return err
	}
	_, err := io.Copy(w, body)
	return err
}
