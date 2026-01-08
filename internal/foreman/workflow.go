package foreman

import (
	"fmt"
	"time"
)

type Phase string

const (
	PhaseIdle                 Phase = "idle"
	PhaseSpecifying           Phase = "specifying"
	PhaseAwaitingSpecApproval Phase = "awaiting_spec_approval"
	PhaseClarifying           Phase = "clarifying"
	PhasePlanning             Phase = "planning"
	PhaseAwaitingPlanApproval Phase = "awaiting_plan_approval"
	PhaseTasking              Phase = "tasking"
	PhaseAwaitingTaskApproval Phase = "awaiting_task_approval"
	PhaseImplementing         Phase = "implementing"
	PhaseReviewing            Phase = "reviewing"
	PhaseAwaitingCodeApproval Phase = "awaiting_code_approval"
	PhaseComplete             Phase = "complete"
	PhaseFailed               Phase = "failed"
)

var validTransitions = map[Phase][]Phase{
	PhaseIdle:                 {PhaseSpecifying},
	PhaseSpecifying:           {PhaseAwaitingSpecApproval, PhaseFailed},
	PhaseAwaitingSpecApproval: {PhaseClarifying, PhaseSpecifying, PhaseFailed},
	PhaseClarifying:           {PhasePlanning, PhaseAwaitingSpecApproval, PhaseFailed},
	PhasePlanning:             {PhaseAwaitingPlanApproval, PhaseFailed},
	PhaseAwaitingPlanApproval: {PhaseTasking, PhasePlanning, PhaseFailed},
	PhaseTasking:              {PhaseAwaitingTaskApproval, PhaseFailed},
	PhaseAwaitingTaskApproval: {PhaseImplementing, PhaseTasking, PhaseFailed},
	PhaseImplementing:         {PhaseReviewing, PhaseFailed},
	PhaseReviewing:            {PhaseAwaitingCodeApproval, PhaseImplementing, PhaseFailed},
	PhaseAwaitingCodeApproval: {PhaseImplementing, PhaseComplete, PhaseFailed},
	PhaseComplete:             {PhaseIdle},
	PhaseFailed:               {PhaseIdle},
}

func CanTransition(from, to Phase) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, p := range allowed {
		if p == to {
			return true
		}
	}
	return false
}

type PhaseInfo struct {
	Emoji       string
	Name        string
	Description string
	NeedsHuman  bool
}

var phaseInfo = map[Phase]PhaseInfo{
	PhaseIdle:                 {"idle", "Idle", "Waiting for new feature request", false},
	PhaseSpecifying:           {"specifying", "Specifying", "Creating feature specification", false},
	PhaseAwaitingSpecApproval: {"awaiting", "Spec Review", "Waiting for spec approval", true},
	PhaseClarifying:           {"clarifying", "Clarifying", "Gathering clarifications", true},
	PhasePlanning:             {"planning", "Planning", "Creating implementation plan", false},
	PhaseAwaitingPlanApproval: {"awaiting", "Plan Review", "Waiting for plan approval", true},
	PhaseTasking:              {"tasking", "Tasking", "Breaking down into tasks", false},
	PhaseAwaitingTaskApproval: {"awaiting", "Task Review", "Waiting for task approval", true},
	PhaseImplementing:         {"implementing", "Implementing", "Coding in progress", false},
	PhaseReviewing:            {"reviewing", "Reviewing", "Code review in progress", false},
	PhaseAwaitingCodeApproval: {"awaiting", "Code Review", "Waiting for PR approval", true},
	PhaseComplete:             {"complete", "Complete", "Feature completed", false},
	PhaseFailed:               {"failed", "Failed", "Feature failed", false},
}

func (p Phase) Info() PhaseInfo {
	if info, ok := phaseInfo[p]; ok {
		return info
	}
	return PhaseInfo{"unknown", string(p), "Unknown phase", false}
}

func (p Phase) String() string {
	info := p.Info()
	return fmt.Sprintf("[%s] %s", info.Emoji, info.Name)
}

type WorkflowEvent struct {
	Timestamp time.Time
	FromPhase Phase
	ToPhase   Phase
	Message   string
	Actor     string
}
