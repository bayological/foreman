package agents

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Codex struct {
	repoPath string
}

func NewCodex(repoPath string) *Codex {
	return &Codex{
		repoPath: repoPath,
	}
}

func (c *Codex) Name() string {
	return "codex"
}

func (c *Codex) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
	start := time.Now()

	args := []string{
		"--prompt", task.Spec,
		"--quiet",
	}

	cmd := exec.CommandContext(ctx, "codex", args...)
	cmd.Dir = task.WorktreePath

	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	result := &TaskResult{
		Duration: duration,
		Summary:  strings.TrimSpace(string(output)),
	}

	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("codex failed: %w\noutput: %s", err, output)
	} else {
		result.Success = true
	}

	return result, nil
}