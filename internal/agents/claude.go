package agents

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type ClaudeCode struct {
	repoPath string
	readOnly bool
}

func NewClaudeCode(repoPath string) *ClaudeCode {
	return &ClaudeCode{
		repoPath: repoPath,
		readOnly: false,
	}
}

func NewClaudeCodeReviewer(repoPath string) *ClaudeCode {
	return &ClaudeCode{
		repoPath: repoPath,
		readOnly: true,
	}
}

func (c *ClaudeCode) Name() string {
	if c.readOnly {
		return "claude-code-reviewer"
	}
	return "claude-code"
}

func (c *ClaudeCode) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
	start := time.Now()

	args := []string{
		"--print",
		"--output-format", "stream-json",
	}

	if c.readOnly {
		args = append(args, "--permission-mode", "read-only")
	}

	args = append(args, task.Spec)

	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = task.WorktreePath

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start claude: %w", err)
	}

	// Collect output
	var output strings.Builder
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		var msg claudeStreamMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		if msg.Type == "assistant" && msg.Content != "" {
			output.WriteString(msg.Content)
		}
	}

	// Collect stderr
	var errOutput strings.Builder
	errScanner := bufio.NewScanner(stderr)
	for errScanner.Scan() {
		errOutput.WriteString(errScanner.Text())
		errOutput.WriteString("\n")
	}

	err = cmd.Wait()
	duration := time.Since(start)

	result := &TaskResult{
		Duration: duration,
		Summary:  output.String(),
	}

	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("claude exited with error: %w\nstderr: %s", err, errOutput.String())
	} else {
		result.Success = true
	}

	return result, nil
}

// Review runs Claude Code in review mode with a specific prompt
func (c *ClaudeCode) Review(ctx context.Context, prompt string, workDir string) (string, error) {
	args := []string{
		"--print",
		"--output-format", "stream-json",
		"--permission-mode", "read-only",
		prompt,
	}

	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = workDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdout: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start claude: %w", err)
	}

	var output strings.Builder
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		var msg claudeStreamMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		if msg.Type == "assistant" && msg.Content != "" {
			output.WriteString(msg.Content)
		}
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("claude review failed: %w", err)
	}

	return output.String(), nil
}

type claudeStreamMessage struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
}