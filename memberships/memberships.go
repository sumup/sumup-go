// Code generated by `go-sdk-gen`. DO NOT EDIT.

package memberships

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sumup/sumup-go/client"
	"github.com/sumup/sumup-go/shared"
)

// Membership: A membership associates a user with a resource, memberships is defined by user, resource, resource
// type, and associated roles.
type Membership struct {
	// Object attributes that modifiable only by SumUp applications.
	Attributes *shared.Attributes `json:"attributes,omitempty"`
	// The timestamp of when the membership was created.
	CreatedAt time.Time `json:"created_at"`
	// ID of the membership.
	Id string `json:"id"`
	// Pending invitation for membership.
	Invite *shared.Invite `json:"invite,omitempty"`
	// Set of user-defined key-value pairs attached to the object. Partial updates are not supported. When updating, always
	// submit whole metadata.
	Metadata *shared.Metadata `json:"metadata,omitempty"`
	// User's permissions.
	Permissions []string `json:"permissions"`
	// Information about the resource the membership is in.
	Resource MembershipResource `json:"resource"`
	// ID of the resource the membership is in.
	ResourceId string `json:"resource_id"`
	// User's roles.
	Roles []string `json:"roles"`
	// The status of the membership.
	Status shared.MembershipStatus `json:"status"`
	// Type of the resource the membership is in.
	Type MembershipType `json:"type"`
	// The timestamp of when the membership was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// MembershipType: Type of the resource the membership is in.
type MembershipType string

const (
	MembershipTypeMerchant MembershipType = "merchant"
)

// MembershipResource: Information about the resource the membership is in.
type MembershipResource struct {
	// Object attributes that modifiable only by SumUp applications.
	Attributes shared.Attributes `json:"attributes"`
	// The timestamp of when the membership resource was created.
	CreatedAt time.Time `json:"created_at"`
	// ID of the resource the membership is in.
	Id string `json:"id"`
	// Logo fo the resource.
	// Format: uri
	// Max length: 256
	Logo *string `json:"logo,omitempty"`
	// Display name of the resource.
	Name string                 `json:"name"`
	Type MembershipResourceType `json:"type"`
	// The timestamp of when the membership resource was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// MembershipResourceType is a schema definition.
type MembershipResourceType string

const (
	MembershipResourceTypeMerchant MembershipResourceType = "merchant"
)

// ListMembershipsParams: query parameters for ListMemberships
type ListMembershipsParams struct {
	// Filter memberships by resource kind.
	Kind *string
	// Maximum number of members to return.
	Limit *int
	// Offset of the first member to return.
	Offset *int
	// Filter memberships by the sandbox status of the resource the membership is in.
	ResourceAttributesSandbox *bool
	// Filter memberships by the name of the resource the membership is in.
	ResourceName *string
}

// QueryValues converts [ListMembershipsParams] into [url.Values].
func (p *ListMembershipsParams) QueryValues() url.Values {
	q := make(url.Values)

	if p.Kind != nil {
		q.Set("kind", *p.Kind)
	}

	if p.Limit != nil {
		q.Set("limit", strconv.Itoa(*p.Limit))
	}

	if p.Offset != nil {
		q.Set("offset", strconv.Itoa(*p.Offset))
	}

	if p.ResourceAttributesSandbox != nil {
		q.Set("resource.attributes.sandbox", strconv.FormatBool(*p.ResourceAttributesSandbox))
	}

	if p.ResourceName != nil {
		q.Set("resource.name", *p.ResourceName)
	}

	return q
}

// ListMemberships200Response is a schema definition.
type ListMemberships200Response struct {
	Items      []Membership `json:"items"`
	TotalCount int          `json:"total_count"`
}

type MembershipsService struct {
	c *client.Client
}

func NewMembershipsService(c *client.Client) *MembershipsService {
	return &MembershipsService{c: c}
}

// List: List memberships
// List memberships of the current user.
func (s *MembershipsService) List(ctx context.Context, params ListMembershipsParams) (*ListMemberships200Response, error) {
	path := fmt.Sprintf("/v0.1/memberships")

	resp, err := s.c.Call(ctx, http.MethodGet, path, client.WithQueryValues(params.QueryValues()))
	if err != nil {
		return nil, fmt.Errorf("error building request: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var v ListMemberships200Response
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			return nil, fmt.Errorf("decode response: %s", err.Error())
		}

		return &v, nil
	default:
		return nil, fmt.Errorf("unexpected response %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
}
