// Code generated by `gogenitor`. DO NOT EDIT.
package sumup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CardResponse is Details of the payment card.
type CardResponse struct {
	// Three-digit card verification value (security code) of the payment card.
	Cvv *string `json:"cvv,omitempty"`
	// Month from the expiration time of the payment card. Accepted format is `MM`.
	ExpiryMonth *CardResponseExpiryMonth `json:"expiry_month,omitempty"`
	// Year from the expiration time of the payment card. Accepted formats are `YY` and `YYYY`.
	ExpiryYear *string `json:"expiry_year,omitempty"`
	// Last 4 digits of the payment card number.
	Last4Digits *string `json:"last_4_digits,omitempty"`
	// Name of the cardholder as it appears on the payment card.
	Name *string `json:"name,omitempty"`
	// Number of the payment card (without spaces).
	Number *string `json:"number,omitempty"`
	// Issuing card network of the payment card.
	Type *CardResponseType `json:"type,omitempty"`
	// Required five-digit ZIP code. Applicable only to merchant users in the USA.
	ZipCode *string `json:"zip_code,omitempty"`
}

// Month from the expiration time of the payment card. Accepted format is `MM`.
type CardResponseExpiryMonth string

const (
	CardResponseExpiryMonth01 CardResponseExpiryMonth = "01"
	CardResponseExpiryMonth02 CardResponseExpiryMonth = "02"
	CardResponseExpiryMonth03 CardResponseExpiryMonth = "03"
	CardResponseExpiryMonth04 CardResponseExpiryMonth = "04"
	CardResponseExpiryMonth05 CardResponseExpiryMonth = "05"
	CardResponseExpiryMonth06 CardResponseExpiryMonth = "06"
	CardResponseExpiryMonth07 CardResponseExpiryMonth = "07"
	CardResponseExpiryMonth08 CardResponseExpiryMonth = "08"
	CardResponseExpiryMonth09 CardResponseExpiryMonth = "09"
	CardResponseExpiryMonth10 CardResponseExpiryMonth = "10"
	CardResponseExpiryMonth11 CardResponseExpiryMonth = "11"
	CardResponseExpiryMonth12 CardResponseExpiryMonth = "12"
)

// Issuing card network of the payment card.
type CardResponseType string

const (
	CardResponseTypeAmex         CardResponseType = "AMEX"
	CardResponseTypeCup          CardResponseType = "CUP"
	CardResponseTypeDiners       CardResponseType = "DINERS"
	CardResponseTypeDiscover     CardResponseType = "DISCOVER"
	CardResponseTypeElo          CardResponseType = "ELO"
	CardResponseTypeElv          CardResponseType = "ELV"
	CardResponseTypeHipercard    CardResponseType = "HIPERCARD"
	CardResponseTypeJcb          CardResponseType = "JCB"
	CardResponseTypeMaestro      CardResponseType = "MAESTRO"
	CardResponseTypeMastercard   CardResponseType = "MASTERCARD"
	CardResponseTypeUnknown      CardResponseType = "UNKNOWN"
	CardResponseTypeVisa         CardResponseType = "VISA"
	CardResponseTypeVisaElectron CardResponseType = "VISA_ELECTRON"
	CardResponseTypeVisaVpay     CardResponseType = "VISA_VPAY"
)

// Event is the type definition for a Event.
type Event struct {
	// Amount of the event.
	Amount *AmountEvent `json:"amount,omitempty"`
	// Amount deducted for the event.
	DeductedAmount *float64 `json:"deducted_amount,omitempty"`
	// Amount of the fee deducted for the event.
	DeductedFeeAmount *float64 `json:"deducted_fee_amount,omitempty"`
	// Amount of the fee related to the event.
	FeeAmount *float64 `json:"fee_amount,omitempty"`
	// Unique ID of the transaction event.
	Id *EventId `json:"id,omitempty"`
	// Consecutive number of the installment.
	InstallmentNumber *int `json:"installment_number,omitempty"`
	// Status of the transaction event.
	Status *EventStatus `json:"status,omitempty"`
	// Date and time of the transaction event.
	Timestamp *TimestampEvent `json:"timestamp,omitempty"`
	// Unique ID of the transaction.
	TransactionId *TransactionId `json:"transaction_id,omitempty"`
	// Type of the transaction event.
	Type *EventType `json:"type,omitempty"`
}

