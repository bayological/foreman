package foreman

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bayological/foreman/internal/agents"
	"github.com/bayological/foreman/internal/git"
	"github.com/bayological/foreman/internal/speckit"
	"github.com/bayological/foreman/internal/telegram"
	"github.com/bayological/foreman/internal/validation"
)

type Foreman struct {
	cfg      *Config
	repo     *git.Repo
	agents   map[string]agents.Agent
	reviewer *agents.Reviewer
	telegram *telegram.Bot
	speckit  *speckit.SpecKit

	taskQueue chan *Task

	features   map[string]*Feature
	featuresMu sync.RWMutex

	active map[string]context.CancelFunc
	mu     sync.RWMutex
	sem    chan struct{}

	// Pending feedback tracking
	pendingFeedback   *PendingFeedback
	pendingFeedbackMu sync.RWMutex
}

// PendingFeedback tracks when we're waiting for feedback text from the user
type PendingFeedback struct {
	FeatureID string
	Phase     string // "spec", "plan", "tasks", "code"
	TaskID    string // only for code phase
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
		speckit:   speckit.New(cfg.Repo.Path),
		taskQueue: make(chan *Task, 100),
		features:  make(map[string]*Feature),
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
	f.reviewer = agents.NewReviewer(cfg.Repo.Path, agents.ReviewerConfig{
		UseLLM:      cfg.Review.UseLLM,
		TestCommand: cfg.Review.Tools.TestCommand,
		Linters:     cfg.Review.Tools.Linters,
	})

	return f, nil
}

func (f *Foreman) Run(ctx context.Context) error {
	f.telegram.Send("Foreman starting up...")

	if err := f.speckit.Initialize(ctx); err != nil {
		log.Printf("Warning: SpecKit init failed: %v", err)
	}

	// Register command handlers
	f.registerHandlers()

	// Start Telegram listener
	go f.telegram.Listen(ctx)
	go f.taskProcessor(ctx)

	f.telegram.Send("Ready! Use /newfeature to start a new feature.")

	<-ctx.Done()
	f.telegram.Send("Foreman shutting down")
	return ctx.Err()
}

func (f *Foreman) taskProcessor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-f.taskQueue:
			f.sem <- struct{}{}
			go func(t *Task) {
				defer func() { <-f.sem }()
				f.executeTask(ctx, t)
			}(task)
		}
	}
}

func (f *Foreman) Assign(task *Task) {
	f.taskQueue <- task
}

// Feature workflow methods

func (f *Foreman) StartFeature(ctx context.Context, name, description string) (*Feature, error) {
	id := generateID()
	feature := NewFeature(id, name, description)

	f.featuresMu.Lock()
	f.features[id] = feature
	f.featuresMu.Unlock()

	f.telegram.Send(fmt.Sprintf(
		"*New Feature Started*\n\nID: `%s`\nName: %s\nBranch: `%s`",
		id, name, feature.Branch,
	))

	go f.runSpecificationPhase(ctx, feature)

	return feature, nil
}

func (f *Foreman) runSpecificationPhase(ctx context.Context, feature *Feature) {
	feature.Transition(PhaseSpecifying, "Starting specification", "foreman")
	f.telegram.Send(fmt.Sprintf("Creating specification for `%s`...", feature.ID))

	result, err := f.speckit.Specify(ctx, feature.Description, feature.Branch)
	if err != nil {
		f.handlePhaseError(feature, err)
		return
	}

	if !result.Success {
		f.handlePhaseError(feature, fmt.Errorf("speckit.specify failed: %s", result.Error))
		return
	}

	featureDir := f.speckit.GetLatestFeatureDir()
	spec, err := speckit.ParseSpec(featureDir)
	if err != nil {
		log.Printf("Warning: Could not parse spec: %v", err)
	} else {
		feature.SetSpec(spec)
	}

	feature.Transition(PhaseAwaitingSpecApproval, "Spec created, awaiting approval", "foreman")
	f.requestSpecApproval(feature)
}

func (f *Foreman) requestSpecApproval(feature *Feature) {
	var summary string
	if feature.Spec != nil {
		summary = feature.Spec.Summary()
	} else {
		summary = "Specification created. Please review in your IDE."
	}

	f.telegram.RequestPhaseApproval(
		feature.ID,
		"spec",
		summary,
		fmt.Sprintf("Branch: `%s`", feature.Branch),
	)
}

func (f *Foreman) ApproveSpec(ctx context.Context, featureID string) {
	feature := f.getFeature(featureID)
	if feature == nil {
		f.telegram.Send(fmt.Sprintf("Feature `%s` not found", featureID))
		return
	}

	if feature.GetPhase() != PhaseAwaitingSpecApproval {
		f.telegram.Send(fmt.Sprintf("Feature `%s` is not awaiting spec approval", featureID))
		return
	}

	f.telegram.Send(fmt.Sprintf("Spec approved for `%s`. Starting clarification...", featureID))

	go f.runClarificationPhase(ctx, feature)
}

