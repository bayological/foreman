package agents

import (
	"strings"
	"testing"
)

func TestNewReviewer(t *testing.T) {
	repoPath := "/test/repo"
	cfg := ReviewerConfig{
		UseLLM:        true,
		UseCodeRabbit: false,
		TestCommand:   "go test ./...",
		Linters:       []string{"golangci-lint"},
	}

	r := NewReviewer(repoPath, cfg)
	if r == nil {
		t.Fatal("NewReviewer() should not return nil")
	}
	if r.repoPath != repoPath {
		t.Errorf("NewReviewer() repoPath = %q, want %q", r.repoPath, repoPath)
	}
	if !r.useLLM {
		t.Error("NewReviewer() should have useLLM=true based on config")
	}
	if r.testCommand != "go test ./..." {
		t.Errorf("NewReviewer() testCommand = %q, want %q", r.testCommand, "go test ./...")
	}
}

func TestNewReviewer_DefaultTestCommand(t *testing.T) {
	cfg := ReviewerConfig{
		TestCommand: "", // Empty should use default
	}

	r := NewReviewer("/test/repo", cfg)
	if r.testCommand != "npm test" {
		t.Errorf("NewReviewer() with empty TestCommand should use 'npm test', got %q", r.testCommand)
	}
}

func TestReviewerConfig(t *testing.T) {
	cfg := ReviewerConfig{
		UseLLM:        false,
		UseCodeRabbit: true,
		TestCommand:   "pytest",
		Linters:       []string{"flake8", "pylint"},
	}

	if cfg.UseLLM != false {
		t.Error("ReviewerConfig.UseLLM should be false")
	}
	if cfg.UseCodeRabbit != true {
		t.Error("ReviewerConfig.UseCodeRabbit should be true")
	}
	if len(cfg.Linters) != 2 {
		t.Errorf("ReviewerConfig.Linters should have 2 items, got %d", len(cfg.Linters))
	}
}

func TestToolBasedReview_AllPassing(t *testing.T) {
	r := &Reviewer{useLLM: false}

	toolOutputs := map[string]string{
		"lint":       "no issues found",
		"tests":      "PASSED",
		"coderabbit": "No suggestions",
	}

	result := r.toolBasedReview(toolOutputs)

	if result.Verdict != VerdictApprove {
		t.Errorf("toolBasedReview() Verdict = %v, want %v", result.Verdict, VerdictApprove)
	}
	if result.Summary != "All checks passed" {
		t.Errorf("toolBasedReview() Summary = %q, want 'All checks passed'", result.Summary)
	}
}

func TestToolBasedReview_LintErrors(t *testing.T) {
	r := &Reviewer{useLLM: false}

	toolOutputs := map[string]string{
		"lint":       "error: unused variable",
		"tests":      "PASSED",
		"coderabbit": "No suggestions",
	}

	result := r.toolBasedReview(toolOutputs)

	if result.Verdict != VerdictRequestChanges {
		t.Errorf("toolBasedReview() with lint errors Verdict = %v, want %v", result.Verdict, VerdictRequestChanges)
	}
}

func TestToolBasedReview_TestsFailing(t *testing.T) {
	r := &Reviewer{useLLM: false}

	toolOutputs := map[string]string{
		"lint":  "no issues",
		"tests": "FAILED: 3 tests failed",
	}

	result := r.toolBasedReview(toolOutputs)

	if result.Verdict != VerdictBlock {
		t.Errorf("toolBasedReview() with failing tests Verdict = %v, want %v", result.Verdict, VerdictBlock)
	}
}

func TestToolBasedReview_ToolErrors(t *testing.T) {
	r := &Reviewer{useLLM: false}

	toolOutputs := map[string]string{
		"lint":       "ERROR: lint command failed",
		"tests":      "PASSED",
		"coderabbit": "ERROR: coderabbit unavailable",
	}

	result := r.toolBasedReview(toolOutputs)

	// Should still process and report issues
	if len(result.BlockingIssues) == 0 {
		t.Error("toolBasedReview() should have blocking issues for tool errors")
	}
}

func TestParseReviewOutput_Approve(t *testing.T) {
	r := &Reviewer{}
	output := `The code looks good. All tests pass and the implementation follows the spec.

VERDICT: APPROVE`

	result := r.parseReviewOutput(output, nil)

	if result.Verdict != VerdictApprove {
		t.Errorf("parseReviewOutput() Verdict = %v, want %v", result.Verdict, VerdictApprove)
	}
}

func TestParseReviewOutput_RequestChanges(t *testing.T) {
	r := &Reviewer{}
	output := `There are some issues that need to be addressed.

VERDICT: REQUEST_CHANGES`

	result := r.parseReviewOutput(output, nil)

	if result.Verdict != VerdictRequestChanges {
		t.Errorf("parseReviewOutput() Verdict = %v, want %v", result.Verdict, VerdictRequestChanges)
	}
}

func TestParseReviewOutput_Block(t *testing.T) {
	r := &Reviewer{}
	output := `Critical security issue found. This must be fixed.

VERDICT: BLOCK`

	result := r.parseReviewOutput(output, nil)

	if result.Verdict != VerdictBlock {
		t.Errorf("parseReviewOutput() Verdict = %v, want %v", result.Verdict, VerdictBlock)
	}
}