// FinancialTransaction is the type definition for a FinancialTransaction.
type FinancialTransaction struct {
	Amount            *float64                  `json:"amount,omitempty"`
	Currency          *string                   `json:"currency,omitempty"`
	ExternalReference *string                   `json:"external_reference,omitempty"`
	Id                *string                   `json:"id,omitempty"`
	Timestamp         *string                   `json:"timestamp,omitempty"`
	TransactionCode   *string                   `json:"transaction_code,omitempty"`
	Type              *FinancialTransactionType `json:"type,omitempty"`
}

type FinancialTransactionType string

const (
	FinancialTransactionTypeChargeBack       FinancialTransactionType = "CHARGE_BACK"
	FinancialTransactionTypeDdReturn         FinancialTransactionType = "DD_RETURN"
	FinancialTransactionTypeDdReturnReversal FinancialTransactionType = "DD_RETURN_REVERSAL"
	FinancialTransactionTypeRefund           FinancialTransactionType = "REFUND"
	FinancialTransactionTypeSale             FinancialTransactionType = "SALE"
)

// FinancialTransactions is the type definition for a FinancialTransactions.
type FinancialTransactions []FinancialTransaction

// HorizontalAccuracy is Indication of the precision of the geographical position received from the payment terminal.
type HorizontalAccuracy float64

// Lat is Latitude value from the coordinates of the payment location (as received from the payment terminal reader).
type Lat float64

// Link is Details of a link to a related resource.
type Link struct {
	// URL for accessing the related resource.
	Href *string `json:"href,omitempty"`
	// Specifies the relation to the current resource.
	Rel *string `json:"rel,omitempty"`
	// Specifies the media type of the related resource.
	Type *string `json:"type,omitempty"`
}

// LinkRefund is the type definition for a LinkRefund.
type LinkRefund struct {
	// URL for accessing the related resource.
	Href *string `json:"href,omitempty"`
	// Maximum allowed amount for the refund.
	MaxAmount *float64 `json:"max_amount,omitempty"`
	// Minimum allowed amount for the refund.
	MinAmount *float64 `json:"min_amount,omitempty"`
	// Specifies the relation to the current resource.
	Rel *string `json:"rel,omitempty"`
	// Specifies the media type of the related resource.
	Type *string `json:"type,omitempty"`
}

// Lon is Longitude value from the coordinates of the payment location (as received from the payment terminal reader).
type Lon float64

// Product is Details of the product for which the payment is made.
type Product struct {
	// Name of the product from the merchant's catalog.
	Name *string `json:"name,omitempty"`
	// Price of the product without VAT.
	Price *float64 `json:"price,omitempty"`
	// Price of a single product item with VAT.
	PriceWithVat *float64 `json:"price_with_vat,omitempty"`
	// Number of product items for the purchase.
	Quantity *float64 `json:"quantity,omitempty"`
	// Amount of the VAT for a single product item (calculated as the product of `price` and `vat_rate`, i.e. `single_vat_amount = price * vat_rate`).
	SingleVatAmount *float64 `json:"single_vat_amount,omitempty"`
	// Total price of the product items without VAT (calculated as the product of `price` and `quantity`, i.e. `total_price = price * quantity`).
	TotalPrice *float64 `json:"total_price,omitempty"`
	// Total price of the product items including VAT (calculated as the product of `price_with_vat` and `quantity`, i.e. `total_with_vat = price_with_vat * quantity`).
	TotalWithVat *float64 `json:"total_with_vat,omitempty"`
	// Total VAT amount for the purchase (calculated as the product of `single_vat_amount` and `quantity`, i.e. `vat_amount = single_vat_amount * quantity`).
	VatAmount *float64 `json:"vat_amount,omitempty"`
	// VAT rate applicable to the product.
	VatRate *float64 `json:"vat_rate,omitempty"`
}

// TransactionEvent is Details of a transaction event.
type TransactionEvent struct {
	// Amount of the event.
	Amount *AmountEvent `json:"amount,omitempty"`
	// Date when the transaction event occurred.
	Date *time.Time `json:"date,omitempty"`
	// Date when the transaction event is due to occur.
	DueDate *time.Time `json:"due_date,omitempty"`
	// Type of the transaction event.
	EventType *EventType `json:"event_type,omitempty"`
	// Unique ID of the transaction event.
	Id *EventId `json:"id,omitempty"`
	// Consequtive number of the installment that is paid. Applicable only payout events, i.e. `event_type = PAYOUT`.
	InstallmentNumber *int `json:"installment_number,omitempty"`
	// Status of the transaction event.
	Status *EventStatus `json:"status,omitempty"`
	// Date and time of the transaction event.
	Timestamp *TimestampEvent `json:"timestamp,omitempty"`
}

