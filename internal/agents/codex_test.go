package agents

import (
	"testing"
)

func TestNewCodex(t *testing.T) {
	agent := NewCodex("/test/repo")

	if agent.repoPath != "/test/repo" {
		t.Errorf("Expected repoPath '/test/repo', got %s", agent.repoPath)
	}
}

func TestCodexName(t *testing.T) {
	agent := NewCodex("/test/repo")

	name := agent.Name()
	if name != "codex" {
		t.Errorf("Expected name 'codex', got %s", name)
	}
}

// Note: Testing Execute method would require mocking the
// codex CLI command. This function is tested manually or via
// integration tests with the actual codex CLI installed.
//
// For proper unit testing, the command execution would need to be
// abstracted behind an interface.
