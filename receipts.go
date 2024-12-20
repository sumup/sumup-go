// Code generated by `gogenitor`. DO NOT EDIT.
package sumup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Receipt is the type definition for a Receipt.
type Receipt struct {
	AcquirerData *ReceiptAcquirerData `json:"acquirer_data,omitempty"`
	EmvData      *ReceiptEmvData      `json:"emv_data,omitempty"`
	// Receipt merchant data
	MerchantData *ReceiptMerchantData `json:"merchant_data,omitempty"`
	// Transaction information.
	TransactionData *ReceiptTransaction `json:"transaction_data,omitempty"`
}

// ReceiptAcquirerData is the type definition for a ReceiptAcquirerData.
type ReceiptAcquirerData struct {
	AuthorizationCode *string `json:"authorization_code,omitempty"`
	LocalTime         *string `json:"local_time,omitempty"`
	ReturnCode        *string `json:"return_code,omitempty"`
	Tid               *string `json:"tid,omitempty"`
}

// ReceiptEmvData is the type definition for a ReceiptEmvData.
type ReceiptEmvData struct {
}

// ReceiptCard is the type definition for a ReceiptCard.
type ReceiptCard struct {
	// Card last 4 digits.
	Last4Digits *string `json:"last_4_digits,omitempty"`
	// Card Scheme.
	Type *string `json:"type,omitempty"`
}

// ReceiptEvent is the type definition for a ReceiptEvent.
type ReceiptEvent struct {
	// Amount of the event.
	Amount *AmountEvent `json:"amount,omitempty"`
	// Unique ID of the transaction event.
	Id        *EventId `json:"id,omitempty"`
	ReceiptNo *string  `json:"receipt_no,omitempty"`
	// Status of the transaction event.
	Status *EventStatus `json:"status,omitempty"`
	// Date and time of the transaction event.
	Timestamp *TimestampEvent `json:"timestamp,omitempty"`
	// Unique ID of the transaction.
	TransactionId *TransactionId `json:"transaction_id,omitempty"`
	// Type of the transaction event.
	Type *EventType `json:"type,omitempty"`
}

// ReceiptMerchantData is Receipt merchant data
type ReceiptMerchantData struct {
	Locale          *string                             `json:"locale,omitempty"`
	MerchantProfile *ReceiptMerchantDataMerchantProfile `json:"merchant_profile,omitempty"`
}

// ReceiptMerchantDataMerchantProfile is the type definition for a ReceiptMerchantDataMerchantProfile.
type ReceiptMerchantDataMerchantProfile struct {
	Address      *ReceiptMerchantDataMerchantProfileAddress `json:"address,omitempty"`
	BusinessName *string                                    `json:"business_name,omitempty"`
	Email        *string                                    `json:"email,omitempty"`
	MerchantCode *string                                    `json:"merchant_code,omitempty"`
}

// ReceiptMerchantDataMerchantProfileAddress is the type definition for a ReceiptMerchantDataMerchantProfileAddress.
type ReceiptMerchantDataMerchantProfileAddress struct {
	AddressLine1      *string `json:"address_line1,omitempty"`
	City              *string `json:"city,omitempty"`
	Country           *string `json:"country,omitempty"`
	CountryEnName     *string `json:"country_en_name,omitempty"`
	CountryNativeName *string `json:"country_native_name,omitempty"`
	Landline          *string `json:"landline,omitempty"`
	PostCode          *string `json:"post_code,omitempty"`
}

// ReceiptTransaction is Transaction information.
type ReceiptTransaction struct {
	// Transaction amount.
	Amount *string      `json:"amount,omitempty"`
	Card   *ReceiptCard `json:"card,omitempty"`
	// Transaction currency.
	Currency *string `json:"currency,omitempty"`
	// Transaction entry mode.
	EntryMode *string `json:"entry_mode,omitempty"`
	// Events
	Events *[]ReceiptEvent `json:"events,omitempty"`
	// Number of installments.
	InstallmentsCount *int `json:"installments_count,omitempty"`
	// Transaction type.
	PaymentType *string `json:"payment_type,omitempty"`
	// Products
	Products *[]ReceiptTransactionProduct `json:"products,omitempty"`
	// Receipt number
	ReceiptNo *string `json:"receipt_no,omitempty"`
	// Transaction processing status.
	Status *string `json:"status,omitempty"`
	// Time created at.
	Timestamp *time.Time `json:"timestamp,omitempty"`
	// Tip amount (included in transaction amount).
	TipAmount *string `json:"tip_amount,omitempty"`
	// Transaction code.
	TransactionCode *string `json:"transaction_code,omitempty"`
	// Transaction VAT amount.
	VatAmount *string `json:"vat_amount,omitempty"`
	// Vat rates.
	VatRates *[]ReceiptTransactionVatRate `json:"vat_rates,omitempty"`
	// Cardholder verification method.
	VerificationMethod *string `json:"verification_method,omitempty"`
}

// ReceiptTransactionProduct is the type definition for a ReceiptTransactionProduct.
type ReceiptTransactionProduct struct {
	// Product description.
	Description *string `json:"description,omitempty"`
	// Product name.
	Name *string `json:"name,omitempty"`
	// Product price.
	Price *float64 `json:"price,omitempty"`
	// Product quantity.
	Quantity *int `json:"quantity,omitempty"`
	// Quantity x product price.
	TotalPrice *float64 `json:"total_price,omitempty"`
}

// ReceiptTransactionVatRate is the type definition for a ReceiptTransactionVatRate.
type ReceiptTransactionVatRate struct {
	// Gross
	Gross *float64 `json:"gross,omitempty"`
	// Net
	Net *float64 `json:"net,omitempty"`
	// Rate
	Rate *float64 `json:"rate,omitempty"`
	// Vat
	Vat *float64 `json:"vat,omitempty"`
}

// GetReceiptParams are query parameters for GetReceipt
type GetReceiptParams struct {
	Mid       string `json:"mid"`
	TxEventId *int   `json:"tx_event_id,omitempty"`
}

type ReceiptsService service

// Get: Retrieve receipt details
// Retrieves receipt specific data for a transaction.
func (s *ReceiptsService) Get(ctx context.Context, id string, params GetReceiptParams) (*Receipt, error) {
	path := fmt.Sprintf("/v1.1/receipts/%v", id)

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

	var v Receipt
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("decode response: %s", err.Error())
	}

	return &v, nil
}