// Payment type used for the transaction.
type TransactionFullPaymentType string

const (
	TransactionFullPaymentTypeBoleto    TransactionFullPaymentType = "BOLETO"
	TransactionFullPaymentTypeEcom      TransactionFullPaymentType = "ECOM"
	TransactionFullPaymentTypeRecurring TransactionFullPaymentType = "RECURRING"
)

// Current status of the transaction.
type TransactionFullStatus string

const (
	TransactionFullStatusCancelled  TransactionFullStatus = "CANCELLED"
	TransactionFullStatusFailed     TransactionFullStatus = "FAILED"
	TransactionFullStatusPending    TransactionFullStatus = "PENDING"
	TransactionFullStatusSuccessful TransactionFullStatus = "SUCCESSFUL"
)

// Entry mode of the payment details.
type TransactionFullEntryMode string

const (
	TransactionFullEntryModeBoleto        TransactionFullEntryMode = "BOLETO"
	TransactionFullEntryModeCustomerEntry TransactionFullEntryMode = "CUSTOMER_ENTRY"
)

// Payout plan of the registered user at the time when the transaction was made.
type TransactionFullPayoutPlan string

const (
	TransactionFullPayoutPlanAcceleratedInstallment TransactionFullPayoutPlan = "ACCELERATED_INSTALLMENT"
	TransactionFullPayoutPlanSinglePayment          TransactionFullPayoutPlan = "SINGLE_PAYMENT"
	TransactionFullPayoutPlanTrueInstallment        TransactionFullPayoutPlan = "TRUE_INSTALLMENT"
)

// TransactionFullLocation is Details of the payment location as received from the payment terminal.
type TransactionFullLocation struct {
	// Indication of the precision of the geographical position received from the payment terminal.
	HorizontalAccuracy *HorizontalAccuracy `json:"horizontal_accuracy,omitempty"`
	// Latitude value from the coordinates of the payment location (as received from the payment terminal reader).
	Lat *Lat `json:"lat,omitempty"`
	// Longitude value from the coordinates of the payment location (as received from the payment terminal reader).
	Lon *Lon `json:"lon,omitempty"`
}

// Payout type for the transaction.
type TransactionFullPayoutType string

const (
	TransactionFullPayoutTypeBalance     TransactionFullPayoutType = "BALANCE"
	TransactionFullPayoutTypeBankAccount TransactionFullPayoutType = "BANK_ACCOUNT"
	TransactionFullPayoutTypePrepaidCard TransactionFullPayoutType = "PREPAID_CARD"
)

// Simple name of the payment type.
type TransactionFullSimplePaymentType string

const (
	TransactionFullSimplePaymentTypeCash              TransactionFullSimplePaymentType = "CASH"
	TransactionFullSimplePaymentTypeCcCustomerEntered TransactionFullSimplePaymentType = "CC_CUSTOMER_ENTERED"
	TransactionFullSimplePaymentTypeCcSignature       TransactionFullSimplePaymentType = "CC_SIGNATURE"
	TransactionFullSimplePaymentTypeElv               TransactionFullSimplePaymentType = "ELV"
	TransactionFullSimplePaymentTypeEmv               TransactionFullSimplePaymentType = "EMV"
	TransactionFullSimplePaymentTypeManualEntry       TransactionFullSimplePaymentType = "MANUAL_ENTRY"
	TransactionFullSimplePaymentTypeMoto              TransactionFullSimplePaymentType = "MOTO"
)

// Status generated from the processing status and the latest transaction state.
type TransactionFullSimpleStatus string

