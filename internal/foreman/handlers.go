package foreman

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

func (f *Foreman) registerHandlers() {
	// Feature lifecycle commands
	f.telegram.RegisterCommand("newfeature", f.handleNewFeature)
	f.telegram.RegisterCommand("features", f.handleListFeatures)
	f.telegram.RegisterCommand("feature", f.handleFeatureStatus)

	// Phase-specific commands
	f.telegram.RegisterCommand("techstack", f.handleSetTechStack)
	f.telegram.RegisterCommand("answer", f.handleAnswer)

	// Legacy task commands (still supported)
	f.telegram.RegisterCommand("assign", f.handleAssign)
	f.telegram.RegisterCommand("cancel", f.handleCancel)

	// General commands
	f.telegram.RegisterCommand("agents", f.handleAgents)
	f.telegram.RegisterCommand("help", f.handleHelp)
	f.telegram.RegisterCommand("status", f.handleStatus)

	// Legacy approval callbacks
	f.telegram.RegisterCallback("approve", f.handleApprove)
	f.telegram.RegisterCallback("reject", f.handleReject)
	f.telegram.RegisterCallback("changes", f.handleChanges)

	// Feature workflow approval callbacks
	f.telegram.RegisterCallback("approve_spec", f.handleApproveSpec)
	f.telegram.RegisterCallback("reject_spec", f.handleRejectSpec)
	f.telegram.RegisterCallback("approve_plan", f.handleApprovePlan)
	f.telegram.RegisterCallback("reject_plan", f.handleRejectPlan)
	f.telegram.RegisterCallback("approve_tasks", f.handleApproveTasks)
	f.telegram.RegisterCallback("reject_tasks", f.handleRejectTasks)
	f.telegram.RegisterCallback("approve_code", f.handleApproveCode)
	f.telegram.RegisterCallback("reject_code", f.handleRejectCode)
	f.telegram.RegisterCallback("request_changes", f.handleRequestChanges)
	f.telegram.RegisterCallback("retry", f.handleRetry)
}

// Feature lifecycle handlers

func (f *Foreman) handleNewFeature(args string) {
	if args == "" {
		f.telegram.Send("Usage: /newfeature <name> | <description>\n\nExample:\n`/newfeature User Auth | Build user authentication with login, signup, and password reset`")
		return
	}

	parts := strings.SplitN(args, "|", 2)
	name := strings.TrimSpace(parts[0])
	description := name
	if len(parts) > 1 {
		description = strings.TrimSpace(parts[1])
	}

	if name == "" {
		f.telegram.Send("Feature name cannot be empty")
		return
	}

	ctx := context.Background()
	f.StartFeature(ctx, name, description)
}

func (f *Foreman) handleListFeatures(args string) {
	features := f.getFeatures()
	if len(features) == 0 {
		f.telegram.Send("No active features. Use /newfeature to start one.")
		return
	}

	msg := "*Active Features*\n\n"
	for _, feat := range features {
		msg += fmt.Sprintf("- `%s` %s\n  %s\n", feat.ID, feat.Name, feat.Progress())
	}
	f.telegram.Send(msg)
}

func (f *Foreman) handleFeatureStatus(args string) {
	featureID := strings.TrimSpace(args)
	if featureID == "" {
		f.telegram.Send("Usage: /feature <feature_id>")
		return
	}

	feature := f.getFeature(featureID)
	if feature == nil {
		f.telegram.Send(fmt.Sprintf("Feature `%s` not found", featureID))
		return
	}

	f.telegram.Send(feature.StatusReport())
}

func (f *Foreman) handleSetTechStack(args string) {
	parts := strings.SplitN(args, " ", 2)
	if len(parts) < 2 {
		f.telegram.Send("Usage: /techstack <feature_id> <tech stack description>")
		return
	}

	featureID := parts[0]
	techStack := parts[1]

	feature := f.getFeature(featureID)
	if feature == nil {
		f.telegram.Send(fmt.Sprintf("Feature `%s` not found", featureID))
		return
	}

	feature.TechStack = techStack
	f.telegram.Send(fmt.Sprintf("Tech stack set for `%s`: %s", featureID, techStack))
}

