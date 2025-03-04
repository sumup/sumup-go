// Code generated by `go-sdk-gen`. DO NOT EDIT.

package shared

import (
	"encoding/json"
	"fmt"
	"time"
)

// AmountEvent: Amount of the event.
type AmountEvent float64

// Attributes: Object attributes that modifiable only by SumUp applications.
type Attributes map[string]any

// Currency: Three-letter [ISO4217](https://en.wikipedia.org/wiki/ISO_4217) code of the currency for the amount.
// Currently supported currency values are enumerated above.
type Currency string

const (
	CurrencyBgn Currency = "BGN"
	CurrencyBrl Currency = "BRL"
	CurrencyChf Currency = "CHF"
	CurrencyClp Currency = "CLP"
	CurrencyCzk Currency = "CZK"
	CurrencyDkk Currency = "DKK"
	CurrencyEur Currency = "EUR"
	CurrencyGbp Currency = "GBP"
	CurrencyHrk Currency = "HRK"
	CurrencyHuf Currency = "HUF"
	CurrencyNok Currency = "NOK"
	CurrencyPln Currency = "PLN"
	CurrencyRon Currency = "RON"
	CurrencySek Currency = "SEK"
	CurrencyUsd Currency = "USD"
)

// Error: Error message structure.
type Error struct {
	// Platform code for the error.
	ErrorCode *string `json:"error_code,omitempty"`
	// Short description of the error.
	Message *string `json:"message,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("error_code=%v, message=%v", e.ErrorCode, e.Message)
}

var _ error = (*Error)(nil)

// ErrorForbidden: Error message for forbidden requests.
type ErrorForbidden struct {
	// Platform code for the error.
	ErrorCode *string `json:"error_code,omitempty"`
	// Short description of the error.
	ErrorMessage *string `json:"error_message,omitempty"`
	// HTTP status code for the error.
	StatusCode *string `json:"status_code,omitempty"`
}

func (e *ErrorForbidden) Error() string {
	return fmt.Sprintf("error_code=%v, error_message=%v, status_code=%v", e.ErrorCode, e.ErrorMessage, e.StatusCode)
}

var _ error = (*ErrorForbidden)(nil)

// EventId: Unique ID of the transaction event.
// Format: int64
type EventId int64

// EventStatus: Status of the transaction event.
type EventStatus string

const (
	EventStatusFailed     EventStatus = "FAILED"
	EventStatusPaidOut    EventStatus = "PAID_OUT"
	EventStatusPending    EventStatus = "PENDING"
	EventStatusRefunded   EventStatus = "REFUNDED"
	EventStatusScheduled  EventStatus = "SCHEDULED"
	EventStatusSuccessful EventStatus = "SUCCESSFUL"
)

// EventType: Type of the transaction event.
type EventType string

const (
	EventTypeChargeBack      EventType = "CHARGE_BACK"
	EventTypePayout          EventType = "PAYOUT"
	EventTypePayoutDeduction EventType = "PAYOUT_DEDUCTION"
	EventTypeRefund          EventType = "REFUND"
)

// Invite: Pending invitation for membership.
type Invite struct {
	// Email address of the invited user.
	// Format: email
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
}

// MandateResponse: Created mandate
type MandateResponse struct {
	// Merchant code which has the mandate
	MerchantCode *string `json:"merchant_code,omitempty"`
	// Mandate status
	Status *string `json:"status,omitempty"`
	// Indicates the mandate type
	Type *string `json:"type,omitempty"`
}

// MembershipStatus: The status of the membership.
type MembershipStatus string

const (
	MembershipStatusAccepted MembershipStatus = "accepted"
	MembershipStatusDisabled MembershipStatus = "disabled"
	MembershipStatusExpired  MembershipStatus = "expired"
	MembershipStatusPending  MembershipStatus = "pending"
	MembershipStatusUnknown  MembershipStatus = "unknown"
)

// Metadata: Set of user-defined key-value pairs attached to the object. Partial updates are not supported. When
// updating, always submit whole metadata.
type Metadata map[string]any

// TimestampEvent: Date and time of the transaction event.
type TimestampEvent string

// TransactionId: Unique ID of the transaction.
type TransactionId string

// TransactionMixinBase: Details of the transaction.
type TransactionMixinBase struct {
	// Total amount of the transaction.
	Amount *float64 `json:"amount,omitempty"`
	// Three-letter [ISO4217](https://en.wikipedia.org/wiki/ISO_4217) code of the currency for the amount. Currently supported
	// currency values are enumerated above.
	Currency *Currency `json:"currency,omitempty"`
	// Unique ID of the transaction.
	Id *string `json:"id,omitempty"`
	// Current number of the installment for deferred payments.
	// Min: 1
	InstallmentsCount *int `json:"installments_count,omitempty"`
	// Payment type used for the transaction.
	PaymentType *TransactionMixinBasePaymentType `json:"payment_type,omitempty"`
	// Current status of the transaction.
	Status *TransactionMixinBaseStatus `json:"status,omitempty"`
	// Date and time of the creation of the transaction. Response format expressed according to [ISO8601](https://en.wikipedia.org/wiki/ISO_8601) code.
	Timestamp *time.Time `json:"timestamp,omitempty"`
	// Transaction code returned by the acquirer/processing entity after processing the transaction.
	TransactionCode *string `json:"transaction_code,omitempty"`
}

// TransactionMixinBasePaymentType: Payment type used for the transaction.
type TransactionMixinBasePaymentType string

const (
	TransactionMixinBasePaymentTypeBoleto    TransactionMixinBasePaymentType = "BOLETO"
	TransactionMixinBasePaymentTypeEcom      TransactionMixinBasePaymentType = "ECOM"
	TransactionMixinBasePaymentTypeRecurring TransactionMixinBasePaymentType = "RECURRING"
)

// TransactionMixinBaseStatus: Current status of the transaction.
type TransactionMixinBaseStatus string

const (
	TransactionMixinBaseStatusCancelled  TransactionMixinBaseStatus = "CANCELLED"
	TransactionMixinBaseStatusFailed     TransactionMixinBaseStatus = "FAILED"
	TransactionMixinBaseStatusPending    TransactionMixinBaseStatus = "PENDING"
	TransactionMixinBaseStatusSuccessful TransactionMixinBaseStatus = "SUCCESSFUL"
)

// TransactionMixinCheckout is a schema definition.
type TransactionMixinCheckout struct {
	// Authorization code for the transaction sent by the payment card issuer or bank. Applicable only to card payments.
	AuthCode *string `json:"auth_code,omitempty"`
	// Entry mode of the payment details.
	EntryMode *TransactionMixinCheckoutEntryMode `json:"entry_mode,omitempty"`
	// Internal unique ID of the transaction on the SumUp platform.
	InternalId *int `json:"internal_id,omitempty"`
	// Unique code of the registered merchant to whom the payment is made.
	MerchantCode *string `json:"merchant_code,omitempty"`
	// Amount of the tip (out of the total transaction amount).
	TipAmount *float64 `json:"tip_amount,omitempty"`
	// Amount of the applicable VAT (out of the total transaction amount).
	VatAmount *float64 `json:"vat_amount,omitempty"`
}

// TransactionMixinCheckoutEntryMode: Entry mode of the payment details.
type TransactionMixinCheckoutEntryMode string

const (
	TransactionMixinCheckoutEntryModeBoleto        TransactionMixinCheckoutEntryMode = "BOLETO"
	TransactionMixinCheckoutEntryModeCustomerEntry TransactionMixinCheckoutEntryMode = "CUSTOMER_ENTRY"
)

type Date struct{ time.Time }

func (d Date) String() string {
	return d.Format(time.DateOnly)
}

const jsonDateFormat = `"` + time.DateOnly + `"`

var _ json.Unmarshaler = (*Date)(nil)

func (d *Date) UnmarshalJSON(b []byte) (err error) {
	date, err := time.Parse(jsonDateFormat, string(b))
	if err != nil {
		return err
	}
	d.Time = date
	return
}

var _ json.Marshaler = (*Date)(nil)

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(d.Time.Format(jsonDateFormat)), nil
}

type Time struct{ time.Time }

func (t Time) String() string {
	return t.Format(time.TimeOnly)
}

const jsonTimeFormat = `"` + time.TimeOnly + `"`

var _ json.Unmarshaler = (*Time)(nil)

func (t *Time) UnmarshalJSON(b []byte) (err error) {
	date, err := time.Parse(jsonTimeFormat, string(b))
	if err != nil {
		return err
	}
	t.Time = date
	return
}

var _ json.Marshaler = (*Time)(nil)

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(t.Time.Format(jsonTimeFormat)), nil
}