const (
	TransactionFullSimpleStatusCancelled     TransactionFullSimpleStatus = "CANCELLED"
	TransactionFullSimpleStatusCancelFailed  TransactionFullSimpleStatus = "CANCEL_FAILED"
	TransactionFullSimpleStatusChargeback    TransactionFullSimpleStatus = "CHARGEBACK"
	TransactionFullSimpleStatusFailed        TransactionFullSimpleStatus = "FAILED"
	TransactionFullSimpleStatusNonCollection TransactionFullSimpleStatus = "NON_COLLECTION"
	TransactionFullSimpleStatusPaidOut       TransactionFullSimpleStatus = "PAID_OUT"
	TransactionFullSimpleStatusRefunded      TransactionFullSimpleStatus = "REFUNDED"
	TransactionFullSimpleStatusRefundFailed  TransactionFullSimpleStatus = "REFUND_FAILED"
	TransactionFullSimpleStatusSuccessful    TransactionFullSimpleStatus = "SUCCESSFUL"
)

// Verification method used for the transaction.
type TransactionFullVerificationMethod string

const (
	TransactionFullVerificationMethodConfirmationCodeVerified TransactionFullVerificationMethod = "confirmation code verified"
	TransactionFullVerificationMethodNone                     TransactionFullVerificationMethod = "none"
	TransactionFullVerificationMethodOfflinePin               TransactionFullVerificationMethod = "offline pin"
	TransactionFullVerificationMethodOfflinePinSignature      TransactionFullVerificationMethod = "offline pin + signature"
	TransactionFullVerificationMethodOnlinePin                TransactionFullVerificationMethod = "online pin"
	TransactionFullVerificationMethodSignature                TransactionFullVerificationMethod = "signature"
)

// TransactionFull is the type definition for a TransactionFull.
type TransactionFull struct {
	// Total amount of the transaction.
	Amount *float64 `json:"amount,omitempty"`
	// Authorization code for the transaction sent by the payment card issuer or bank. Applicable only to card payments.
	AuthCode *string `json:"auth_code,omitempty"`
	// Details of the payment card.
	Card *CardResponse `json:"card,omitempty"`
	// Three-letter [ISO4217](https://en.wikipedia.org/wiki/ISO_4217) code of the currency for the amount. Currently supported currency values are enumerated above.
	Currency *Currency `json:"currency,omitempty"`
	// Entry mode of the payment details.
	EntryMode *TransactionFullEntryMode `json:"entry_mode,omitempty"`
	// List of events related to the transaction.
	Events *[]Event `json:"events,omitempty"`
	// Indication of the precision of the geographical position received from the payment terminal.
	HorizontalAccuracy *HorizontalAccuracy `json:"horizontal_accuracy,omitempty"`
	// Unique ID of the transaction.
	Id *string `json:"id,omitempty"`
	// Current number of the installment for deferred payments.
	InstallmentsCount *int `json:"installments_count,omitempty"`
	// Internal unique ID of the transaction on the SumUp platform.
	InternalId *int `json:"internal_id,omitempty"`
	// Latitude value from the coordinates of the payment location (as received from the payment terminal reader).
	Lat *Lat `json:"lat,omitempty"`
	// List of hyperlinks for accessing related resources.
	Links *[]interface{} `json:"links,omitempty"`
	// Local date and time of the creation of the transaction.
	LocalTime *time.Time `json:"local_time,omitempty"`
	// Details of the payment location as received from the payment terminal.
	Location *TransactionFullLocation `json:"location,omitempty"`
	// Longitude value from the coordinates of the payment location (as received from the payment terminal reader).
	Lon *Lon `json:"lon,omitempty"`
	// Unique code of the registered merchant to whom the payment is made.
	MerchantCode *string `json:"merchant_code,omitempty"`
	// Payment type used for the transaction.
	PaymentType *TransactionFullPaymentType `json:"payment_type,omitempty"`
	// Payout plan of the registered user at the time when the transaction was made.
	PayoutPlan *TransactionFullPayoutPlan `json:"payout_plan,omitempty"`
	// Payout type for the transaction.
	PayoutType *TransactionFullPayoutType `json:"payout_type,omitempty"`
	// Number of payouts that are made to the registered user specified in the `user` property.
	PayoutsReceived *int `json:"payouts_received,omitempty"`
	// Total number of payouts to the registered user specified in the `user` property.
	PayoutsTotal *int `json:"payouts_total,omitempty"`
	// Short description of the payment. The value is taken from the `description` property of the related checkout resource.
	ProductSummary *string `json:"product_summary,omitempty"`
	// List of products from the merchant's catalogue for which the transaction serves as a payment.
	Products *[]Product `json:"products,omitempty"`
	// Simple name of the payment type.
	SimplePaymentType *TransactionFullSimplePaymentType `json:"simple_payment_type,omitempty"`
	// Status generated from the processing status and the latest transaction state.
	SimpleStatus *TransactionFullSimpleStatus `json:"simple_status,omitempty"`
	// Current status of the transaction.
	Status *TransactionFullStatus `json:"status,omitempty"`
	// Indicates whether tax deduction is enabled for the transaction.
	TaxEnabled *bool `json:"tax_enabled,omitempty"`
	// Date and time of the creation of the transaction. Response format expressed according to [ISO8601](https://en.wikipedia.org/wiki/ISO_8601) code.
	Timestamp *time.Time `json:"timestamp,omitempty"`
	// Amount of the tip (out of the total transaction amount).
	TipAmount *float64 `json:"tip_amount,omitempty"`
	// Transaction code returned by the acquirer/processing entity after processing the transaction.
	TransactionCode *string `json:"transaction_code,omitempty"`
	// List of transaction events related to the transaction.
	TransactionEvents *[]TransactionEvent `json:"transaction_events,omitempty"`
	// Email address of the registered user (merchant) to whom the payment is made.
	Username *string `json:"username,omitempty"`
	// Amount of the applicable VAT (out of the total transaction amount).
	VatAmount *float64 `json:"vat_amount,omitempty"`
	// List of VAT rates applicable to the transaction.
	VatRates *[]interface{} `json:"vat_rates,omitempty"`
	// Verification method used for the transaction.
	VerificationMethod *TransactionFullVerificationMethod `json:"verification_method,omitempty"`
}