func (f *Foreman) handleAnswer(args string) {
	// Parse: /answer <feature_id> Q1: answer1, Q2: answer2
	parts := strings.SplitN(args, " ", 2)
	if len(parts) < 2 {
		f.telegram.Send("Usage: /answer <feature_id> Q1: answer1, Q2: answer2")
		return
	}

	featureID := parts[0]
	answersStr := parts[1]

	// Parse answers
	answerRegex := regexp.MustCompile(`(Q\d+):\s*([^,]+)`)
	matches := answerRegex.FindAllStringSubmatch(answersStr, -1)

	answers := make(map[string]string)
	for _, match := range matches {
		if len(match) >= 3 {
			answers[match[1]] = strings.TrimSpace(match[2])
		}
	}

	if len(answers) == 0 {
		f.telegram.Send("No valid answers found. Use format: Q1: answer1, Q2: answer2")
		return
	}

	ctx := context.Background()
	f.AnswerClarification(ctx, featureID, answers)
}

// Feature workflow approval callbacks

func (f *Foreman) handleApproveSpec(data string) {
	featureID := strings.TrimPrefix(data, "approve_spec:")
	ctx := context.Background()
	f.ApproveSpec(ctx, featureID)
}

func (f *Foreman) handleRejectSpec(data string) {
	featureID := strings.TrimPrefix(data, "reject_spec:")
	f.telegram.Send(fmt.Sprintf("Spec rejected for `%s`. Please provide feedback with /answer or restart with /newfeature", featureID))
}

func (f *Foreman) handleApprovePlan(data string) {
	featureID := strings.TrimPrefix(data, "approve_plan:")
	ctx := context.Background()
	f.ApprovePlan(ctx, featureID)
}

func (f *Foreman) handleRejectPlan(data string) {
	featureID := strings.TrimPrefix(data, "reject_plan:")
	f.telegram.Send(fmt.Sprintf("Plan rejected for `%s`. Please provide feedback for re-planning.", featureID))
}

func (f *Foreman) handleApproveTasks(data string) {
	featureID := strings.TrimPrefix(data, "approve_tasks:")
	ctx := context.Background()
	f.ApproveTasks(ctx, featureID)
}

func (f *Foreman) handleRejectTasks(data string) {
	featureID := strings.TrimPrefix(data, "reject_tasks:")
	f.telegram.Send(fmt.Sprintf("Tasks rejected for `%s`. Please provide feedback for re-tasking.", featureID))
}

func (f *Foreman) handleApproveCode(data string) {
	featureID := strings.TrimPrefix(data, "approve_code:")
	ctx := context.Background()
	f.approveFeatureCode(ctx, featureID)
}

func (f *Foreman) handleRejectCode(data string) {
	featureID := strings.TrimPrefix(data, "reject_code:")
	feature := f.getFeature(featureID)
	if feature != nil && feature.CurrentTask != nil {
		f.cancelTask(feature.CurrentTask.ID)
	}
	f.telegram.Send(fmt.Sprintf("Code rejected for `%s`. Task cancelled.", featureID))
}

func (f *Foreman) handleRequestChanges(data string) {
	featureID := strings.TrimPrefix(data, "request_changes:")
	f.telegram.Send(fmt.Sprintf("Please reply with specific changes needed for feature `%s`:", featureID))
}

func (f *Foreman) handleRetry(data string) {
	taskID := strings.TrimPrefix(data, "retry:")
	f.telegram.Send(fmt.Sprintf("Retrying task `%s`...", taskID))

	// Find the task and re-queue it
	f.featuresMu.RLock()
	var targetTask *Task
	for _, feature := range f.features {
		for _, task := range feature.Tasks {
			if task.ID == taskID {
				targetTask = task
				break
			}
		}
		if targetTask != nil {
			break
		}
	}
	f.featuresMu.RUnlock()

	if targetTask != nil {
		targetTask.Attempt = 0
		targetTask.Status = StatusPending
		f.taskQueue <- targetTask
	} else {
		f.telegram.Send(fmt.Sprintf("Task `%s` not found", taskID))
	}
}

// Legacy handlers (still supported for backward compatibility)

func (f *Foreman) handleApprove(data string) {
	taskID := strings.TrimPrefix(data, "approve:")
	if err := f.approveTask(taskID); err != nil {
		f.telegram.Send(fmt.Sprintf("Merge failed for `%s`: %v", taskID, err))
		return
	}
	f.telegram.Send(fmt.Sprintf("Task `%s` merged successfully", taskID))
}