func (f *Foreman) runClarificationPhase(ctx context.Context, feature *Feature) {
	feature.Transition(PhaseClarifying, "Running clarification", "foreman")

	result, err := f.speckit.Clarify(ctx)
	if err != nil {
		f.handlePhaseError(feature, err)
		return
	}

	questions := speckit.ParseClarifications(result.Output)
	feature.PendingQuestions = questions

	if len(questions) > 0 {
		f.sendClarificationQuestions(feature)
	} else {
		f.telegram.Send(fmt.Sprintf("No clarifications needed for `%s`. Proceeding to planning...", feature.ID))
		go f.runPlanningPhase(ctx, feature)
	}
}

func (f *Foreman) sendClarificationQuestions(feature *Feature) {
	msg := fmt.Sprintf("*Clarification Needed* for `%s`\n\n", feature.ID)

	for _, q := range feature.PendingQuestions {
		msg += fmt.Sprintf("*%s:* %s\n\n", q.ID, q.Question)
	}

	msg += "Reply with answers in format:\n`/answer " + feature.ID + " Q1: Your answer, Q2: Your answer`"

	f.telegram.Send(msg)
}

func (f *Foreman) AnswerClarification(ctx context.Context, featureID string, answers map[string]string) {
	feature := f.getFeature(featureID)
	if feature == nil {
		return
	}

	for k, v := range answers {
		feature.Answers[k] = v
	}

	allAnswered := true
	for _, q := range feature.PendingQuestions {
		if _, ok := feature.Answers[q.ID]; !ok {
			allAnswered = false
			break
		}
	}

	if allAnswered {
		f.telegram.Send(fmt.Sprintf("All clarifications answered for `%s`. Proceeding to planning...", featureID))
		go f.runPlanningPhase(ctx, feature)
	}
}

func (f *Foreman) runPlanningPhase(ctx context.Context, feature *Feature) {
	feature.Transition(PhasePlanning, "Creating implementation plan", "foreman")
	f.telegram.Send(fmt.Sprintf("Creating implementation plan for `%s`...", feature.ID))

	techStack := feature.TechStack
	if techStack == "" {
		techStack = f.cfg.DefaultTechStack
	}

	result, err := f.speckit.Plan(ctx, techStack)
	if err != nil {
		f.handlePhaseError(feature, err)
		return
	}

	if !result.Success {
		f.handlePhaseError(feature, fmt.Errorf("speckit.plan failed: %s", result.Error))
		return
	}

	featureDir := f.speckit.GetLatestFeatureDir()
	plan, err := speckit.ParsePlan(featureDir)
	if err != nil {
		log.Printf("Warning: Could not parse plan: %v", err)
	} else {
		feature.SetPlan(plan)
	}

	feature.Transition(PhaseAwaitingPlanApproval, "Plan created, awaiting approval", "foreman")
	f.requestPlanApproval(feature)
}

func (f *Foreman) requestPlanApproval(feature *Feature) {
	var summary string
	if feature.Plan != nil {
		summary = feature.Plan.Summary()
	} else {
		summary = "Implementation plan created. Please review in your IDE."
	}

	f.telegram.RequestPhaseApproval(
		feature.ID,
		"plan",
		summary,
		fmt.Sprintf("Branch: `%s`", feature.Branch),
	)
}

func (f *Foreman) ApprovePlan(ctx context.Context, featureID string) {
	feature := f.getFeature(featureID)
	if feature == nil {
		f.telegram.Send(fmt.Sprintf("Feature `%s` not found", featureID))
		return
	}

	if feature.GetPhase() != PhaseAwaitingPlanApproval {
		f.telegram.Send(fmt.Sprintf("Feature `%s` is not awaiting plan approval", featureID))
		return
	}

	f.telegram.Send(fmt.Sprintf("Plan approved for `%s`. Generating tasks...", featureID))

	go f.runTaskingPhase(ctx, feature)
}

