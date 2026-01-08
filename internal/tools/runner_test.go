package tools

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestRunCommand(t *testing.T) {
	ctx := context.Background()
	tmpDir, err := os.MkdirTemp("", "tools-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test successful command
	output, err := RunCommand(ctx, tmpDir, "echo", "hello")
	if err != nil {
		t.Errorf("RunCommand(echo hello) error = %v", err)
	}
	if output != "hello\n" && output != "hello" {
		t.Errorf("RunCommand(echo hello) output = %q, want 'hello'", output)
	}
}

func TestRunCommand_WithError(t *testing.T) {
	ctx := context.Background()
	tmpDir, err := os.MkdirTemp("", "tools-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test failing command
	_, err = RunCommand(ctx, tmpDir, "false")
	if err == nil {
		t.Error("RunCommand(false) should return error")
	}
}

func TestRunCommand_NotFound(t *testing.T) {
	ctx := context.Background()
	tmpDir, err := os.MkdirTemp("", "tools-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = RunCommand(ctx, tmpDir, "nonexistentcommand12345")
	if err == nil {
		t.Error("RunCommand with nonexistent command should return error")
	}
}

func TestRunCommand_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	tmpDir, err := os.MkdirTemp("", "tools-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = RunCommand(ctx, tmpDir, "sleep", "10")
	if err == nil {
		t.Error("RunCommand should return error on timeout")
	}
}

func TestRunCommandWithResult(t *testing.T) {
	ctx := context.Background()
	tmpDir, err := os.MkdirTemp("", "tools-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	result := RunCommandWithResult(ctx, tmpDir, "echo", "test")

	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.Err != nil {
		t.Errorf("Err = %v, want nil", result.Err)
	}
	if result.Stdout != "test\n" && result.Stdout != "test" {
		t.Errorf("Stdout = %q, want 'test'", result.Stdout)
	}
}

func TestRunCommandWithResult_Failure(t *testing.T) {
	ctx := context.Background()
	tmpDir, err := os.MkdirTemp("", "tools-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	result := RunCommandWithResult(ctx, tmpDir, "false")

	if result.ExitCode == 0 {
		t.Error("ExitCode should not be 0 for failed command")
	}
	if result.Err == nil {
		t.Error("Err should not be nil for failed command")
	}
}

func TestCommandAvailable(t *testing.T) {
	// Test with a command that should exist on all systems
	if !CommandAvailable("echo") {
		t.Error("CommandAvailable(echo) should return true")
	}

	// Test with a nonexistent command
	if CommandAvailable("nonexistentcommand12345xyz") {
		t.Error("CommandAvailable(nonexistent) should return false")
	}
}

func TestFormatCommandError(t *testing.T) {
	err := os.ErrNotExist

	// Without stderr
	msg := FormatCommandError("mycommand", []string{"arg1", "arg2"}, err, "")
	if !containsString(msg, "mycommand") {
		t.Errorf("FormatCommandError should contain command name, got %q", msg)
	}
	if !containsString(msg, "arg1") {
		t.Errorf("FormatCommandError should contain args, got %q", msg)
	}

	// With stderr
	msgWithStderr := FormatCommandError("cmd", nil, err, "error output here")
	if !containsString(msgWithStderr, "error output here") {
		t.Errorf("FormatCommandError should contain stderr, got %q", msgWithStderr)
	}
}

func TestTruncateOutput(t *testing.T) {
	// Short string - no truncation
	short := truncateOutput("hello", 10)
	if short != "hello" {
		t.Errorf("truncateOutput(hello, 10) = %q, want %q", short, "hello")
	}

	// Long string - should truncate
	long := truncateOutput("this is a very long string", 10)
	if len(long) > 10 {
		t.Errorf("truncateOutput() length = %d, should be <= 10", len(long))
	}
	if !containsString(long, "...") {
		t.Errorf("truncateOutput() should end with '...', got %q", long)
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
