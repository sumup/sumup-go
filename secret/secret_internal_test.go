package secret_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/sumup/sumup-go/secret"
)

func TestNew(t *testing.T) {
	s := secret.New("super-secret")
	if s.Value() != "super-secret" {
		t.Errorf("expected 'super-secret', got %q", s.Value())
	}
}

func TestValue(t *testing.T) {
	s := secret.New("my-password")
	if got := s.Value(); got != "my-password" {
		t.Errorf("Value() = %q, want %q", got, "my-password")
	}
}

func TestString(t *testing.T) {
	s := secret.New("super-secret")
	got := s.String()
	if got != "***" {
		t.Errorf("String() = %q, want %q", got, "***")
	}
	if got == "super-secret" {
		t.Error("String() exposed the secret value")
	}
}

func TestGoString(t *testing.T) {
	s := secret.New("super-secret")
	got := s.GoString()
	expected := `secret.Secret{value:"***"}`
	if got != expected {
		t.Errorf("GoString() = %q, want %q", got, expected)
	}
	if got == "super-secret" {
		t.Error("GoString() exposed the secret value")
	}
}

func TestFormat(t *testing.T) {
	s := secret.New("super-secret")

	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{"percent s", "%s", "***"},
		{"percent v", "%v", "***"},
		{"percent+v", "%+v", "***"},
		{"percent#v", "%#v", `secret.Secret{value:"***"}`},
		{"percent q", "%q", `"***"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fmt.Sprintf(tt.format, s)
			if got != tt.expected {
				t.Errorf("fmt.Sprintf(%q, secret) = %q, want %q", tt.format, got, tt.expected)
			}
			if got == "super-secret" || got == `"super-secret"` {
				t.Errorf("fmt.Sprintf(%q, secret) exposed the secret value", tt.format)
			}
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	s := secret.New("super-secret")
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	expected := `"super-secret"`
	if string(data) != expected {
		t.Errorf("json.Marshal() = %q, want %q", string(data), expected)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	data := []byte(`"super-secret"`)
	var s secret.Secret
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if s.Value() != "super-secret" {
		t.Errorf("after unmarshal, Value() = %q, want %q", s.Value(), "super-secret")
	}
}

func TestJSONRoundTrip(t *testing.T) {
	original := secret.New("super-secret")

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded secret.Secret
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Value() != original.Value() {
		t.Errorf("round trip: got %q, want %q", decoded.Value(), original.Value())
	}
}

func TestJSONInStruct(t *testing.T) {
	type testStruct struct {
		Username string        `json:"username"`
		Password secret.Secret `json:"password"`
	}

	// Marshal
	input := testStruct{
		Username: "user@example.com",
		Password: secret.New("my-password"),
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	expectedJSON := `{"username":"user@example.com","password":"my-password"}`
	if string(data) != expectedJSON {
		t.Errorf("json.Marshal() = %q, want %q", string(data), expectedJSON)
	}

	// Unmarshal
	var output testStruct
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if output.Username != input.Username {
		t.Errorf("username: got %q, want %q", output.Username, input.Username)
	}
	if output.Password.Value() != input.Password.Value() {
		t.Errorf("password: got %q, want %q", output.Password.Value(), input.Password.Value())
	}
}

func TestEmptySecret(t *testing.T) {
	s := secret.New("")
	if s.Value() != "" {
		t.Errorf("Value() = %q, want empty string", s.Value())
	}
	if s.String() != "***" {
		t.Errorf("String() = %q, want %q", s.String(), "***")
	}
}

func TestSecretNotExposedInErrorMessages(t *testing.T) {
	s := secret.New("super-secret")
	err := fmt.Errorf("authentication failed with password: %v", s)
	errStr := err.Error()

	if errStr == "authentication failed with password: super-secret" {
		t.Error("error message exposed the secret value")
	}
	expected := "authentication failed with password: ***"
	if errStr != expected {
		t.Errorf("error message = %q, want %q", errStr, expected)
	}
}

func TestSecretNotExposedInLogs(t *testing.T) {
	s := secret.New("super-secret")
	logMsg := fmt.Sprintf("User logged in with password %s", s)

	if logMsg == "User logged in with password super-secret" {
		t.Error("log message exposed the secret value")
	}
	expected := "User logged in with password ***"
	if logMsg != expected {
		t.Errorf("log message = %q, want %q", logMsg, expected)
	}
}

func TestSecretInPointer(t *testing.T) {
	type testStruct struct {
		Password *secret.Secret `json:"password,omitempty"`
	}

	// Test with non-nil pointer
	s := secret.New("my-password")
	input := testStruct{Password: &s}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var output testStruct
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if output.Password == nil {
		t.Fatal("Password is nil after unmarshal")
	}
	if output.Password.Value() != "my-password" {
		t.Errorf("password: got %q, want %q", output.Password.Value(), "my-password")
	}

	// Test with nil pointer
	input2 := testStruct{Password: nil}
	data2, err := json.Marshal(input2)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	expected := `{}`
	if string(data2) != expected {
		t.Errorf("json.Marshal() with nil = %q, want %q", string(data2), expected)
	}
}