func (f *Foreman) handleReject(data string) {
	taskID := strings.TrimPrefix(data, "reject:")
	if f.cancelTask(taskID) {
		f.telegram.Send(fmt.Sprintf("Task `%s` rejected and cancelled", taskID))
	} else {
		f.repo.DeleteBranch(fmt.Sprintf("task/%s", taskID))
		f.telegram.Send(fmt.Sprintf("Task `%s` rejected", taskID))
	}
}

func (f *Foreman) handleChanges(data string) {
	taskID := strings.TrimPrefix(data, "changes:")
	f.telegram.Send(fmt.Sprintf("Reply with feedback for task `%s`:", taskID))
}

func (f *Foreman) handleStatus(args string) {
	// Show both tasks and features
	ids := f.getActiveTaskIDs()
	features := f.getFeatures()

	if len(ids) == 0 && len(features) == 0 {
		f.telegram.Send("No active tasks or features")
		return
	}

	msg := "*Status*\n\n"

	if len(features) > 0 {
		msg += "*Features:*\n"
		for _, feat := range features {
			msg += fmt.Sprintf("  - `%s` %s: %s\n", feat.ID, feat.Name, feat.Progress())
		}
		msg += "\n"
	}

	if len(ids) > 0 {
		msg += "*Active Tasks:*\n"
		for _, id := range ids {
			msg += fmt.Sprintf("  - `%s`\n", id)
		}
	}

	f.telegram.Send(msg)
}

func (f *Foreman) handleAssign(args string) {
	if args == "" {
		f.telegram.Send("Usage: /assign <agent> <spec>\nExample: /assign claude-code Implement user login")
		return
	}

	parts := strings.SplitN(args, " ", 2)
	if len(parts) < 2 {
		f.telegram.Send("Usage: /assign <agent> <spec>\nExample: /assign claude-code Implement user login")
		return
	}

	agentName := parts[0]
	spec := parts[1]

	if _, ok := f.agents[agentName]; !ok {
		f.telegram.Send(fmt.Sprintf("Unknown agent: %s\nAvailable: %v", agentName, f.getAgentNames()))
		return
	}

	task := NewTask(spec, agentName, f.cfg.Concurrency.TaskTimeout)
	f.Assign(task)
	f.telegram.Send(fmt.Sprintf("Task created: `%s`\nSpec: %s", task.ID, truncate(spec, 100)))
}

func (f *Foreman) handleCancel(args string) {
	id := strings.TrimSpace(args)
	if id == "" {
		f.telegram.Send("Usage: /cancel <task_id or feature_id>")
		return
	}

	// Try cancelling as task first
	if f.cancelTask(id) {
		f.telegram.Send(fmt.Sprintf("Cancelled task `%s`", id))
		return
	}

	// Try cancelling as feature
	feature := f.getFeature(id)
	if feature != nil {
		feature.Transition(PhaseFailed, "Cancelled by user", "user")
		if feature.CurrentTask != nil {
			f.cancelTask(feature.CurrentTask.ID)
		}
		f.telegram.Send(fmt.Sprintf("Cancelled feature `%s`", id))
		return
	}

	f.telegram.Send(fmt.Sprintf("Task or feature `%s` not found or not running", id))
}

func (f *Foreman) handleAgents(args string) {
	names := f.getAgentNames()
	if len(names) == 0 {
		f.telegram.Send("No agents configured")
		return
	}

	msg := "*Available Agents*\n\n"
	for _, name := range names {
		msg += fmt.Sprintf("  - %s\n", name)
	}
	f.telegram.Send(msg)
}

func (f *Foreman) handleHelp(args string) {
	help := `*Foreman Commands*

*Feature Workflow:*
/newfeature <name> | <description> - Start new feature
/features - List all features
/feature <id> - Show feature status
/techstack <id> <stack> - Set tech stack
/answer <id> Q1: ans1, Q2: ans2 - Answer clarifications

*Legacy Commands:*
/assign <agent> <spec> - Create task directly
/cancel <id> - Cancel task or feature
/status - Show all active work
/agents - List available agents
/help - Show this message

*Workflow Phases:*
1. Specify - Create feature spec
2. Clarify - Answer questions
3. Plan - Technical planning
4. Task - Generate tasks
5. Implement - Code execution
6. Review - Automated review
7. Approve - Human approval`

	f.telegram.Send(help)
}
