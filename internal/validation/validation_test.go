package validation

import (
	"errors"
	"testing"
)

func TestIsValidBranchName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid simple", "main", true},
		{"valid with slash", "feature/test", true},
		{"valid with hyphen", "feature-test", true},
		{"valid with underscore", "feature_test", true},
		{"valid complex", "feature/add-user-auth_v2", true},
		{"empty string", "", false},
		{"path traversal", "../etc/passwd", false},
		{"double dot in middle", "feature/../main", false},
		{"starts with dot", ".hidden", false},
		{"starts with slash", "/invalid", false},
		{"starts with hyphen", "-invalid", false},
		{"contains space", "feature test", false},
		{"too long", string(make([]byte, 201)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidBranchName(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidBranchName(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid uuid", "123e4567-e89b-12d3-a456-426614174000", true},
		{"valid uuid uppercase", "123E4567-E89B-12D3-A456-426614174000", true},
		{"empty string", "", false},
		{"too short", "123e4567-e89b-12d3-a456", false},
		{"no hyphens", "123e4567e89b12d3a456426614174000", false},
		{"invalid chars", "123e4567-e89b-12d3-a456-42661417400g", false},
		{"random string", "not-a-valid-uuid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidUUID(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidUUID(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    error
		contains string
		notContains string
	}{
		{
			name:     "nil error",
			input:    nil,
			contains: "",
		},
		{
			name:        "unix path removed",
			input:       errors.New("failed to read /home/user/secret/file.txt"),
			contains:    "[path]",
			notContains: "/home/user",
		},
		{
			name:        "mac path removed",
			input:       errors.New("failed to read /Users/admin/Documents/passwords.txt"),
			contains:    "[path]",
			notContains: "/Users/admin",
		},
		{
			name:        "tmp path removed",
			input:       errors.New("temp file error at /tmp/secret123/data"),
			contains:    "[path]",
			notContains: "/tmp/secret",
		},
		{
			name:     "generic error preserved",
			input:    errors.New("connection refused"),
			contains: "connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeErrorMessage(tt.input)
			if tt.contains != "" && !containsString(result, tt.contains) {
				t.Errorf("SanitizeErrorMessage() = %q, should contain %q", result, tt.contains)
			}
			if tt.notContains != "" && containsString(result, tt.notContains) {
				t.Errorf("SanitizeErrorMessage() = %q, should not contain %q", result, tt.notContains)
			}
		})
	}
}

func TestSanitizeErrorMessage_Truncation(t *testing.T) {
	// Create an error with a very long message
	longMsg := string(make([]byte, 600))
	for i := range longMsg {
		longMsg = longMsg[:i] + "a" + longMsg[i+1:]
	}
	err := errors.New(longMsg)

	result := SanitizeErrorMessage(err)
	if len(result) > 500 {
		t.Errorf("SanitizeErrorMessage() length = %d, want <= 500", len(result))
	}
}

func containsString(s, substr string) bool {
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
