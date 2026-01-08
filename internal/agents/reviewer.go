package agents

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/bayological/foreman/internal/tools"
)

type Reviewer struct {
	repoPath    string
	claude      *ClaudeCode
	coderabbit  *tools.CodeRabbit
	linter      *tools.Linter
	useLLM      bool
	testCommand string
}

type ReviewerConfig struct {
	UseLLM      bool
	TestCommand string
	Linters     []string
}

func NewReviewer(repoPath string, cfg ReviewerConfig) *Reviewer {
	testCmd := cfg.TestCommand
	if testCmd == "" {
		testCmd = "npm test" // default
	}

	return &Reviewer{
		repoPath:    repoPath,
		claude:      NewClaudeCodeReviewer(repoPath),
		coderabbit:  tools.NewCodeRabbit(),
		linter:      tools.NewLinter(cfg.Linters...),
		useLLM:      cfg.UseLLM,
		testCommand: testCmd,
	}
}

type toolResult struct {
	name   string
	output string
	err    error
}

func (r *Reviewer) Review(ctx context.Context, req *ReviewRequest) (*ReviewResult, error) {
	// Run tools concurrently
	results := make(chan toolResult, 3)
	var wg sync.WaitGroup

	// CodeRabbit
	wg.Add(1)
	go func() {
		defer wg.Done()
		out, err := r.coderabbit.Run(ctx, req.WorktreePath, req.Branch)
		results <- toolResult{"coderabbit", out, err}
	}()

	// Linter
	wg.Add(1)
	go func() {
		defer wg.Done()
		out, err := r.linter.Run(ctx, req.WorktreePath)
		results <- toolResult{"lint", out, err}
	}()

	// Tests
	wg.Add(1)
	go func() {
		defer wg.Done()
		out, err := r.runTests(ctx, req.WorktreePath)
		results <- toolResult{"tests", out, err}
	}()

	// Close channel when all done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	toolOutputs := make(map[string]string)
	for res := range results {
		if res.err != nil {
			toolOutputs[res.name] = fmt.Sprintf("ERROR: %v", res.err)
		} else {
			toolOutputs[res.name] = res.output
		}
	}

	// Get diff
	diff, _ := r.getDiff(ctx, req.WorktreePath, req.BaseBranch, req.Branch)

	// If LLM review enabled, use Claude to synthesize
	if r.useLLM {
		return r.llmReview(ctx, req, toolOutputs, diff)
	}

	// Otherwise, make decision based on tool outputs
	return r.toolBasedReview(toolOutputs), nil
}

func (r *Reviewer) llmReview(ctx context.Context, req *ReviewRequest, toolOutputs map[string]string, diff string) (*ReviewResult, error) {
	prompt := fmt.Sprintf(`You are a senior engineer reviewing a PR.

## Original Spec
%s

## Diff Summary
%s

## CodeRabbit Analysis
%s

## Linter Output
%s

## Test Results
%s

Provide a review covering:
1. Does this implementation match the spec?
2. Architectural concerns (if any)
3. Security issues (beyond what tools caught)
4. Suggestions for improvement

End with a verdict on a new line: VERDICT: APPROVE or VERDICT: REQUEST_CHANGES or VERDICT: BLOCK

Be pragmatic. Not everything needs to be perfect.
Distinguish between blocking issues and nice-to-haves.`,
		req.Spec,
		truncateString(diff, 2000),
		toolOutputs["coderabbit"],
		toolOutputs["lint"],
		toolOutputs["tests"],
	)

	output, err := r.claude.Review(ctx, prompt, req.WorktreePath)
	if err != nil {
		return nil, fmt.Errorf("LLM review failed: %w", err)
	}

	return r.parseReviewOutput(output, toolOutputs), nil
}

func (r *Reviewer) parseReviewOutput(output string, toolOutputs map[string]string) *ReviewResult {
	result := &ReviewResult{
		Summary:     output,
		ToolOutputs: toolOutputs,
		Verdict:     VerdictRequestChanges, // default to safe option
	}

	// Extract verdict from output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "VERDICT:") {
			verdict := strings.TrimSpace(strings.TrimPrefix(line, "VERDICT:"))
			switch strings.ToUpper(verdict) {
			case "APPROVE":
				result.Verdict = VerdictApprove
			case "REQUEST_CHANGES":
				result.Verdict = VerdictRequestChanges
			case "BLOCK":
				result.Verdict = VerdictBlock
			}
			break
		}
	}

	return result
}

func (r *Reviewer) toolBasedReview(toolOutputs map[string]string) *ReviewResult {
	result := &ReviewResult{
		ToolOutputs: toolOutputs,
		Verdict:     VerdictApprove,
	}

	var issues []string

	// Check for errors in tools
	for name, output := range toolOutputs {
		if strings.HasPrefix(output, "ERROR:") {
			issues = append(issues, fmt.Sprintf("%s: %s", name, output))
		}
	}

	// Check linter
	if strings.Contains(toolOutputs["lint"], "error") {
		result.Verdict = VerdictRequestChanges
		issues = append(issues, "Linter errors found")
	}

	// Check tests
	if strings.Contains(toolOutputs["tests"], "FAILED") {
		result.Verdict = VerdictBlock
		issues = append(issues, "Tests failing")
	}

	result.BlockingIssues = issues
	result.Summary = strings.Join(issues, "\n")
	if len(issues) == 0 {
		result.Summary = "All checks passed"
	}

	return result
}

func (r *Reviewer) runTests(ctx context.Context, workDir string) (string, error) {
	// Parse the test command (e.g., "npm test" -> ["npm", "test"])
	parts := strings.Fields(r.testCommand)
	if len(parts) == 0 {
		return "No test command configured", nil
	}
	return tools.RunCommand(ctx, workDir, parts[0], parts[1:]...)
}

func (r *Reviewer) getDiff(ctx context.Context, workDir, base, head string) (string, error) {
	return tools.RunCommand(ctx, workDir, "git", "diff", base+"..."+head, "--stat")
}

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}