// Payment type used for the transaction.
type TransactionHistoryPaymentType string

const (
	TransactionHistoryPaymentTypeBoleto    TransactionHistoryPaymentType = "BOLETO"
	TransactionHistoryPaymentTypeEcom      TransactionHistoryPaymentType = "ECOM"
	TransactionHistoryPaymentTypeRecurring TransactionHistoryPaymentType = "RECURRING"
)

// Current status of the transaction.
type TransactionHistoryStatus string

const (
	TransactionHistoryStatusCancelled  TransactionHistoryStatus = "CANCELLED"
	TransactionHistoryStatusFailed     TransactionHistoryStatus = "FAILED"
	TransactionHistoryStatusPending    TransactionHistoryStatus = "PENDING"
	TransactionHistoryStatusSuccessful TransactionHistoryStatus = "SUCCESSFUL"
)

// Payout plan of the registered user at the time when the transaction was made.
type TransactionHistoryPayoutPlan string

const (
	TransactionHistoryPayoutPlanAcceleratedInstallment TransactionHistoryPayoutPlan = "ACCELERATED_INSTALLMENT"
	TransactionHistoryPayoutPlanSinglePayment          TransactionHistoryPayoutPlan = "SINGLE_PAYMENT"
	TransactionHistoryPayoutPlanTrueInstallment        TransactionHistoryPayoutPlan = "TRUE_INSTALLMENT"
)

// Issuing card network of the payment card used for the transaction.
type TransactionHistoryCardType string

const (
	TransactionHistoryCardTypeAmex         TransactionHistoryCardType = "AMEX"
	TransactionHistoryCardTypeCup          TransactionHistoryCardType = "CUP"
	TransactionHistoryCardTypeDiners       TransactionHistoryCardType = "DINERS"
	TransactionHistoryCardTypeDiscover     TransactionHistoryCardType = "DISCOVER"
	TransactionHistoryCardTypeElo          TransactionHistoryCardType = "ELO"
	TransactionHistoryCardTypeElv          TransactionHistoryCardType = "ELV"
	TransactionHistoryCardTypeHipercard    TransactionHistoryCardType = "HIPERCARD"
	TransactionHistoryCardTypeJcb          TransactionHistoryCardType = "JCB"
	TransactionHistoryCardTypeMaestro      TransactionHistoryCardType = "MAESTRO"
	TransactionHistoryCardTypeMastercard   TransactionHistoryCardType = "MASTERCARD"
	TransactionHistoryCardTypeUnknown      TransactionHistoryCardType = "UNKNOWN"
	TransactionHistoryCardTypeVisa         TransactionHistoryCardType = "VISA"
	TransactionHistoryCardTypeVisaElectron TransactionHistoryCardType = "VISA_ELECTRON"
	TransactionHistoryCardTypeVisaVpay     TransactionHistoryCardType = "VISA_VPAY"
)

