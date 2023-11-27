// Code generated by `gogenitor`. DO NOT EDIT.
package sumup

import (
	"time"
)

// AmountEvent is Amount of the event.
type AmountEvent float64

// Card is __Required when payment type is `card`.__ Details of the payment card.
type Card struct {
	// Three or four-digit card verification value (security code) of the payment card.
	Cvv string `json:"cvv"`
	// Month from the expiration time of the payment card. Accepted format is `MM`.
	ExpiryMonth CardExpiryMonth `json:"expiry_month"`
	// Year from the expiration time of the payment card. Accepted formats are `YY` and `YYYY`.
	ExpiryYear string `json:"expiry_year"`
	// Last 4 digits of the payment card number.
	Last4Digits string `json:"last_4_digits"`
	// Name of the cardholder as it appears on the payment card.
	Name string `json:"name"`
	// Number of the payment card (without spaces).
	Number string `json:"number"`
	// Issuing card network of the payment card.
	Type CardType `json:"type"`
	// Required five-digit ZIP code. Applicable only to merchant users in the USA.
	ZipCode *string `json:"zip_code,omitempty"`
}

// Month from the expiration time of the payment card. Accepted format is `MM`.
type CardExpiryMonth string

const (
	CardExpiryMonth01 CardExpiryMonth = "01"
	CardExpiryMonth02 CardExpiryMonth = "02"
	CardExpiryMonth03 CardExpiryMonth = "03"
	CardExpiryMonth04 CardExpiryMonth = "04"
	CardExpiryMonth05 CardExpiryMonth = "05"
	CardExpiryMonth06 CardExpiryMonth = "06"
	CardExpiryMonth07 CardExpiryMonth = "07"
	CardExpiryMonth08 CardExpiryMonth = "08"
	CardExpiryMonth09 CardExpiryMonth = "09"
	CardExpiryMonth10 CardExpiryMonth = "10"
	CardExpiryMonth11 CardExpiryMonth = "11"
	CardExpiryMonth12 CardExpiryMonth = "12"
)

// Issuing card network of the payment card.
type CardType string

const (
	CardTypeAmex         CardType = "AMEX"
	CardTypeCup          CardType = "CUP"
	CardTypeDiners       CardType = "DINERS"
	CardTypeDiscover     CardType = "DISCOVER"
	CardTypeElo          CardType = "ELO"
	CardTypeElv          CardType = "ELV"
	CardTypeHipercard    CardType = "HIPERCARD"
	CardTypeJcb          CardType = "JCB"
	CardTypeMaestro      CardType = "MAESTRO"
	CardTypeMastercard   CardType = "MASTERCARD"
	CardTypeUnknown      CardType = "UNKNOWN"
	CardTypeVisa         CardType = "VISA"
	CardTypeVisaElectron CardType = "VISA_ELECTRON"
	CardTypeVisaVpay     CardType = "VISA_VPAY"
)

// Three-letter [ISO4217](https://en.wikipedia.org/wiki/ISO_4217) code of the currency for the amount. Currently supported currency values are enumerated above.
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

// Error is Error message structure.
type Error struct {
	// Platform code for the error.
	ErrorCode *string `json:"error_code,omitempty"`
	// Short description of the error.
	Message *string `json:"message,omitempty"`
}

// ErrorExtended is the type definition for a ErrorExtended.
type ErrorExtended struct {
	// Platform code for the error.
	ErrorCode *string `json:"error_code,omitempty"`
	// Short description of the error.
	Message *string `json:"message,omitempty"`
	// Parameter name (with relative location) to which the error applies. Parameters from embedded resources are displayed using dot notation. For example, `card.name` refers to the `name` parameter embedded in the `card` object.
	Param *string `json:"param,omitempty"`
}

