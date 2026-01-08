package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CommandResult provides structured output from a command
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

// RunCommand executes a command and returns its output
func RunCommand(ctx context.Context, workDir string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()

	// Prefer stdout, fallback to stderr if stdout is empty
	if output == "" && stderr.Len() > 0 {
		output = stderr.String()
	}

	// If command failed, include stderr in the output for context
	if err != nil && stderr.Len() > 0 && stdout.Len() > 0 {
		output = output + "\n\nStderr:\n" + stderr.String()
	}

	return output, err
}

// RunCommandWithResult provides detailed command execution results
func RunCommandWithResult(ctx context.Context, workDir string, name string, args ...string) *CommandResult {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &CommandResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}

	// Extract exit code if available
	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	} else if err == nil {
		result.ExitCode = 0
	} else {
		result.ExitCode = -1
	}

	return result
}

// CommandAvailable checks if a command is available in PATH
func CommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// FormatCommandError creates a user-friendly error message
func FormatCommandError(name string, args []string, err error, stderr string) string {
	cmdStr := name
	if len(args) > 0 {
		cmdStr += " " + strings.Join(args, " ")
	}

	msg := fmt.Sprintf("Command failed: %s\nError: %v", cmdStr, err)
	if stderr != "" {
		msg += "\nOutput: " + truncateOutput(stderr, 500)
	}
	return msg
}

func truncateOutput(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}