// Type of the transaction for the registered user specified in the `user` property.
type TransactionHistoryType string

const (
	TransactionHistoryTypeChargeBack TransactionHistoryType = "CHARGE_BACK"
	TransactionHistoryTypePayment    TransactionHistoryType = "PAYMENT"
	TransactionHistoryTypeRefund     TransactionHistoryType = "REFUND"
)

// TransactionHistory is the type definition for a TransactionHistory.
type TransactionHistory struct {
	// Total amount of the transaction.
	Amount *float64 `json:"amount,omitempty"`
	// Issuing card network of the payment card used for the transaction.
	CardType *TransactionHistoryCardType `json:"card_type,omitempty"`
	// Three-letter [ISO4217](https://en.wikipedia.org/wiki/ISO_4217) code of the currency for the amount. Currently supported currency values are enumerated above.
	Currency *Currency `json:"currency,omitempty"`
	// Unique ID of the transaction.
	Id *string `json:"id,omitempty"`
	// Current number of the installment for deferred payments.
	InstallmentsCount *int `json:"installments_count,omitempty"`
	// Payment type used for the transaction.
	PaymentType *TransactionHistoryPaymentType `json:"payment_type,omitempty"`
	// Payout plan of the registered user at the time when the transaction was made.
	PayoutPlan *TransactionHistoryPayoutPlan `json:"payout_plan,omitempty"`
	// Number of payouts that are made to the registered user specified in the `user` property.
	PayoutsReceived *int `json:"payouts_received,omitempty"`
	// Total number of payouts to the registered user specified in the `user` property.
	PayoutsTotal *int `json:"payouts_total,omitempty"`
	// Short description of the payment. The value is taken from the `description` property of the related checkout resource.
	ProductSummary *string `json:"product_summary,omitempty"`
	// Current status of the transaction.
	Status *TransactionHistoryStatus `json:"status,omitempty"`
	// Date and time of the creation of the transaction. Response format expressed according to [ISO8601](https://en.wikipedia.org/wiki/ISO_8601) code.
	Timestamp *time.Time `json:"timestamp,omitempty"`
	// Transaction code returned by the acquirer/processing entity after processing the transaction.
	TransactionCode *string `json:"transaction_code,omitempty"`
	// Unique ID of the transaction.
	TransactionId *TransactionId `json:"transaction_id,omitempty"`
	// Type of the transaction for the registered user specified in the `user` property.
	Type *TransactionHistoryType `json:"type,omitempty"`
	// Email address of the registered user (merchant) to whom the payment is made.
	User *string `json:"user,omitempty"`
}

// TransactionMixinHistory is the type definition for a TransactionMixinHistory.
type TransactionMixinHistory struct {
	// Payout plan of the registered user at the time when the transaction was made.
	PayoutPlan *TransactionMixinHistoryPayoutPlan `json:"payout_plan,omitempty"`
	// Number of payouts that are made to the registered user specified in the `user` property.
	PayoutsReceived *int `json:"payouts_received,omitempty"`
	// Total number of payouts to the registered user specified in the `user` property.
	PayoutsTotal *int `json:"payouts_total,omitempty"`
	// Short description of the payment. The value is taken from the `description` property of the related checkout resource.
	ProductSummary *string `json:"product_summary,omitempty"`
}

// Payout plan of the registered user at the time when the transaction was made.
type TransactionMixinHistoryPayoutPlan string

const (
	TransactionMixinHistoryPayoutPlanAcceleratedInstallment TransactionMixinHistoryPayoutPlan = "ACCELERATED_INSTALLMENT"
	TransactionMixinHistoryPayoutPlanSinglePayment          TransactionMixinHistoryPayoutPlan = "SINGLE_PAYMENT"
	TransactionMixinHistoryPayoutPlanTrueInstallment        TransactionMixinHistoryPayoutPlan = "TRUE_INSTALLMENT"
)

// ListFinancialTransactionsParams are query parameters for ListFinancialTransactions
type ListFinancialTransactionsParams struct {
	EndDate   time.Time `json:"end_date"`
	Format    *string   `json:"format,omitempty"`
	Limit     *int      `json:"limit,omitempty"`
	Order     *string   `json:"order,omitempty"`
	StartDate time.Time `json:"start_date"`
}

// RefundTransaction request body.
type RefundTransactionBody struct {
	// Amount to be refunded. Eligible amount can't exceed the amount of the transaction and varies based on country and currency. If you do not specify a value, the system performs a full refund of the transaction.
	Amount *float64 `json:"amount,omitempty"`
}

