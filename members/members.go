// Code generated by `go-sdk-gen`. DO NOT EDIT.

package members

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sumup/sumup-go/client"
	"github.com/sumup/sumup-go/shared"
)

// Member: A member is user within specific resource identified by resource id, resource type, and associated roles.
type Member struct {
	// Object attributes that modifiable only by SumUp applications.
	Attributes *shared.Attributes `json:"attributes,omitempty"`
	// The timestamp of when the member was created.
	CreatedAt time.Time `json:"created_at"`
	// ID of the member.
	Id string `json:"id"`
	// Pending invitation for membership.
	Invite *shared.Invite `json:"invite,omitempty"`
	// Set of user-defined key-value pairs attached to the object. Partial updates are not supported. When updating, always
	// submit whole metadata.
	Metadata *shared.Metadata `json:"metadata,omitempty"`
	// User's permissions.
	Permissions []string `json:"permissions"`
	// User's roles.
	Roles []string `json:"roles"`
	// The status of the membership.
	Status shared.MembershipStatus `json:"status"`
	// The timestamp of when the member was last updated.
	UpdatedAt time.Time `json:"updated_at"`
	// Information about the user associated with the membership.
	User *MembershipUser `json:"user,omitempty"`
}

// MembershipUser: Information about the user associated with the membership.
type MembershipUser struct {
	// Classic identifiers of the user.
	// Deprecated: this operation is deprecated
	Classic *MembershipUserClassic `json:"classic,omitempty"`
	// Time when the user has been disabled. Applies only to virtual users (`virtual_user: true`).
	DisabledAt *time.Time `json:"disabled_at,omitempty"`
	// End-User's preferred e-mail address. Its value MUST conform to the RFC 5322 [RFC5322] addr-spec syntax. The
	// RP MUST NOT rely upon this value being unique, for unique identification use ID instead.
	Email string `json:"email"`
	// Identifier for the End-User (also called Subject).
	Id string `json:"id"`
	// True if the user has enabled MFA on login.
	MfaOnLoginEnabled bool `json:"mfa_on_login_enabled"`
	// User's preferred name. Used for display purposes only.
	Nickname *string `json:"nickname,omitempty"`
	// URL of the End-User's profile picture. This URL refers to an image file (for example, a PNG, JPEG, or GIF
	// image file), rather than to a Web page containing an image.
	// Format: uri
	Picture *string `json:"picture,omitempty"`
	// True if the user is a virtual user (operator).
	VirtualUser bool `json:"virtual_user"`
}

// MembershipUserClassic: Classic identifiers of the user.
// Deprecated: this operation is deprecated
type MembershipUserClassic struct {
	// Format: int32
	UserId int `json:"user_id"`
}

// CreateMerchantMemberBody is a schema definition.
type CreateMerchantMemberBody struct {
	// Object attributes that modifiable only by SumUp applications.
	Attributes *shared.Attributes `json:"attributes,omitempty"`
	// Email address of the member to add.
	// Format: email
	Email string `json:"email"`
	// True if the user is managed by the merchant. In this case, we'll created a virtual user with the provided password
	// and nickname.
	IsManagedUser *bool `json:"is_managed_user,omitempty"`
	// Set of user-defined key-value pairs attached to the object. Partial updates are not supported. When updating, always
	// submit whole metadata.
	Metadata *shared.Metadata `json:"metadata,omitempty"`
	// Nickname of the member to add. Only used if `is_managed_user` is true. Used for display purposes only.
	Nickname *string `json:"nickname,omitempty"`
	// Password of the member to add. Only used if `is_managed_user` is true.
	// Format: password
	// Min length: 8
	Password *string `json:"password,omitempty"`
	// List of roles to assign to the new member.
	Roles []string `json:"roles"`
}

// UpdateMerchantMemberBody is a schema definition.
type UpdateMerchantMemberBody struct {
	// Object attributes that modifiable only by SumUp applications.
	Attributes *shared.Attributes `json:"attributes,omitempty"`
	// Set of user-defined key-value pairs attached to the object. Partial updates are not supported. When updating, always
	// submit whole metadata.
	Metadata *shared.Metadata `json:"metadata,omitempty"`
	Roles    *[]string        `json:"roles,omitempty"`
	// Allows you to update user data of managed users.
	User *UpdateMerchantMemberBodyUser `json:"user,omitempty"`
}

// UpdateMerchantMemberBodyUser: Allows you to update user data of managed users.
type UpdateMerchantMemberBodyUser struct {
	// User's preferred name. Used for display purposes only.
	Nickname *string `json:"nickname,omitempty"`
	// Password of the member to add. Only used if `is_managed_user` is true.
	// Format: password
	// Min length: 8
	Password *string `json:"password,omitempty"`
}

// ListMerchantMembersParams: query parameters for ListMerchantMembers
type ListMerchantMembersParams struct {
	// Filter the returned users by email address prefix.
	Email *string
	// Maximum number of member to return.
	Limit *int
	// Offset of the first member to return.
	Offset *int
	// Filter the returned users by role.
	Roles *[]string
	// Indicates to skip count query.
	Scroll *bool
	// Filter the returned members by the membership status.
	Status *shared.MembershipStatus
}

