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

// Apikey is the type definition for a Apikey.
type Apikey struct {
	CreatedAt time.Time `json:"created_at"`
	// Unique identifier of the API Key.
	Id string `json:"id"`
	// User-assigned name of the API Key.
	Name string `json:"name"`
	// The plaintext value of the API key. This field is returned only in the response to API key creation and is never again available in the plaintext form.
	Plaintext *string `json:"plaintext,omitempty"`
	// Last 8 characters of the API key.
	Preview   string       `json:"preview"`
	Scopes    Oauth2Scopes `json:"scopes"`
	Type      ApikeyType   `json:"type"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type ApikeyType string

const (
	ApikeyTypePublic ApikeyType = "public"
	ApikeyTypeSecret ApikeyType = "secret"
)

// ApikeysList is the type definition for a ApikeysList.
type ApikeysList struct {
	Items      []Apikey `json:"items"`
	TotalCount int      `json:"total_count"`
}

type Oauth2Scope string

const (
	Oauth2ScopeAccountingRead      Oauth2Scope = "accounting.read"
	Oauth2ScopeAccountingWrite     Oauth2Scope = "accounting.write"
	Oauth2ScopeEmail               Oauth2Scope = "email"
	Oauth2ScopeInvoicesRead        Oauth2Scope = "invoices.read"
	Oauth2ScopeInvoicesWrite       Oauth2Scope = "invoices.write"
	Oauth2ScopePaymentInstruments  Oauth2Scope = "payment_instruments"
	Oauth2ScopePayments            Oauth2Scope = "payments"
	Oauth2ScopeProducts            Oauth2Scope = "products"
	Oauth2ScopeProfile             Oauth2Scope = "profile"
	Oauth2ScopeTerminalsRead       Oauth2Scope = "terminals.read"
	Oauth2ScopeTerminalsWrite      Oauth2Scope = "terminals.write"
	Oauth2ScopeTransactionsHistory Oauth2Scope = "transactions.history"
	Oauth2ScopeUserAppSettings     Oauth2Scope = "user.app-settings"
	Oauth2ScopeUserPayoutSettings  Oauth2Scope = "user.payout-settings"
	Oauth2ScopeUserProfile         Oauth2Scope = "user.profile"
	Oauth2ScopeUserProfileReadonly Oauth2Scope = "user.profile_readonly"
	Oauth2ScopeUserSubaccounts     Oauth2Scope = "user.subaccounts"
)

// Oauth2Scopes is the type definition for a Oauth2Scopes.
type Oauth2Scopes []Oauth2Scope

// ListApikeysParams are query parameters for ListApikeys
type ListApikeysParams struct {
	Limit  *int `json:"limit,omitempty"`
	Offset *int `json:"offset,omitempty"`
}

// CreateAPIKey request body.
type CreateApikeyBody struct {
	Name   string       `json:"name"`
	Scopes Oauth2Scopes `json:"scopes"`
}

// UpdateAPIKey request body.
type UpdateApikeyBody struct {
	// New name for the API key.
	Name   string       `json:"name"`
	Scopes Oauth2Scopes `json:"scopes"`
}

type ApiKeysService service

// ListApikeys: List API keys
// Returns paginated list of API keys.
func (s *ApiKeysService) ListApikeys(ctx context.Context, merchantCode string, params ListApikeysParams) (*ApikeysList, error) {
	path := fmt.Sprintf("/v0.1/merchants/%v/api-keys", merchantCode)

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

	var v ApikeysList
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("decode response: %s", err.Error())
	}

	return &v, nil
}

// CreateApikey: Create an API key
// Creates a new API key for the user.
func (s *ApiKeysService) CreateApikey(ctx context.Context, merchantCode string, body CreateApikeyBody) (*Apikey, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return nil, fmt.Errorf("encoding json body request failed: %v", err)
	}

	path := fmt.Sprintf("/v0.1/merchants/%v/api-keys", merchantCode)

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

	var v Apikey
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("decode response: %s", err.Error())
	}

	return &v, nil
}

// RevokeApikey: Revoke an API key
// Revokes an API key.
func (s *ApiKeysService) RevokeApikey(ctx context.Context, merchantCode string, keyId string) error {
	path := fmt.Sprintf("/v0.1/merchants/%v/api-keys/%v", merchantCode, keyId)

	req, err := s.client.NewRequest(ctx, http.MethodDelete, path, http.NoBody)
	if err != nil {
		return fmt.Errorf("error building request: %v", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("invalid response: %d - %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	dec := json.NewDecoder(resp.Body)
	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := dec.Decode(&apiErr); err != nil {
			return fmt.Errorf("read error response: %s", err.Error())
		}

		return &apiErr
	}

	return nil
}

// GetApikey: Retrieve an API Key
// Gets an API key.
func (s *ApiKeysService) GetApikey(ctx context.Context, merchantCode string, keyId string) (*Apikey, error) {
	path := fmt.Sprintf("/v0.1/merchants/%v/api-keys/%v", merchantCode, keyId)

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

	var v Apikey
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("decode response: %s", err.Error())
	}

	return &v, nil
}

// UpdateApikey: Update an API key
// Updates an API key.
func (s *ApiKeysService) UpdateApikey(ctx context.Context, merchantCode string, keyId string, body UpdateApikeyBody) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return fmt.Errorf("encoding json body request failed: %v", err)
	}

	path := fmt.Sprintf("/v0.1/merchants/%v/api-keys/%v", merchantCode, keyId)

	req, err := s.client.NewRequest(ctx, http.MethodPut, path, buf)
	if err != nil {
		return fmt.Errorf("error building request: %v", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("invalid response: %d - %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	dec := json.NewDecoder(resp.Body)
	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := dec.Decode(&apiErr); err != nil {
			return fmt.Errorf("read error response: %s", err.Error())
		}

		return &apiErr
	}

	return nil
}
