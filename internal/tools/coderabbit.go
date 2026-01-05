package tools

import (
	"context"
)

type CodeRabbit struct{}

func NewCodeRabbit() *CodeRabbit {
	return &CodeRabbit{}
}

func (c *CodeRabbit) Run(ctx context.Context, workDir string, branch string) (string, error) {
	// Adjust command based on actual CodeRabbit CLI interface
	return RunCommand(ctx, workDir, "coderabbit", "review", "--branch", branch)
}