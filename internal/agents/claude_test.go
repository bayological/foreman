package agents

import (
	"testing"
)

func TestNewClaudeCode(t *testing.T) {
	agent := NewClaudeCode("/test/repo")

	if agent.repoPath != "/test/repo" {
		t.Errorf("Expected repoPath '/test/repo', got %s", agent.repoPath)
	}
	if agent.readOnly {
		t.Error("Expected readOnly to be false for regular agent")
	}
}

func TestNewClaudeCodeReviewer(t *testing.T) {
	agent := NewClaudeCodeReviewer("/test/repo")

	if agent.repoPath != "/test/repo" {
		t.Errorf("Expected repoPath '/test/repo', got %s", agent.repoPath)
	}
	if !agent.readOnly {
		t.Error("Expected readOnly to be true for reviewer agent")
	}
}

func TestClaudeCodeName(t *testing.T) {
	tests := []struct {
		readOnly bool
		expected string
	}{
		{false, "claude-code"},
		{true, "claude-code-reviewer"},
	}

	for _, tc := range tests {
		agent := &ClaudeCode{
			repoPath: "/test",
			readOnly: tc.readOnly,
		}

		name := agent.Name()
		if name != tc.expected {
			t.Errorf("Name() with readOnly=%v: expected %s, got %s", tc.readOnly, tc.expected, name)
		}
	}
}

func TestClaudeStreamMessage(t *testing.T) {
	msg := claudeStreamMessage{
		Type:    "assistant",
		Content: "Hello, world!",
	}

	if msg.Type != "assistant" {
		t.Errorf("Expected Type 'assistant', got %s", msg.Type)
	}
	if msg.Content != "Hello, world!" {
		t.Errorf("Expected Content 'Hello, world!', got %s", msg.Content)
	}
}

// Note: Testing Execute and Review methods would require mocking the
// claude CLI command. These functions are tested manually or via
// integration tests with the actual claude CLI installed.
//
// For proper unit testing, the command execution would need to be
// abstracted behind an interface.
