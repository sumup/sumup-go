package secret_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/sumup/sumup-go/members"
	"github.com/sumup/sumup-go/secret"
)

// TestIntegrationWithGeneratedTypes verifies that the Secret type works
// correctly with generated SDK types that use password fields.
func TestIntegrationWithGeneratedTypes(t *testing.T) {
	// Create a member creation request with a password
	pwd := secret.New("super-secret-password")
	body := members.CreateMerchantMemberBody{
		Email:    "test@example.com",
		Password: &pwd,
		Roles:    []string{"admin"},
	}

	// Marshal to JSON - password should be preserved
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Verify the password is in the JSON
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded["password"] != "super-secret-password" {
		t.Errorf("password in JSON = %v, want %q", decoded["password"], "super-secret-password")
	}

	// Verify the password is masked when printing
	str := fmt.Sprintf("%v", body.Password)
	if str != "***" {
		t.Errorf("fmt.Sprintf(%%v, password) = %q, want %q", str, "***")
	}

	// Verify the password is masked in struct string representation
	structStr := fmt.Sprintf("%+v", body)
	if contains(structStr, "super-secret-password") {
		t.Errorf("struct string representation exposed secret: %s", structStr)
	}
	if !contains(structStr, "***") {
		t.Errorf("struct string representation should contain masked password, got: %s", structStr)
	}

	// Unmarshal back
	var decodedBody members.CreateMerchantMemberBody
	if err := json.Unmarshal(data, &decodedBody); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decodedBody.Password == nil {
		t.Fatal("password should not be nil after unmarshal")
	}
	if decodedBody.Password.Value() != "super-secret-password" {
		t.Errorf("password.Value() = %q, want %q", decodedBody.Password.Value(), "super-secret-password")
	}
}

// TestUpdateMemberPasswordMasking verifies password masking in update operations.
func TestUpdateMemberPasswordMasking(t *testing.T) {
	pwd := secret.New("new-password-123")
	updateBody := members.UpdateMerchantMemberBody{
		User: &members.UpdateMerchantMemberBodyUser{
			Password: &pwd,
		},
	}

	// Verify password is masked in logs
	logMsg := fmt.Sprintf("Updating member with data: %+v", updateBody)
	if contains(logMsg, "new-password-123") {
		t.Errorf("log message exposed secret: %s", logMsg)
	}

	// Verify JSON still contains actual password
	data, err := json.Marshal(updateBody)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	user, ok := decoded["user"].(map[string]interface{})
	if !ok {
		t.Fatal("user field not found or not an object")
	}

	if user["password"] != "new-password-123" {
		t.Errorf("password in JSON = %v, want %q", user["password"], "new-password-123")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
