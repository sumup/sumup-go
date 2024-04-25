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

// Meta is Set of user-defined key-value pairs attached to the object.
type Meta struct {
}

// Reader is A physical card reader device that can accept in-person payments.
type Reader struct {
	// Reader creation timestamp.
	CreatedAt time.Time `json:"created_at"`
	// Information about the underlying physical device.
	Device ReaderDevice `json:"device"`
	// Unique identifier of the object.
	//
	// Note that this identifies the instance of the physical devices pairing with your SumUp account.
	//
	// If you DELETE a reader, and pair the device again, the ID will be different. Do not use this ID to refer to a physical device.
	Id ReaderId `json:"id"`
	// Set of user-defined key-value pairs attached to the object.
	Meta *Meta `json:"meta,omitempty"`
	// Custom human-readable, user-defined name for easier identification of the reader.
	Name ReaderName `json:"name"`
	// The status of the reader object gives information about the current state of the reader.
	// Possible values:
	// * `unknown` - The reader status is unknown.
	// * `processing` - The reader is created and waits for the physical device to confirm the pairing.
	// * `paired` - The reader is paired with a merchant account and can be used with SumUp APIs.
	// * `expired` - The pairing is expired and no longer usable with the account. The ressource needs to get recreated
	Status ReaderStatus `json:"status"`
	// Reader last-modification timestamp.
	UpdatedAt time.Time `json:"updated_at"`
}

// ReaderDevice is Information about the underlying physical device.
type ReaderDevice struct {
	// A unique identifier of the physical device (e.g. serial number).
	Identifier string `json:"identifier"`
	// Identifier of the model of the device.
	Model ReaderDeviceModel `json:"model"`
}

// Identifier of the model of the device.
type ReaderDeviceModel string

const (
	ReaderDeviceModelSolo        ReaderDeviceModel = "solo"
	ReaderDeviceModelVirtualSolo ReaderDeviceModel = "virtual-solo"
)

// ReaderId is Unique identifier of the object.
//
// Note that this identifies the instance of the physical devices pairing with your SumUp account.
//
// If you DELETE a reader, and pair the device again, the ID will be different. Do not use this ID to refer to a physical device.
type ReaderId string

// ReaderName is Custom human-readable, user-defined name for easier identification of the reader.
type ReaderName string

// ReaderPairingCode is The pairing code is a 8 or 9 character alphanumeric string that is displayed on a SumUp Device after initiating the pairing.
// It is used to link the physical device to the created pairing.
type ReaderPairingCode string

// The status of the reader object gives information about the current state of the reader.
// Possible values:
// * `unknown` - The reader status is unknown.
// * `processing` - The reader is created and waits for the physical device to confirm the pairing.
// * `paired` - The reader is paired with a merchant account and can be used with SumUp APIs.
// * `expired` - The pairing is expired and no longer usable with the account. The ressource needs to get recreated
type ReaderStatus string

const (
	ReaderStatusExpired    ReaderStatus = "expired"
	ReaderStatusPaired     ReaderStatus = "paired"
	ReaderStatusProcessing ReaderStatus = "processing"
	ReaderStatusUnknown    ReaderStatus = "unknown"
)

// ListReadersResponse is the type definition for a ListReadersResponse.
type ListReadersResponse struct {
	Items []Reader `json:"items"`
}

// CreateReader request body.
type CreateReaderBody struct {
	// Set of user-defined key-value pairs attached to the object.
	Meta *Meta `json:"meta,omitempty"`
	// Custom human-readable, user-defined name for easier identification of the reader.
	Name *ReaderName `json:"name,omitempty"`
	// The pairing code is a 8 or 9 character alphanumeric string that is displayed on a SumUp Device after initiating the pairing.
	// It is used to link the physical device to the created pairing.
	PairingCode ReaderPairingCode `json:"pairing_code"`
}

// GetReaderParams are query parameters for GetReader
type GetReaderParams struct {
	IfModifiedSince *string `json:"If-Modified-Since,omitempty"`
}

// UpdateReader request body.
type UpdateReaderBody struct {
	// Set of user-defined key-value pairs attached to the object.
	Meta *Meta `json:"meta,omitempty"`
	// Custom human-readable, user-defined name for easier identification of the reader.
	Name *ReaderName `json:"name,omitempty"`
}

type ReadersService service

// ListReaders: List Readers
// Returns list of all readers of the merchant.
func (s *ReadersService) ListReaders(ctx context.Context, merchantCode string) (*ListReadersResponse, error) {
	path := fmt.Sprintf("/v0.1/merchants/%v/readers", merchantCode)

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

	var v ListReadersResponse
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("decode response: %s", err.Error())
	}

	return &v, nil
}

// CreateReader: Create a Reader
// Create a new reader linked to the merchant account.
func (s *ReadersService) CreateReader(ctx context.Context, merchantCode string, body CreateReaderBody) (*Reader, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return nil, fmt.Errorf("encoding json body request failed: %v", err)
	}

	path := fmt.Sprintf("/v0.1/merchants/%v/readers", merchantCode)

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

	var v Reader
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("decode response: %s", err.Error())
	}

	return &v, nil
}

// DeleteReader: Delete a reader
// Deletes a Reader.
func (s *ReadersService) DeleteReader(ctx context.Context, merchantCode string, id ReaderId) error {
	path := fmt.Sprintf("/v0.1/merchants/%v/readers/%v", merchantCode, id)

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

// GetReader: Retrieve a Reader
// Gets a Reader.
func (s *ReadersService) GetReader(ctx context.Context, merchantCode string, id ReaderId, params GetReaderParams) (*Reader, error) {
	path := fmt.Sprintf("/v0.1/merchants/%v/readers/%v", merchantCode, id)

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

	var v Reader
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("decode response: %s", err.Error())
	}

	return &v, nil
}

// UpdateReader: Update a Reader
// Updates a Reader.
func (s *ReadersService) UpdateReader(ctx context.Context, merchantCode string, id ReaderId, body UpdateReaderBody) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return fmt.Errorf("encoding json body request failed: %v", err)
	}

	path := fmt.Sprintf("/v0.1/merchants/%v/readers/%v", merchantCode, id)

	req, err := s.client.NewRequest(ctx, http.MethodPatch, path, buf)
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