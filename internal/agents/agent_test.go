package agents

import (
	"testing"
	"time"
)

func TestTask(t *testing.T) {
	task := &Task{
		ID:           "test-123",
		Spec:         "Implement user login",
		WorktreePath: "/tmp/worktree",
	}

	if task.ID != "test-123" {
		t.Errorf("Expected ID 'test-123', got %s", task.ID)
	}
	if task.Spec != "Implement user login" {
		t.Errorf("Expected Spec 'Implement user login', got %s", task.Spec)
	}
	if task.WorktreePath != "/tmp/worktree" {
		t.Errorf("Expected WorktreePath '/tmp/worktree', got %s", task.WorktreePath)
	}
}

func TestTaskResult(t *testing.T) {
	result := &TaskResult{
		Success:   true,
		Summary:   "Task completed successfully",
		Duration:  5 * time.Minute,
		Artifacts: []string{"file1.go", "file2.go"},
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}
	if result.Summary != "Task completed successfully" {
		t.Errorf("Expected Summary 'Task completed successfully', got %s", result.Summary)
	}
	if result.Duration != 5*time.Minute {
		t.Errorf("Expected Duration 5m, got %v", result.Duration)
	}
	if len(result.Artifacts) != 2 {
		t.Errorf("Expected 2 artifacts, got %d", len(result.Artifacts))
	}
}

func TestTaskResultFailure(t *testing.T) {
	result := &TaskResult{
		Success: false,
		Summary: "Task failed",
		Error:   errMock,
	}

	if result.Success {
		t.Error("Expected Success to be false")
	}
	if result.Error == nil {
		t.Error("Expected Error to be set")
	}
}

func TestReviewRequest(t *testing.T) {
	req := &ReviewRequest{
		Branch:       "feature/login",
		BaseBranch:   "main",
		WorktreePath: "/tmp/worktree",
		Spec:         "Implement user authentication",
	}

	if req.Branch != "feature/login" {
		t.Errorf("Expected Branch 'feature/login', got %s", req.Branch)
	}
	if req.BaseBranch != "main" {
		t.Errorf("Expected BaseBranch 'main', got %s", req.BaseBranch)
	}
	if req.WorktreePath != "/tmp/worktree" {
		t.Errorf("Expected WorktreePath '/tmp/worktree', got %s", req.WorktreePath)
	}
	if req.Spec != "Implement user authentication" {
		t.Errorf("Expected Spec 'Implement user authentication', got %s", req.Spec)
	}
}

func TestReviewVerdicts(t *testing.T) {
	tests := []struct {
		verdict  ReviewVerdict
		expected string
	}{
		{VerdictApprove, "APPROVE"},
		{VerdictRequestChanges, "REQUEST_CHANGES"},
		{VerdictBlock, "BLOCK"},
	}

	for _, tc := range tests {
		if string(tc.verdict) != tc.expected {
			t.Errorf("Expected verdict %s, got %s", tc.expected, tc.verdict)
		}
	}
}

func TestReviewResult(t *testing.T) {
	result := &ReviewResult{
		Verdict:        VerdictApprove,
		BlockingIssues: []string{},
		Suggestions:    []string{"Consider adding more tests"},
		ToolOutputs:    map[string]string{"lint": "No issues found"},
		Summary:        "Code looks good",
	}

	if result.Verdict != VerdictApprove {
		t.Errorf("Expected Verdict APPROVE, got %s", result.Verdict)
	}
	if len(result.BlockingIssues) != 0 {
		t.Errorf("Expected 0 blocking issues, got %d", len(result.BlockingIssues))
	}
	if len(result.Suggestions) != 1 {
		t.Errorf("Expected 1 suggestion, got %d", len(result.Suggestions))
	}
	if result.ToolOutputs["lint"] != "No issues found" {
		t.Errorf("Expected lint output 'No issues found', got %s", result.ToolOutputs["lint"])
	}
	if result.Summary != "Code looks good" {
		t.Errorf("Expected Summary 'Code looks good', got %s", result.Summary)
	}
}

func TestReviewResultWithBlockingIssues(t *testing.T) {
	result := &ReviewResult{
		Verdict:        VerdictBlock,
		BlockingIssues: []string{"Security vulnerability", "Missing tests"},
		Summary:        "Cannot approve due to critical issues",
	}

	if result.Verdict != VerdictBlock {
		t.Errorf("Expected Verdict BLOCK, got %s", result.Verdict)
	}
	if len(result.BlockingIssues) != 2 {
		t.Errorf("Expected 2 blocking issues, got %d", len(result.BlockingIssues))
	}
}

// Mock error for testing
type mockError string

func (e mockError) Error() string { return string(e) }

var errMock = mockError("mock error")
