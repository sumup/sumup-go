// Code generated by `gogenitor`. DO NOT EDIT.
package sumup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// FinancialPayout is the type definition for a FinancialPayout.
type FinancialPayout struct {
	Amount          *float64               `json:"amount,omitempty"`
	Currency        *string                `json:"currency,omitempty"`
	Date            *time.Time             `json:"date,omitempty"`
	Fee             *float64               `json:"fee,omitempty"`
	Id              *int                   `json:"id,omitempty"`
	Reference       *string                `json:"reference,omitempty"`
	Status          *FinancialPayoutStatus `json:"status,omitempty"`
	TransactionCode *string                `json:"transaction_code,omitempty"`
	Type            *FinancialPayoutType   `json:"type,omitempty"`
}

type FinancialPayoutStatus string

const (
	FinancialPayoutStatusFailed     FinancialPayoutStatus = "FAILED"
	FinancialPayoutStatusSuccessful FinancialPayoutStatus = "SUCCESSFUL"
)

type FinancialPayoutType string

const (
	FinancialPayoutTypeBalanceDeduction    FinancialPayoutType = "BALANCE_DEDUCTION"
	FinancialPayoutTypeChargeBackDeduction FinancialPayoutType = "CHARGE_BACK_DEDUCTION"
	FinancialPayoutTypeDdReturnDeduction   FinancialPayoutType = "DD_RETURN_DEDUCTION"
	FinancialPayoutTypePayout              FinancialPayoutType = "PAYOUT"
	FinancialPayoutTypeRefundDeduction     FinancialPayoutType = "REFUND_DEDUCTION"
)

// FinancialPayouts is the type definition for a FinancialPayouts.
type FinancialPayouts []FinancialPayout

// ListPayoutsParams are query parameters for ListPayouts
type ListPayoutsParams struct {
	EndDate   time.Time `json:"end_date"`
	Format    *string   `json:"format,omitempty"`
	Limit     *int      `json:"limit,omitempty"`
	Order     *string   `json:"order,omitempty"`
	StartDate time.Time `json:"start_date"`
}

type PayoutsService service

// List: List payouts
// Lists ordered payouts for the merchant profile.
func (s *PayoutsService) List(ctx context.Context, params ListPayoutsParams) (*FinancialPayouts, error) {
	path := fmt.Sprintf("/v0.1/me/financials/payouts")

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

	var v FinancialPayouts
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("decode response: %s", err.Error())
	}

	return &v, nil
}
