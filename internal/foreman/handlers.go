package foreman

import (
	"fmt"
	"strings"
)

func (f *Foreman) registerHandlers() {
	f.telegram.RegisterCallback("approve", f.handleApprove)
	f.telegram.RegisterCallback("reject", f.handleReject)
	f.telegram.RegisterCallback("changes", f.handleChanges)

	f.telegram.RegisterCommand("status", f.handleStatus)
	f.telegram.RegisterCommand("assign", f.handleAssign)
	f.telegram.RegisterCommand("cancel", f.handleCancel)
	f.telegram.RegisterCommand("agents", f.handleAgents)
	f.telegram.RegisterCommand("help", f.handleHelp)
}

func (f *Foreman) handleApprove(data string) {
	taskID := strings.TrimPrefix(data, "approve:")
	if err := f.approveTask(taskID); err != nil {
		f.telegram.Send(fmt.Sprintf("❌ Merge failed for `%s`: %v", taskID, err))
		return
	}
	f.telegram.Send(fmt.Sprintf("✅ Task `%s` merged successfully", taskID))
}

func (f *Foreman) handleReject(data string) {
	taskID := strings.TrimPrefix(data, "reject:")
	if f.cancelTask(taskID) {
		f.telegram.Send(fmt.Sprintf("🗑️ Task `%s` rejected and cancelled", taskID))
	} else {
		// Task not running, just clean up the branch
		f.repo.DeleteBranch(fmt.Sprintf("task/%s", taskID))
		f.telegram.Send(fmt.Sprintf("🗑️ Task `%s` rejected", taskID))
	}
}

func (f *Foreman) handleChanges(data string) {
	taskID := strings.TrimPrefix(data, "changes:")
	f.telegram.Send(fmt.Sprintf("💬 Reply with feedback for task `%s`:", taskID))
	// The next message handler would need to capture this feedback
	// For simplicity, this could be enhanced with a state machine
}

func (f *Foreman) handleStatus(args string) {
	ids := f.getActiveTaskIDs()
	if len(ids) == 0 {
		f.telegram.Send("😴 No active tasks")
		return
	}

	msg := "📊 *Active Tasks*\n\n"
	for _, id := range ids {
		msg += fmt.Sprintf("• `%s`\n", id)
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
		f.telegram.Send(fmt.Sprintf("❌ Unknown agent: %s\nAvailable: %v", agentName, f.getAgentNames()))
		return
	}

	task := NewTask(spec, agentName, f.cfg.Concurrency.TaskTimeout)
	f.Assign(task)
	f.telegram.Send(fmt.Sprintf("📝 Task created: `%s`\nSpec: %s", task.ID, truncate(spec, 100)))
}

func (f *Foreman) handleCancel(args string) {
	taskID := strings.TrimSpace(args)
	if taskID == "" {
		f.telegram.Send("Usage: /cancel <task_id>")
		return
	}

	if f.cancelTask(taskID) {
		f.telegram.Send(fmt.Sprintf("🛑 Cancelled task `%s`", taskID))
	} else {
		f.telegram.Send(fmt.Sprintf("⚠️ Task `%s` not found or not running", taskID))
	}
}

func (f *Foreman) handleAgents(args string) {
	names := f.getAgentNames()
	if len(names) == 0 {
		f.telegram.Send("⚠️ No agents configured")
		return
	}

	msg := "🤖 *Available Agents*\n\n"
	for _, name := range names {
		msg += fmt.Sprintf("• %s\n", name)
	}
	f.telegram.Send(msg)
}

func (f *Foreman) handleHelp(args string) {
	help := `🏗️ *Foreman Commands*

/status - Show active tasks
/assign <agent> <spec> - Create new task
/cancel <task_id> - Cancel a running task
/agents - List available agents
/help - Show this message

*Approval Buttons*
When a task is ready for review, you'll see:
- ✅ Approve - Merge the changes
- ❌ Reject - Discard the changes
- 🔄 Request Changes - Send feedback`

	f.telegram.Send(help)
}