// QueryValues converts [ListMerchantMembersParams] into [url.Values].
func (p *ListMerchantMembersParams) QueryValues() url.Values {
	q := make(url.Values)

	if p.Email != nil {
		q.Set("email", *p.Email)
	}

	if p.Limit != nil {
		q.Set("limit", strconv.Itoa(*p.Limit))
	}

	if p.Offset != nil {
		q.Set("offset", strconv.Itoa(*p.Offset))
	}

	if p.Roles != nil {
		for _, v := range *p.Roles {
			q.Add("roles", v)
		}
	}

	if p.Scroll != nil {
		q.Set("scroll", strconv.FormatBool(*p.Scroll))
	}

	if p.Status != nil {
		q.Set("status", string(*p.Status))
	}

	return q
}

// ListMerchantMembers200Response is a schema definition.
type ListMerchantMembers200Response struct {
	Items      []Member `json:"items"`
	TotalCount *int     `json:"total_count,omitempty"`
}

type MembersService struct {
	c *client.Client
}

func NewMembersService(c *client.Client) *MembersService {
	return &MembersService{c: c}
}

// List: List members
// Lists merchant members with their roles and permissions.
func (s *MembersService) List(ctx context.Context, merchantCode string, params ListMerchantMembersParams) (*ListMerchantMembers200Response, error) {
	path := fmt.Sprintf("/v0.1/merchants/%v/members", merchantCode)

	resp, err := s.c.Call(ctx, http.MethodGet, path, client.WithQueryValues(params.QueryValues()))
	if err != nil {
		return nil, fmt.Errorf("error building request: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var v ListMerchantMembers200Response
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			return nil, fmt.Errorf("decode response: %s", err.Error())
		}

		return &v, nil
	case http.StatusNotFound:
		return nil, errors.New("Merchant not found.")
	default:
		return nil, fmt.Errorf("unexpected response %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
}

// Create: Create a merchant member.
func (s *MembersService) Create(ctx context.Context, merchantCode string, body CreateMerchantMemberBody) (*Member, error) {
	path := fmt.Sprintf("/v0.1/merchants/%v/members", merchantCode)

	resp, err := s.c.Call(ctx, http.MethodPost, path, client.WithJSONBody(body))
	if err != nil {
		return nil, fmt.Errorf("error building request: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusCreated:
		var v Member
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			return nil, fmt.Errorf("decode response: %s", err.Error())
		}

		return &v, nil
	case http.StatusBadRequest:
		return nil, errors.New("Invalid request.")
	case http.StatusNotFound:
		return nil, errors.New("Merchant not found.")
	case http.StatusTooManyRequests:
		return nil, errors.New("Too many invitations sent to that user. The limit is 10 requests per 5 minutes and the Retry-After header is set to the number of minutes until the reset of the limit.")
	default:
		return nil, fmt.Errorf("unexpected response %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
}

// Delete: Delete member
// Deletes member by ID.
func (s *MembersService) Delete(ctx context.Context, merchantCode string, memberId string) error {
	path := fmt.Sprintf("/v0.1/merchants/%v/members/%v", merchantCode, memberId)

	resp, err := s.c.Call(ctx, http.MethodDelete, path)
	if err != nil {
		return fmt.Errorf("error building request: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return errors.New("Merchant or member not found.")
	default:
		return fmt.Errorf("unexpected response %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
}

// Get: Get merchant member
// Returns merchant member details.
func (s *MembersService) Get(ctx context.Context, merchantCode string, memberId string) (*Member, error) {
	path := fmt.Sprintf("/v0.1/merchants/%v/members/%v", merchantCode, memberId)

	resp, err := s.c.Call(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("error building request: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var v Member
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			return nil, fmt.Errorf("decode response: %s", err.Error())
		}

		return &v, nil
	case http.StatusNotFound:
		return nil, errors.New("Merchant or member not found.")
	default:
		return nil, fmt.Errorf("unexpected response %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
}

// Update: Update merchant member
// Update assigned roles of the member.
func (s *MembersService) Update(ctx context.Context, merchantCode string, memberId string, body UpdateMerchantMemberBody) (*Member, error) {
	path := fmt.Sprintf("/v0.1/merchants/%v/members/%v", merchantCode, memberId)

	resp, err := s.c.Call(ctx, http.MethodPut, path, client.WithJSONBody(body))
	if err != nil {
		return nil, fmt.Errorf("error building request: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var v Member
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			return nil, fmt.Errorf("decode response: %s", err.Error())
		}

		return &v, nil
	case http.StatusBadRequest:
		return nil, errors.New("Cannot set password or nickname for an invited user.")
	case http.StatusForbidden:
		return nil, errors.New("Cannot change password for managed user. Password was already used before.")
	case http.StatusNotFound:
		return nil, errors.New("Merchant or member not found.")
	case http.StatusConflict:
		return nil, errors.New("Cannot update member as some data conflict with existing members.")
	default:
		return nil, fmt.Errorf("unexpected response %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
}
