package tools

import (
	"context"
	"os"
	"testing"
)

func TestNewCodeRabbit(t *testing.T) {
	cr := NewCodeRabbit()
	if cr == nil {
		t.Fatal("NewCodeRabbit() should not return nil")
	}
	if !cr.enabled {
		t.Error("NewCodeRabbit() should have enabled=true by default")
	}
}

func TestCodeRabbitSetEnabled(t *testing.T) {
	cr := NewCodeRabbit()

	// Test disabling
	cr.SetEnabled(false)
	if cr.enabled {
		t.Error("SetEnabled(false) should set enabled to false")
	}

	// Test enabling
	cr.SetEnabled(true)
	if !cr.enabled {
		t.Error("SetEnabled(true) should set enabled to true")
	}
}

func TestCodeRabbitIsAvailable(t *testing.T) {
	cr := NewCodeRabbit()
	// We don't know if coderabbit is installed, so just verify the method doesn't panic
	_ = cr.IsAvailable()
}

func TestCodeRabbitRun_Disabled(t *testing.T) {
	ctx := context.Background()
	tmpDir, err := os.MkdirTemp("", "coderabbit-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cr := NewCodeRabbit()
	cr.SetEnabled(false)

	output, err := cr.Run(ctx, tmpDir, "test-branch")
	if err != nil {
		t.Errorf("Run() with disabled should not error, got %v", err)
	}
	if !containsString(output, "disabled") {
		t.Errorf("Output should mention 'disabled', got %q", output)
	}
}

func TestCodeRabbitRun_NotInstalled(t *testing.T) {
	ctx := context.Background()
	tmpDir, err := os.MkdirTemp("", "coderabbit-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cr := NewCodeRabbit()
	cr.SetEnabled(true)

	// Unless coderabbit is installed, this should gracefully handle the missing CLI
	output, err := cr.Run(ctx, tmpDir, "test-branch")

	// If coderabbit isn't installed, we expect a graceful message
	// If it is installed, we expect either output or an error about the branch
	if err != nil {
		// This is acceptable - the command might fail if no git repo
		t.Logf("CodeRabbit returned error (may be expected): %v", err)
	}

	// Output should contain something
	if output == "" {
		t.Error("Output should not be empty")
	}
}

func TestCodeRabbitRun_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	tmpDir, err := os.MkdirTemp("", "coderabbit-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cr := NewCodeRabbit()
	cr.SetEnabled(true)

	// Should handle cancelled context gracefully
	_, _ = cr.Run(ctx, tmpDir, "test-branch")
	// We don't check specific behavior - just that it doesn't panic
}