func TestParseReviewOutput_NoVerdict(t *testing.T) {
	r := &Reviewer{}
	output := `Some review comments without a verdict.`

	result := r.parseReviewOutput(output, nil)

	// Default to REQUEST_CHANGES when no verdict
	if result.Verdict != VerdictRequestChanges {
		t.Errorf("parseReviewOutput() with no verdict Verdict = %v, want %v", result.Verdict, VerdictRequestChanges)
	}
}

func TestParseReviewOutput_CaseInsensitive(t *testing.T) {
	r := &Reviewer{}

	tests := []struct {
		output string
		want   ReviewVerdict
	}{
		{"VERDICT: approve", VerdictApprove},
		{"VERDICT: APPROVE", VerdictApprove},
		{"VERDICT: Approve", VerdictApprove},
		{"VERDICT: request_changes", VerdictRequestChanges},
		{"VERDICT: block", VerdictBlock},
	}

	for _, tc := range tests {
		result := r.parseReviewOutput(tc.output, nil)
		if result.Verdict != tc.want {
			t.Errorf("parseReviewOutput(%q) Verdict = %v, want %v", tc.output, result.Verdict, tc.want)
		}
	}
}

func TestParseReviewOutput_PreservesToolOutputs(t *testing.T) {
	r := &Reviewer{}
	toolOutputs := map[string]string{
		"lint":  "no issues",
		"tests": "PASSED",
	}

	result := r.parseReviewOutput("VERDICT: APPROVE", toolOutputs)

	if result.ToolOutputs == nil {
		t.Fatal("parseReviewOutput() should preserve tool outputs")
	}
	if result.ToolOutputs["lint"] != "no issues" {
		t.Errorf("parseReviewOutput() ToolOutputs[lint] = %q, want 'no issues'", result.ToolOutputs["lint"])
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "he..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "..."},
		{"", 5, ""},
	}

	for _, tc := range tests {
		got := truncateString(tc.input, tc.max)
		if got != tc.want {
			t.Errorf("truncateString(%q, %d) = %q, want %q", tc.input, tc.max, got, tc.want)
		}
		if len(got) > tc.max {
			t.Errorf("truncateString(%q, %d) length = %d, should be <= %d", tc.input, tc.max, len(got), tc.max)
		}
	}
}

func TestReviewVerdict_Constants(t *testing.T) {
	// Verify verdict constants are as expected
	if VerdictApprove != "APPROVE" {
		t.Errorf("VerdictApprove = %q, want 'APPROVE'", VerdictApprove)
	}
	if VerdictRequestChanges != "REQUEST_CHANGES" {
		t.Errorf("VerdictRequestChanges = %q, want 'REQUEST_CHANGES'", VerdictRequestChanges)
	}
	if VerdictBlock != "BLOCK" {
		t.Errorf("VerdictBlock = %q, want 'BLOCK'", VerdictBlock)
	}
}

func TestReviewResult_Structure(t *testing.T) {
	result := &ReviewResult{
		Verdict:        VerdictApprove,
		BlockingIssues: []string{"issue1", "issue2"},
		Suggestions:    []string{"suggestion1"},
		ToolOutputs:    map[string]string{"lint": "ok"},
		Summary:        "Test summary",
	}

	if result.Verdict != VerdictApprove {
		t.Errorf("ReviewResult.Verdict = %v, want %v", result.Verdict, VerdictApprove)
	}
	if len(result.BlockingIssues) != 2 {
		t.Errorf("ReviewResult.BlockingIssues length = %d, want 2", len(result.BlockingIssues))
	}
	if len(result.Suggestions) != 1 {
		t.Errorf("ReviewResult.Suggestions length = %d, want 1", len(result.Suggestions))
	}
	if result.Summary != "Test summary" {
		t.Errorf("ReviewResult.Summary = %q, want 'Test summary'", result.Summary)
	}
}

func TestToolBasedReview_ToolOutputsPreserved(t *testing.T) {
	r := &Reviewer{useLLM: false}

	toolOutputs := map[string]string{
		"lint":       "no issues",
		"tests":      "PASSED",
		"coderabbit": "looks good",
	}

	result := r.toolBasedReview(toolOutputs)

	if len(result.ToolOutputs) != 3 {
		t.Errorf("toolBasedReview() should preserve all tool outputs, got %d", len(result.ToolOutputs))
	}
	for key := range toolOutputs {
		if _, ok := result.ToolOutputs[key]; !ok {
			t.Errorf("toolBasedReview() missing tool output for %q", key)
		}
	}
}

func TestToolBasedReview_MultipleLintErrors(t *testing.T) {
	r := &Reviewer{useLLM: false}

	toolOutputs := map[string]string{
		"lint":  "error: line 1\nerror: line 2",
		"tests": "PASSED",
	}

	result := r.toolBasedReview(toolOutputs)

	if result.Verdict != VerdictRequestChanges {
		t.Errorf("toolBasedReview() with multiple lint errors Verdict = %v, want %v", result.Verdict, VerdictRequestChanges)
	}
	if !strings.Contains(result.Summary, "Linter errors") {
		t.Errorf("toolBasedReview() Summary should mention linter errors, got %q", result.Summary)
	}
}
