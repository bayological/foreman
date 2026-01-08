package foreman

import (
	"testing"
)

func TestParseNewFeatureArgs(t *testing.T) {
	tests := []struct {
		input       string
		wantName    string
		wantDesc    string
		wantErr     bool
	}{
		{"User Auth | Build authentication", "User Auth", "Build authentication", false},
		{"Simple Feature", "Simple Feature", "Simple Feature", false},
		{"", "", "", true},
		{"Name | Description | Extra", "Name", "Description | Extra", false},
		{"  Trimmed  |  Also Trimmed  ", "Trimmed", "Also Trimmed", false},
	}

	for _, tc := range tests {
		name, desc, err := parseNewFeatureArgs(tc.input)

		if tc.wantErr {
			if err == nil {
				t.Errorf("parseNewFeatureArgs(%q) expected error, got nil", tc.input)
			}
			continue
		}

		if err != nil {
			t.Errorf("parseNewFeatureArgs(%q) unexpected error: %v", tc.input, err)
			continue
		}

		if name != tc.wantName {
			t.Errorf("parseNewFeatureArgs(%q) name = %q, want %q", tc.input, name, tc.wantName)
		}
		if desc != tc.wantDesc {
			t.Errorf("parseNewFeatureArgs(%q) desc = %q, want %q", tc.input, desc, tc.wantDesc)
		}
	}
}

func TestParseAnswerArgs(t *testing.T) {
	tests := []struct {
		input       string
		wantID      string
		wantAnswers map[string]string
		wantErr     bool
	}{
		{
			"feat-1 Q1: answer one, Q2: answer two",
			"feat-1",
			map[string]string{"Q1": "answer one", "Q2": "answer two"},
			false,
		},
		{
			"feat-123 Q1: single answer",
			"feat-123",
			map[string]string{"Q1": "single answer"},
			false,
		},
		{
			"",
			"",
			nil,
			true,
		},
		{
			"feat-1",
			"",
			nil,
			true,
		},
	}

	for _, tc := range tests {
		id, answers, err := parseAnswerArgs(tc.input)

		if tc.wantErr {
			if err == nil {
				t.Errorf("parseAnswerArgs(%q) expected error, got nil", tc.input)
			}
			continue
		}

		if err != nil {
			t.Errorf("parseAnswerArgs(%q) unexpected error: %v", tc.input, err)
			continue
		}

		if id != tc.wantID {
			t.Errorf("parseAnswerArgs(%q) id = %q, want %q", tc.input, id, tc.wantID)
		}

		if len(answers) != len(tc.wantAnswers) {
			t.Errorf("parseAnswerArgs(%q) got %d answers, want %d", tc.input, len(answers), len(tc.wantAnswers))
			continue
		}

		for k, v := range tc.wantAnswers {
			if answers[k] != v {
				t.Errorf("parseAnswerArgs(%q) answers[%q] = %q, want %q", tc.input, k, answers[k], v)
			}
		}
	}
}

func TestParseTechStackArgs(t *testing.T) {
	tests := []struct {
		input     string
		wantID    string
		wantStack string
		wantErr   bool
	}{
		{"feat-1 React, TypeScript, Node.js", "feat-1", "React, TypeScript, Node.js", false},
		{"feat-123 Go", "feat-123", "Go", false},
		{"", "", "", true},
		{"feat-1", "", "", true},
	}

	for _, tc := range tests {
		id, stack, err := parseTechStackArgs(tc.input)

		if tc.wantErr {
			if err == nil {
				t.Errorf("parseTechStackArgs(%q) expected error, got nil", tc.input)
			}
			continue
		}

		if err != nil {
			t.Errorf("parseTechStackArgs(%q) unexpected error: %v", tc.input, err)
			continue
		}

		if id != tc.wantID {
			t.Errorf("parseTechStackArgs(%q) id = %q, want %q", tc.input, id, tc.wantID)
		}
		if stack != tc.wantStack {
			t.Errorf("parseTechStackArgs(%q) stack = %q, want %q", tc.input, stack, tc.wantStack)
		}
	}
}

// Helper functions to parse command arguments
func parseNewFeatureArgs(args string) (name, desc string, err error) {
	if args == "" {
		return "", "", errEmptyArgs
	}

	parts := splitOnPipe(args)
	name = trimSpace(parts[0])
	desc = name
	if len(parts) > 1 {
		desc = trimSpace(parts[1])
		// Handle extra pipe-separated content
		for i := 2; i < len(parts); i++ {
			desc += " | " + trimSpace(parts[i])
		}
	}

	if name == "" {
		return "", "", errEmptyArgs
	}

	return name, desc, nil
}

func parseAnswerArgs(args string) (featureID string, answers map[string]string, err error) {
	if args == "" {
		return "", nil, errEmptyArgs
	}

	parts := splitN(args, " ", 2)
	if len(parts) < 2 {
		return "", nil, errInvalidFormat
	}

	featureID = parts[0]
	answersStr := parts[1]

	answers = parseAnswerString(answersStr)
	if len(answers) == 0 {
		return "", nil, errInvalidFormat
	}

	return featureID, answers, nil
}

func parseTechStackArgs(args string) (featureID string, techStack string, err error) {
	if args == "" {
		return "", "", errEmptyArgs
	}

	parts := splitN(args, " ", 2)
	if len(parts) < 2 {
		return "", "", errInvalidFormat
	}

	return parts[0], parts[1], nil
}

// Simple error types for testing
type parseError string

func (e parseError) Error() string { return string(e) }

var (
	errEmptyArgs     = parseError("empty arguments")
	errInvalidFormat = parseError("invalid format")
)

// Helper functions
func splitOnPipe(s string) []string {
	var parts []string
	var current string
	for _, r := range s {
		if r == '|' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	parts = append(parts, current)
	return parts
}

func splitN(s string, sep string, n int) []string {
	var parts []string
	for i := 0; i < n-1 && len(s) > 0; i++ {
		idx := indexOf(s, sep)
		if idx < 0 {
			break
		}
		parts = append(parts, s[:idx])
		s = s[idx+len(sep):]
	}
	if len(s) > 0 {
		parts = append(parts, s)
	}
	return parts
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func parseAnswerString(s string) map[string]string {
	answers := make(map[string]string)

	// Simple regex-like parsing for Q#: answer format
	parts := splitOnComma(s)
	for _, part := range parts {
		part = trimSpace(part)
		if len(part) < 4 { // At least "Q1:x"
			continue
		}

		// Look for Q# pattern
		if part[0] != 'Q' {
			continue
		}

		colonIdx := indexOf(part, ":")
		if colonIdx < 0 {
			continue
		}

		qid := trimSpace(part[:colonIdx])
		answer := trimSpace(part[colonIdx+1:])

		if qid != "" && answer != "" {
			answers[qid] = answer
		}
	}

	return answers
}

func splitOnComma(s string) []string {
	var parts []string
	var current string
	for _, r := range s {
		if r == ',' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	parts = append(parts, current)
	return parts
}