func (f *Foreman) runTaskingPhase(ctx context.Context, feature *Feature) {
	feature.Transition(PhaseTasking, "Generating tasks", "foreman")

	result, err := f.speckit.Tasks(ctx)
	if err != nil {
		f.handlePhaseError(feature, err)
		return
	}

	if !result.Success {
		f.handlePhaseError(feature, fmt.Errorf("speckit.tasks failed: %s", result.Error))
		return
	}

	featureDir := f.speckit.GetLatestFeatureDir()
	taskItems, err := speckit.ParseTasks(featureDir)
	if err != nil {
		f.handlePhaseError(feature, err)
		return
	}

	var tasks []*Task
	for _, item := range taskItems {
		task := NewTask(item.Title, f.cfg.DefaultAgent, f.cfg.Concurrency.TaskTimeout)
		task.ID = item.ID
		task.Spec = item.Title
		task.FeatureID = feature.ID
		task.Branch = feature.Branch
		task.IsParallel = item.IsParallel
		task.Metadata["user_story"] = item.UserStoryRef
		task.Metadata["is_test"] = fmt.Sprintf("%v", item.IsTest)
		tasks = append(tasks, task)
	}

	feature.SetTasks(tasks)

	feature.Transition(PhaseAwaitingTaskApproval, "Tasks generated, awaiting approval", "foreman")
	f.requestTaskApproval(feature, taskItems)
}

func (f *Foreman) requestTaskApproval(feature *Feature, taskItems []speckit.TaskItem) {
	summary := fmt.Sprintf("*%d Tasks Generated*\n\n", len(taskItems))

	storyTasks := make(map[string][]speckit.TaskItem)
	for _, t := range taskItems {
		storyTasks[t.UserStoryRef] = append(storyTasks[t.UserStoryRef], t)
	}

	for story, tasks := range storyTasks {
		if story != "" {
			summary += fmt.Sprintf("*%s:*\n", story)
		}
		for _, t := range tasks {
			parallel := ""
			if t.IsParallel {
				parallel = " [P]"
			}
			title := t.Title
			if len(title) > 40 {
				title = title[:37] + "..."
			}
			summary += fmt.Sprintf("  - `%s` %s%s\n", t.ID, title, parallel)
		}
		summary += "\n"
	}

	f.telegram.RequestPhaseApproval(
		feature.ID,
		"tasks",
		summary,
		"",
	)
}

func (f *Foreman) ApproveTasks(ctx context.Context, featureID string) {
	feature := f.getFeature(featureID)
	if feature == nil {
		f.telegram.Send(fmt.Sprintf("Feature `%s` not found", featureID))
		return
	}

	if feature.GetPhase() != PhaseAwaitingTaskApproval {
		f.telegram.Send(fmt.Sprintf("Feature `%s` is not awaiting task approval", featureID))
		return
	}

	f.telegram.Send(fmt.Sprintf("Tasks approved for `%s`. Starting implementation...", featureID))

	go f.runImplementationPhase(ctx, feature)
}

func (f *Foreman) runImplementationPhase(ctx context.Context, feature *Feature) {
	feature.Transition(PhaseImplementing, "Starting implementation", "foreman")

	f.telegram.Send(fmt.Sprintf(
		"*Implementation Started*\n\nFeature: `%s`\nTasks: %d\n\nProgress updates will follow...",
		feature.ID, len(feature.Tasks),
	))

	// Queue all parallel tasks first
	for _, task := range feature.Tasks {
		if task.IsParallel {
			task.Status = StatusPending
			f.taskQueue <- task
		}
	}

	// Queue the first sequential task (non-parallel)
	// Subsequent sequential tasks will be queued one at a time via approveFeatureCode
	for _, task := range feature.Tasks {
		if !task.IsParallel && task.Status == StatusPending {
			feature.CurrentTask = task
			f.taskQueue <- task
			break
		}
	}
}

func (f *Foreman) completeFeature(feature *Feature) {
	feature.Transition(PhaseComplete, "All tasks completed", "foreman")

	f.telegram.Send(fmt.Sprintf(
		"*Feature Complete!*\n\nFeature: `%s`\nName: %s\nBranch: `%s`\n\nReady for final review and merge.",
		feature.ID, feature.Name, feature.Branch,
	))
}

func (f *Foreman) handlePhaseError(feature *Feature, err error) {
	feature.Transition(PhaseFailed, err.Error(), "foreman")
	f.telegram.Send(fmt.Sprintf("*Phase Failed*\n\nFeature: `%s`\nError: %v", feature.ID, err))
}

func (f *Foreman) getFeature(id string) *Feature {
	f.featuresMu.RLock()
	defer f.featuresMu.RUnlock()
	return f.features[id]
}

func (f *Foreman) getFeatures() []*Feature {
	f.featuresMu.RLock()
	defer f.featuresMu.RUnlock()
	features := make([]*Feature, 0, len(f.features))
	for _, feat := range f.features {
		features = append(features, feat)
	}
	return features
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()%100000)
}

// Task execution methods

