// Code generated by `gogenitor`. DO NOT EDIT.
package sumup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// FinancialPayout is a schema definition.
type FinancialPayout struct {
	Amount   *float64 `json:"amount,omitempty"`
	Currency *string  `json:"currency,omitempty"`
	// Format: date
	Date            *Date                  `json:"date,omitempty"`
	Fee             *float64               `json:"fee,omitempty"`
	Id              *int                   `json:"id,omitempty"`
	Reference       *string                `json:"reference,omitempty"`
	Status          *FinancialPayoutStatus `json:"status,omitempty"`
	TransactionCode *string                `json:"transaction_code,omitempty"`
	Type            *FinancialPayoutType   `json:"type,omitempty"`
}

// FinancialPayoutStatus is a schema definition.
type FinancialPayoutStatus string

const (
	FinancialPayoutStatusFailed     FinancialPayoutStatus = "FAILED"
	FinancialPayoutStatusSuccessful FinancialPayoutStatus = "SUCCESSFUL"
)

// FinancialPayoutType is a schema definition.
type FinancialPayoutType string

const (
	FinancialPayoutTypeBalanceDeduction    FinancialPayoutType = "BALANCE_DEDUCTION"
	FinancialPayoutTypeChargeBackDeduction FinancialPayoutType = "CHARGE_BACK_DEDUCTION"
	FinancialPayoutTypeDdReturnDeduction   FinancialPayoutType = "DD_RETURN_DEDUCTION"
	FinancialPayoutTypePayout              FinancialPayoutType = "PAYOUT"
	FinancialPayoutTypeRefundDeduction     FinancialPayoutType = "REFUND_DEDUCTION"
)

// FinancialPayouts is a schema definition.
type FinancialPayouts []FinancialPayout

// ListPayoutsParams: query parameters for ListPayouts
type ListPayoutsParams struct {
	// End date (in [ISO8601](https://en.wikipedia.org/wiki/ISO_8601) format).
	EndDate Date
	Format  *string
	Limit   *int
	Order   *string
	// Start date (in [ISO8601](https://en.wikipedia.org/wiki/ISO_8601) format).
	StartDate Date
}

// QueryValues converts [ListPayoutsParams] into [url.Values].
func (p *ListPayoutsParams) QueryValues() url.Values {
	q := make(url.Values)

	q.Set("end_date", p.EndDate.Format(time.DateOnly))

	if p.Format != nil {
		q.Set("format", *p.Format)
	}

	if p.Limit != nil {
		q.Set("limit", strconv.Itoa(*p.Limit))
	}

	if p.Order != nil {
		q.Set("order", *p.Order)
	}

	q.Set("start_date", p.StartDate.Format(time.DateOnly))

	return q
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
	req.URL.RawQuery = params.QueryValues().Encode()

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var v FinancialPayouts
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			return nil, fmt.Errorf("decode response: %s", err.Error())
		}

		return &v, nil
	case http.StatusUnauthorized:
		var apiErr Error
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("read error response: %s", err.Error())
		}

		return nil, &apiErr
	default:
		return nil, fmt.Errorf("unexpected response %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
}
