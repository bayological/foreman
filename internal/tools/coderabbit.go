package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type CodeRabbit struct {
	available bool
	enabled   bool
}

func NewCodeRabbit() *CodeRabbit {
	return &CodeRabbit{
		enabled: true,
	}
}

// SetEnabled allows disabling CodeRabbit reviews
func (c *CodeRabbit) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// IsAvailable checks if the CodeRabbit CLI is installed
func (c *CodeRabbit) IsAvailable() bool {
	_, err := exec.LookPath("coderabbit")
	c.available = err == nil
	return c.available
}

func (c *CodeRabbit) Run(ctx context.Context, workDir string, branch string) (string, error) {
	if !c.enabled {
		return "CodeRabbit review disabled", nil
	}

	if !c.IsAvailable() {
		return "CodeRabbit CLI not installed (skipped)", nil
	}

	// Run CodeRabbit review
	// The CLI supports different modes - we use the PR review mode
	output, err := RunCommand(ctx, workDir, "coderabbit", "review", "--branch", branch)
	if err != nil {
		// Check if it's a command not found vs actual error
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "executable file") {
			return "CodeRabbit CLI not available", nil
		}
		return fmt.Sprintf("CodeRabbit error: %v\n%s", err, output), err
	}

	if output == "" {
		return "No issues found by CodeRabbit", nil
	}

	return output, nil
}