func (f *Foreman) executeTask(ctx context.Context, task *Task) {
	taskCtx, cancel := context.WithTimeout(ctx, task.Timeout)
	defer cancel()

	f.trackTask(task.ID, cancel)
	defer f.untrackTask(task.ID)

	task.Status = StatusRunning
	f.telegram.Send(fmt.Sprintf(
		"*Task Started*\nID: `%s`\nAgent: %s\nBranch: `%s`",
		task.ID, task.AgentName, task.Branch,
	))

	// Setup worktree
	wt, err := f.repo.CreateWorktree(task.Branch)
	if err != nil {
		f.failTask(task, fmt.Errorf("worktree setup failed: %w", err))
		return
	}
	defer f.repo.RemoveWorktree(task.Branch)

	task.WorktreePath = wt.Path

	// Get agent
	agent, ok := f.agents[task.AgentName]
	if !ok {
		f.failTask(task, fmt.Errorf("unknown agent: %s", task.AgentName))
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
		f.failTask(task, fmt.Errorf("git push failed: %w", err))
		return
	}

	// Review
	task.Status = StatusReview
	f.telegram.Send(fmt.Sprintf("Reviewing `%s`...", task.ID))

	review, err := f.reviewer.Review(taskCtx, &agents.ReviewRequest{
		Branch:       task.Branch,
		BaseBranch:   f.cfg.Repo.MainBranch,
		WorktreePath: wt.Path,
		Spec:         task.Spec,
	})

	if err != nil {
		f.failTask(task, fmt.Errorf("review failed: %w", err))
		return
	}

	f.handleReview(task, result, review)
}

func (f *Foreman) handleReview(task *Task, result *agents.TaskResult, review *agents.ReviewResult) {
	switch review.Verdict {
	case agents.VerdictApprove:
		task.Status = StatusApproval

		// Check if this task belongs to a feature
		if task.FeatureID != "" {
			f.telegram.RequestPhaseApproval(task.FeatureID, "code", review.Summary, fmt.Sprintf("Task: `%s`", task.ID))
		} else {
			f.telegram.RequestApproval(task.ID, review.Summary, task.PRURL("https://github.com/owner/repo"))
		}

	case agents.VerdictRequestChanges:
		if task.Attempt < f.cfg.Review.MaxRetries {
			f.telegram.Send(fmt.Sprintf(
				"*Changes Requested* - Attempt %d/%d\n\n%s",
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
			"*Execution Error* - Retrying (%d/%d)\nError: %s",
			task.Attempt+1, f.cfg.Review.MaxRetries, validation.SanitizeErrorMessage(err),
		))
		task.Attempt++
		task.AddContext(fmt.Sprintf("Previous attempt failed with error: %v", err))
		f.taskQueue <- task
	} else {
		f.failTask(task, err)
	}
}

func (f *Foreman) handleAgentFailure(task *Task, result *agents.TaskResult) {
	if task.Attempt < f.cfg.Review.MaxRetries {
		f.telegram.Send(fmt.Sprintf(
			"*Agent Failed* - Retrying (%d/%d)\n%s",
			task.Attempt+1, f.cfg.Review.MaxRetries, result.Summary,
		))
		task.Attempt++
		task.AddContext(fmt.Sprintf("Previous attempt failed:\n%s", result.Summary))
		f.taskQueue <- task
	} else {
		f.failTask(task, fmt.Errorf("agent failed: %s", result.Summary))
	}
}

func (f *Foreman) failTask(task *Task, err error) {
	task.Status = StatusFailed
	f.telegram.Send(fmt.Sprintf("*Task Failed*\nID: `%s`\nError: %s", task.ID, validation.SanitizeErrorMessage(err)))
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

func (f *Foreman) approveFeatureCode(ctx context.Context, featureID string) {
	feature := f.getFeature(featureID)
	if feature == nil {
		f.telegram.Send(fmt.Sprintf("Feature `%s` not found", featureID))
		return
	}

	// Mark current task as complete
	if feature.CurrentTask != nil {
		feature.CurrentTask.Status = StatusComplete
	}

	// Check if all tasks are complete
	allComplete := true
	for _, task := range feature.Tasks {
		if task.Status != StatusComplete {
			allComplete = false
			break
		}
	}

	if allComplete {
		f.completeFeature(feature)
		return
	}

	// Find the next sequential (non-parallel) task that isn't complete
	for _, task := range feature.Tasks {
		if !task.IsParallel && task.Status == StatusPending {
			feature.CurrentTask = task
			f.telegram.Send(fmt.Sprintf("Starting next task `%s` for feature `%s`...", task.ID, featureID))
			f.taskQueue <- task
			return
		}
	}

	// If we get here, there are still parallel tasks running
	f.telegram.Send(fmt.Sprintf("Waiting for remaining parallel tasks in feature `%s`...", featureID))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
