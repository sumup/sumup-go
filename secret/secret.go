package secret

import (
	"encoding/json"
	"fmt"
)

var (
	_ json.Marshaler   = (*Secret)(nil)
	_ json.Unmarshaler = (*Secret)(nil)
	_ fmt.Stringer     = (*Secret)(nil)
)

// Secret wraps sensitive string values like passwords to prevent accidental
// exposure in logs, debug output, or error messages. The underlying value
// is kept private and masked when formatted or printed.
type Secret struct {
	value string
}

// New creates a new Secret from a string value.
func New(value string) Secret {
	return Secret{value: value}
}

// Value returns the underlying secret value as a string.
func (s Secret) Value() string {
	return s.value
}

// String returns a masked representation of the secret.
// This prevents accidental exposure in logs and debug output.
func (s Secret) String() string {
	return "***"
}

// MarshalJSON serializes the secret as a plain JSON string.
// The actual value is preserved during serialization for API communication.
func (s Secret) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.value)
}

// UnmarshalJSON deserializes a JSON string into the secret.
func (s *Secret) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &s.value)
}

// GoString returns a masked representation for Go's %#v format.
// This prevents exposure when using fmt.Printf with the %#v verb.
func (s Secret) GoString() string {
	return `secret.Secret{value:"***"}`
}
