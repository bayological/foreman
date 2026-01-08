package foreman

import "testing"

func TestCanTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     Phase
		to       Phase
		expected bool
	}{
		// Valid transitions
		{"idle to specifying", PhaseIdle, PhaseSpecifying, true},
		{"specifying to awaiting spec", PhaseSpecifying, PhaseAwaitingSpecApproval, true},
		{"specifying to failed", PhaseSpecifying, PhaseFailed, true},
		{"awaiting spec to clarifying", PhaseAwaitingSpecApproval, PhaseClarifying, true},
		{"awaiting spec to specifying (re-run)", PhaseAwaitingSpecApproval, PhaseSpecifying, true},
		{"clarifying to planning", PhaseClarifying, PhasePlanning, true},
		{"planning to awaiting plan", PhasePlanning, PhaseAwaitingPlanApproval, true},
		{"awaiting plan to tasking", PhaseAwaitingPlanApproval, PhaseTasking, true},
		{"awaiting plan to planning (re-run)", PhaseAwaitingPlanApproval, PhasePlanning, true},
		{"tasking to awaiting task", PhaseTasking, PhaseAwaitingTaskApproval, true},
		{"awaiting task to implementing", PhaseAwaitingTaskApproval, PhaseImplementing, true},
		{"implementing to reviewing", PhaseImplementing, PhaseReviewing, true},
		{"reviewing to awaiting code", PhaseReviewing, PhaseAwaitingCodeApproval, true},
		{"reviewing to implementing (retry)", PhaseReviewing, PhaseImplementing, true},
		{"awaiting code to complete", PhaseAwaitingCodeApproval, PhaseComplete, true},
		{"awaiting code to implementing (more tasks)", PhaseAwaitingCodeApproval, PhaseImplementing, true},
		{"complete to idle (reset)", PhaseComplete, PhaseIdle, true},
		{"failed to idle (restart)", PhaseFailed, PhaseIdle, true},

		// Invalid transitions
		{"idle to planning (skip)", PhaseIdle, PhasePlanning, false},
		{"specifying to complete (skip)", PhaseSpecifying, PhaseComplete, false},
		{"complete to specifying", PhaseComplete, PhaseSpecifying, false},
		{"implementing to complete (skip review)", PhaseImplementing, PhaseComplete, false},
		{"unknown phase", Phase("unknown"), PhaseIdle, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanTransition(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("CanTransition(%s, %s) = %v, want %v", tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestPhaseInfo(t *testing.T) {
	tests := []struct {
		phase       Phase
		wantEmoji   string
		wantHuman   bool
	}{
		{PhaseIdle, "idle", false},
		{PhaseSpecifying, "specifying", false},
		{PhaseAwaitingSpecApproval, "awaiting", true},
		{PhaseClarifying, "clarifying", true},
		{PhasePlanning, "planning", false},
		{PhaseAwaitingPlanApproval, "awaiting", true},
		{PhaseTasking, "tasking", false},
		{PhaseAwaitingTaskApproval, "awaiting", true},
		{PhaseImplementing, "implementing", false},
		{PhaseReviewing, "reviewing", false},
		{PhaseAwaitingCodeApproval, "awaiting", true},
		{PhaseComplete, "complete", false},
		{PhaseFailed, "failed", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.phase), func(t *testing.T) {
			info := tt.phase.Info()
			if info.Emoji != tt.wantEmoji {
				t.Errorf("Phase(%s).Info().Emoji = %q, want %q", tt.phase, info.Emoji, tt.wantEmoji)
			}
			if info.NeedsHuman != tt.wantHuman {
				t.Errorf("Phase(%s).Info().NeedsHuman = %v, want %v", tt.phase, info.NeedsHuman, tt.wantHuman)
			}
		})
	}
}

func TestPhaseString(t *testing.T) {
	tests := []struct {
		phase    Phase
		contains string
	}{
		{PhaseIdle, "Idle"},
		{PhaseSpecifying, "Specifying"},
		{PhaseComplete, "Complete"},
		{PhaseFailed, "Failed"},
	}

	for _, tt := range tests {
		t.Run(string(tt.phase), func(t *testing.T) {
			result := tt.phase.String()
			if !containsSubstring(result, tt.contains) {
				t.Errorf("Phase(%s).String() = %q, should contain %q", tt.phase, result, tt.contains)
			}
		})
	}
}

func TestUnknownPhaseInfo(t *testing.T) {
	unknown := Phase("nonexistent")
	info := unknown.Info()

	if info.Emoji != "unknown" {
		t.Errorf("Unknown phase emoji = %q, want %q", info.Emoji, "unknown")
	}
	if info.Name != "nonexistent" {
		t.Errorf("Unknown phase name = %q, want %q", info.Name, "nonexistent")
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
