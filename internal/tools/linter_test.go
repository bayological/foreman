package tools

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestNewLinter(t *testing.T) {
	// Test with no linters - should use defaults
	l := NewLinter()
	if len(l.linters) != 2 {
		t.Errorf("NewLinter() with no args should have 2 default linters, got %d", len(l.linters))
	}

	// Test with custom linters
	l2 := NewLinter("golangci-lint", "pylint")
	if len(l2.linters) != 2 {
		t.Errorf("NewLinter(golangci-lint, pylint) should have 2 linters, got %d", len(l2.linters))
	}
	if l2.linters[0] != "golangci-lint" {
		t.Errorf("First linter should be golangci-lint, got %s", l2.linters[0])
	}
}

func TestLinterConfigs(t *testing.T) {
	// Verify known linters have proper configs
	knownLinters := []string{"eslint", "ruff", "golangci-lint", "flake8", "pylint"}
	for _, name := range knownLinters {
		cfg, ok := linterConfigs[name]
		if !ok {
			t.Errorf("linterConfigs should contain %s", name)
			continue
		}
		if cfg.command == "" {
			t.Errorf("linterConfigs[%s].command should not be empty", name)
		}
		if cfg.check == "" {
			t.Errorf("linterConfigs[%s].check should not be empty", name)
		}
	}
}

func TestLinterRun_NoLintersAvailable(t *testing.T) {
	ctx := context.Background()
	tmpDir, err := os.MkdirTemp("", "linter-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Use a linter that definitely doesn't exist
	l := NewLinter("nonexistentlinter12345")
	output, err := l.Run(ctx, tmpDir)
	if err != nil {
		t.Errorf("Linter.Run() should not error for unavailable linter, got %v", err)
	}
	if !containsString(output, "not installed") {
		t.Errorf("Output should mention 'not installed', got %q", output)
	}
}

func TestLinterRun_UnknownLinter(t *testing.T) {
	ctx := context.Background()
	tmpDir, err := os.MkdirTemp("", "linter-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Unknown linter that doesn't exist
	l := NewLinter("unknownlinter999")
	output, err := l.Run(ctx, tmpDir)
	if err != nil {
		t.Errorf("Linter.Run() should not error, got %v", err)
	}
	if !containsString(output, "not installed") {
		t.Errorf("Output should mention 'not installed' for unknown linter, got %q", output)
	}
}

func TestLinterRun_WithEcho(t *testing.T) {
	ctx := context.Background()
	tmpDir, err := os.MkdirTemp("", "linter-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Use 'echo' as a mock linter - it exists on all Unix systems
	l := NewLinter("echo")
	output, err := l.Run(ctx, tmpDir)
	if err != nil {
		t.Errorf("Linter.Run() should not error, got %v", err)
	}
	// Echo with no args produces empty/newline output
	if output == "" {
		t.Error("Output should not be empty when linter runs")
	}
}

func TestLinterRun_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	tmpDir, err := os.MkdirTemp("", "linter-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// This tests context cancellation behavior
	l := NewLinter("sleep")
	_, _ = l.Run(ctx, tmpDir) // May or may not error depending on timing
}

func TestLinterRun_EmptyLinters(t *testing.T) {
	ctx := context.Background()
	tmpDir, err := os.MkdirTemp("", "linter-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create linter with empty list by directly setting it
	l := &Linter{linters: []string{}}
	output, err := l.Run(ctx, tmpDir)
	if err != nil {
		t.Errorf("Linter.Run() with empty linters should not error, got %v", err)
	}
	if !containsString(output, "No linters") {
		t.Errorf("Output should mention 'No linters', got %q", output)
	}
}