// ErrorForbidden is Error message for forbidden requests.
type ErrorForbidden struct {
	// Platform code for the error.
	ErrorCode *string `json:"error_code,omitempty"`
	// Short description of the error.
	ErrorMessage *string `json:"error_message,omitempty"`
	// HTTP status code for the error.
	StatusCode *string `json:"status_code,omitempty"`
}

// EventId is Unique ID of the transaction event.
type EventId int64

// Status of the transaction event.
type EventStatus string

const (
	EventStatusFailed     EventStatus = "FAILED"
	EventStatusPaidOut    EventStatus = "PAID_OUT"
	EventStatusPending    EventStatus = "PENDING"
	EventStatusRefunded   EventStatus = "REFUNDED"
	EventStatusScheduled  EventStatus = "SCHEDULED"
	EventStatusSuccessful EventStatus = "SUCCESSFUL"
)

// Type of the transaction event.
type EventType string

const (
	EventTypeChargeBack      EventType = "CHARGE_BACK"
	EventTypePayout          EventType = "PAYOUT"
	EventTypePayoutDeduction EventType = "PAYOUT_DEDUCTION"
	EventTypeRefund          EventType = "REFUND"
)

// MandateResponse is Created mandate
type MandateResponse struct {
	// Merchant code which has the mandate
	MerchantCode *string `json:"merchant_code,omitempty"`
	// Mandate status
	Status *string `json:"status,omitempty"`
	// Indicates the mandate type
	Type *string `json:"type,omitempty"`
}

// Permissions is User permissions
type Permissions struct {
	// Create MOTO payments
	CreateMotoPayments *bool `json:"create_moto_payments,omitempty"`
	// Create referral
	CreateReferral *bool `json:"create_referral,omitempty"`
	// Can view full merchant transaction history
	FullTransactionHistoryView *bool `json:"full_transaction_history_view,omitempty"`
	// Refund transactions
	RefundTransactions *bool `json:"refund_transactions,omitempty"`
}

// TimestampEvent is Date and time of the transaction event.
type TimestampEvent string

// TransactionId is Unique ID of the transaction.
type TransactionId string

// TransactionMixinBase is Details of the transaction.
type TransactionMixinBase struct {
	// Total amount of the transaction.
	Amount *float64 `json:"amount,omitempty"`
	// Three-letter [ISO4217](https://en.wikipedia.org/wiki/ISO_4217) code of the currency for the amount. Currently supported currency values are enumerated above.
	Currency *Currency `json:"currency,omitempty"`
	// Unique ID of the transaction.
	Id *string `json:"id,omitempty"`
	// Current number of the installment for deferred payments.
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

// Payment type used for the transaction.
type TransactionMixinBasePaymentType string

const (
	TransactionMixinBasePaymentTypeBoleto    TransactionMixinBasePaymentType = "BOLETO"
	TransactionMixinBasePaymentTypeEcom      TransactionMixinBasePaymentType = "ECOM"
	TransactionMixinBasePaymentTypeRecurring TransactionMixinBasePaymentType = "RECURRING"
)

// Current status of the transaction.
type TransactionMixinBaseStatus string

const (
	TransactionMixinBaseStatusCancelled  TransactionMixinBaseStatus = "CANCELLED"
	TransactionMixinBaseStatusFailed     TransactionMixinBaseStatus = "FAILED"
	TransactionMixinBaseStatusPending    TransactionMixinBaseStatus = "PENDING"
	TransactionMixinBaseStatusSuccessful TransactionMixinBaseStatus = "SUCCESSFUL"
)

// TransactionMixinCheckout is the type definition for a TransactionMixinCheckout.
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

// Entry mode of the payment details.
type TransactionMixinCheckoutEntryMode string

const (
	TransactionMixinCheckoutEntryModeBoleto        TransactionMixinCheckoutEntryMode = "BOLETO"
	TransactionMixinCheckoutEntryModeCustomerEntry TransactionMixinCheckoutEntryMode = "CUSTOMER_ENTRY"
)