// RefundTransactionResponse is the type definition for a RefundTransactionResponse.
type RefundTransactionResponse struct {
}

// GetTransactionParams are query parameters for GetTransaction
type GetTransactionParams struct {
	Id              *string `json:"id,omitempty"`
	InternalId      *string `json:"internal_id,omitempty"`
	TransactionCode *string `json:"transaction_code,omitempty"`
}

// ListTransactionsParams are query parameters for ListTransactions
type ListTransactionsParams struct {
	ChangesSince    *time.Time `json:"changes_since,omitempty"`
	Limit           *int       `json:"limit,omitempty"`
	NewestRef       *string    `json:"newest_ref,omitempty"`
	NewestTime      *time.Time `json:"newest_time,omitempty"`
	OldestRef       *string    `json:"oldest_ref,omitempty"`
	OldestTime      *time.Time `json:"oldest_time,omitempty"`
	Order           *string    `json:"order,omitempty"`
	PaymentTypes    *[]string  `json:"payment_types,omitempty"`
	Statuses        *[]string  `json:"statuses,omitempty"`
	TransactionCode *string    `json:"transaction_code,omitempty"`
	Types           *[]string  `json:"types,omitempty"`
	Users           *[]string  `json:"users,omitempty"`
}

// ListTransactionsResponse is the type definition for a ListTransactionsResponse.
type ListTransactionsResponse struct {
	Items *[]TransactionHistory `json:"items,omitempty"`
	Links *[]Link               `json:"links,omitempty"`
}

type TransactionsService service

// List: List financial transactions
// Lists a less detailed history of all transactions associated with the merchant profile.
func (s *TransactionsService) List(ctx context.Context, params ListFinancialTransactionsParams) (*FinancialTransactions, error) {
	path := fmt.Sprintf("/v0.1/me/financials/transactions")

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error building request: %v", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("invalid response: %d - %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	dec := json.NewDecoder(resp.Body)
	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := dec.Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("read error response: %s", err.Error())
		}

		return nil, &apiErr
	}

	var v FinancialTransactions
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("decode response: %s", err.Error())
	}

	return &v, nil
}

// Refund: Refund a transaction
// Refunds an identified transaction either in full or partially.
func (s *TransactionsService) Refund(ctx context.Context, txnId string, body RefundTransactionBody) (*RefundTransactionResponse, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return nil, fmt.Errorf("encoding json body request failed: %v", err)
	}

	path := fmt.Sprintf("/v0.1/me/refund/%v", txnId)

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, buf)
	if err != nil {
		return nil, fmt.Errorf("error building request: %v", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("invalid response: %d - %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	dec := json.NewDecoder(resp.Body)
	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := dec.Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("read error response: %s", err.Error())
		}

		return nil, &apiErr
	}

	var v RefundTransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("decode response: %s", err.Error())
	}

	return &v, nil
}

// Get: Retrieve a transaction
// Retrieves the full details of an identified transaction. The transaction resource is identified by a query parameter and *one* of following parameters is required:
//   - `id`
//   - `internal_id`
//   - `transaction_code`
//   - `foreign_transaction_id`
//   - `client_transaction_id`
func (s *TransactionsService) Get(ctx context.Context, params GetTransactionParams) (*TransactionFull, error) {
	path := fmt.Sprintf("/v0.1/me/transactions")

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error building request: %v", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("invalid response: %d - %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	dec := json.NewDecoder(resp.Body)
	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := dec.Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("read error response: %s", err.Error())
		}

		return nil, &apiErr
	}

	var v TransactionFull
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("decode response: %s", err.Error())
	}

	return &v, nil
}

// ListDetailed: List transactions
// Lists detailed history of all transactions associated with the merchant profile.
func (s *TransactionsService) ListDetailed(ctx context.Context, params ListTransactionsParams) (*ListTransactionsResponse, error) {
	path := fmt.Sprintf("/v0.1/me/transactions/history")

	req, err := s.client.NewRequest(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error building request: %v", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("invalid response: %d - %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	dec := json.NewDecoder(resp.Body)
	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := dec.Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("read error response: %s", err.Error())
		}

		return nil, &apiErr
	}

	var v ListTransactionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("decode response: %s", err.Error())
	}

	return &v, nil
}
