package foreman

import (
	"context"
	"fmt"
	"sync"

	"github.com/bayological/foreman/internal/agents"
	"github.com/bayological/foreman/internal/git"
	"github.com/bayological/foreman/internal/telegram"
	"github.com/bayological/foreman/internal/validation"
)

type Foreman struct {
	cfg       *Config
	repo      *git.Repo
	agents    map[string]agents.Agent
	reviewer  *agents.Reviewer
	telegram  *telegram.Bot
	taskQueue chan *Task

	mu     sync.RWMutex
	active map[string]context.CancelFunc
	sem    chan struct{}
}

func New(cfg *Config) (*Foreman, error) {
	repo, err := git.NewRepo(cfg.Repo.Path, cfg.Repo.Remote, cfg.Repo.MainBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to open repo: %w", err)
	}

	tg, err := telegram.NewBot(cfg.Telegram.Token, cfg.Telegram.ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	f := &Foreman{
		cfg:       cfg,
		repo:      repo,
		telegram:  tg,
		taskQueue: make(chan *Task, 100),
		active:    make(map[string]context.CancelFunc),
		sem:       make(chan struct{}, cfg.Concurrency.MaxTasks),
		agents:    make(map[string]agents.Agent),
	}

	// Initialize agents based on config
	if cfg.Agents.ClaudeCode.Enabled {
		f.agents["claude-code"] = agents.NewClaudeCode(cfg.Repo.Path)
	}
	if cfg.Agents.Codex.Enabled {
		f.agents["codex"] = agents.NewCodex(cfg.Repo.Path)
	}

	// Initialize reviewer
	f.reviewer = agents.NewReviewer(cfg.Repo.Path, cfg.Review.UseLLM)

	return f, nil
}

func (f *Foreman) Run(ctx context.Context) error {
	f.telegram.Send("🏗️ Foreman starting up...")

	// Register command handlers
	f.registerHandlers()

	// Start Telegram listener
	go f.telegram.Listen(ctx)

	// Main work loop
	for {
		select {
		case <-ctx.Done():
			f.telegram.Send("🛑 Foreman shutting down")
			return ctx.Err()

		case task := <-f.taskQueue:
			f.sem <- struct{}{}
			go func(t *Task) {
				defer func() { <-f.sem }()
				f.processTask(ctx, t)
			}(task)
		}
	}
}

func (f *Foreman) Assign(task *Task) {
	f.taskQueue <- task
}

func (f *Foreman) processTask(ctx context.Context, task *Task) {
	taskCtx, cancel := context.WithTimeout(ctx, task.Timeout)
	defer cancel()

	f.trackTask(task.ID, cancel)
	defer f.untrackTask(task.ID)

	task.Status = StatusRunning
	f.telegram.Send(fmt.Sprintf(
		"🔨 *Task Started*\nID: `%s`\nAgent: %s\nBranch: `%s`",
		task.ID, task.AgentName, task.Branch,
	))

	// Setup worktree
	wt, err := f.repo.CreateWorktree(task.Branch)
	if err != nil {
		f.fail(task, fmt.Errorf("worktree setup failed: %w", err))
		return
	}
	defer f.repo.RemoveWorktree(task.Branch)

	task.WorktreePath = wt.Path

	// Get agent
	agent, ok := f.agents[task.AgentName]
	if !ok {
		f.fail(task, fmt.Errorf("unknown agent: %s", task.AgentName))
		return
	}

	// Build full prompt with context
	fullSpec := task.Spec
	if task.Context != "" {
		fullSpec = fmt.Sprintf("%s\n\n## Additional Context\n%s", task.Spec, task.Context)
	}

	// Execute
	result, err := agent.Execute(taskCtx, &agents.Task{
		ID:           task.ID,
		Spec:         fullSpec,
		WorktreePath: task.WorktreePath,
	})

	if err != nil {
		f.handleExecutionError(task, err)
		return
	}

	if !result.Success {
		f.handleAgentFailure(task, result)
		return
	}

	// Commit and push
	if err := f.repo.CommitAndPush(wt, fmt.Sprintf("Task %s: %s", task.ID, truncate(task.Spec, 50))); err != nil {
		f.fail(task, fmt.Errorf("git push failed: %w", err))
		return
	}

	// Review
	task.Status = StatusReview
	f.telegram.Send(fmt.Sprintf("🔍 Reviewing `%s`...", task.ID))

	review, err := f.reviewer.Review(taskCtx, &agents.ReviewRequest{
		Branch:       task.Branch,
		BaseBranch:   f.cfg.Repo.MainBranch,
		WorktreePath: wt.Path,
		Spec:         task.Spec,
	})

	if err != nil {
		f.fail(task, fmt.Errorf("review failed: %w", err))
		return
	}

	f.handleReview(task, result, review)
}

func (f *Foreman) handleReview(task *Task, result *agents.TaskResult, review *agents.ReviewResult) {
	switch review.Verdict {
	case agents.VerdictApprove:
		task.Status = StatusApproval
		f.telegram.RequestApproval(task.ID, review.Summary, task.PRURL("https://github.com/owner/repo"))

	case agents.VerdictRequestChanges:
		if task.Attempt < f.cfg.Review.MaxRetries {
			f.telegram.Send(fmt.Sprintf(
				"🔄 *Changes Requested* - Attempt %d/%d\n\n%s",
				task.Attempt+1, f.cfg.Review.MaxRetries, review.Summary,
			))
			task.Attempt++
			task.AddContext(fmt.Sprintf("Review Feedback (attempt %d):\n%s", task.Attempt, review.Summary))
			task.Status = StatusPending
			f.taskQueue <- task
		} else {
			f.escalate(task, review, "Max retries exceeded")
		}

	case agents.VerdictBlock:
		f.escalate(task, review, "Blocking issues found")
	}
}

func (f *Foreman) handleExecutionError(task *Task, err error) {
	if task.Attempt < f.cfg.Review.MaxRetries {
		f.telegram.Send(fmt.Sprintf(
			"⚠️ *Execution Error* - Retrying (%d/%d)\nError: %s",
			task.Attempt+1, f.cfg.Review.MaxRetries, validation.SanitizeErrorMessage(err),
		))
		task.Attempt++
		task.AddContext(fmt.Sprintf("Previous attempt failed with error: %v", err))
		f.taskQueue <- task
	} else {
		f.fail(task, err)
	}
}

func (f *Foreman) handleAgentFailure(task *Task, result *agents.TaskResult) {
	if task.Attempt < f.cfg.Review.MaxRetries {
		f.telegram.Send(fmt.Sprintf(
			"⚠️ *Agent Failed* - Retrying (%d/%d)\n%s",
			task.Attempt+1, f.cfg.Review.MaxRetries, result.Summary,
		))
		task.Attempt++
		task.AddContext(fmt.Sprintf("Previous attempt failed:\n%s", result.Summary))
		f.taskQueue <- task
	} else {
		f.fail(task, fmt.Errorf("agent failed: %s", result.Summary))
	}
}

func (f *Foreman) fail(task *Task, err error) {
	task.Status = StatusFailed
	f.telegram.Send(fmt.Sprintf("❌ *Task Failed*\nID: `%s`\nError: %s", task.ID, validation.SanitizeErrorMessage(err)))
}

func (f *Foreman) escalate(task *Task, review *agents.ReviewResult, reason string) {
	task.Status = StatusApproval
	f.telegram.Escalate(task.ID, reason, review.Summary)
}

func (f *Foreman) trackTask(id string, cancel context.CancelFunc) {
	f.mu.Lock()
	f.active[id] = cancel
	f.mu.Unlock()
}

func (f *Foreman) untrackTask(id string) {
	f.mu.Lock()
	delete(f.active, id)
	f.mu.Unlock()
}

func (f *Foreman) cancelTask(id string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if cancel, ok := f.active[id]; ok {
		cancel()
		return true
	}
	return false
}

func (f *Foreman) getActiveTaskIDs() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	ids := make([]string, 0, len(f.active))
	for id := range f.active {
		ids = append(ids, id)
	}
	return ids
}

func (f *Foreman) getAgentNames() []string {
	names := make([]string, 0, len(f.agents))
	for name := range f.agents {
		names = append(names, name)
	}
	return names
}

func (f *Foreman) approveTask(taskID string) error {
	return f.repo.MergeBranch(fmt.Sprintf("task/%s", taskID))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}