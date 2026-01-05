package agents

import (
	"context"
	"time"
)

// Task represents a unit of work for an agent
type Task struct {
	ID           string
	Spec         string
	WorktreePath string
}

// TaskResult represents the outcome of an agent's work
type TaskResult struct {
	Success   bool
	Summary   string
	Error     error
	Duration  time.Duration
	Artifacts []string
}

// Agent defines the interface all coding agents must implement
type Agent interface {
	Name() string
	Execute(ctx context.Context, task *Task) (*TaskResult, error)
}

// ReviewRequest contains info needed for a code review
type ReviewRequest struct {
	Branch       string
	BaseBranch   string
	WorktreePath string
	Spec         string
}

// ReviewVerdict represents the outcome of a review
type ReviewVerdict string

const (
	VerdictApprove        ReviewVerdict = "APPROVE"
	VerdictRequestChanges ReviewVerdict = "REQUEST_CHANGES"
	VerdictBlock          ReviewVerdict = "BLOCK"
)

// ReviewResult contains the full review output
type ReviewResult struct {
	Verdict        ReviewVerdict
	BlockingIssues []string
	Suggestions    []string
	ToolOutputs    map[string]string
	Summary        string
}