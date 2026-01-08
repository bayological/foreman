package foreman

import (
	"testing"
)

func TestNewFeature(t *testing.T) {
	f := NewFeature("123", "User Auth", "Add user authentication")

	if f.ID != "123" {
		t.Errorf("Feature.ID = %q, want %q", f.ID, "123")
	}
	if f.Name != "User Auth" {
		t.Errorf("Feature.Name = %q, want %q", f.Name, "User Auth")
	}
	if f.Description != "Add user authentication" {
		t.Errorf("Feature.Description = %q, want %q", f.Description, "Add user authentication")
	}
	if f.Phase != PhaseIdle {
		t.Errorf("Feature.Phase = %q, want %q", f.Phase, PhaseIdle)
	}
	if f.Answers == nil {
		t.Error("Feature.Answers should be initialized")
	}
	if !containsSubstring(f.Branch, "feature/123") {
		t.Errorf("Feature.Branch = %q, should contain 'feature/123'", f.Branch)
	}
}

func TestFeatureBranchNameSanitization(t *testing.T) {
	tests := []struct {
		name         string
		featureName  string
		shouldContain string
		shouldNotContain string
	}{
		{
			name:         "spaces replaced with hyphens",
			featureName:  "User Auth System",
			shouldContain: "User-Auth-System",
		},
		{
			name:         "special chars removed",
			featureName:  "Feature!@#$%",
			shouldContain: "Feature",
			shouldNotContain: "!",
		},
		{
			name:         "long name truncated",
			featureName:  "This Is A Very Long Feature Name That Should Be Truncated",
			shouldContain: "feature/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFeature("1", tt.featureName, "description")
			if tt.shouldContain != "" && !containsSubstring(f.Branch, tt.shouldContain) {
				t.Errorf("Branch = %q, should contain %q", f.Branch, tt.shouldContain)
			}
			if tt.shouldNotContain != "" && containsSubstring(f.Branch, tt.shouldNotContain) {
				t.Errorf("Branch = %q, should not contain %q", f.Branch, tt.shouldNotContain)
			}
		})
	}
}

func TestFeatureTransition(t *testing.T) {
	f := NewFeature("1", "Test", "Test feature")

	// Valid transition
	err := f.Transition(PhaseSpecifying, "Starting spec", "foreman")
	if err != nil {
		t.Errorf("Valid transition returned error: %v", err)
	}
	if f.Phase != PhaseSpecifying {
		t.Errorf("Phase = %q, want %q", f.Phase, PhaseSpecifying)
	}
	if len(f.Events) != 1 {
		t.Errorf("Events length = %d, want 1", len(f.Events))
	}

	// Invalid transition
	err = f.Transition(PhaseComplete, "Skip to complete", "foreman")
	if err == nil {
		t.Error("Invalid transition should return error")
	}
	if f.Phase != PhaseSpecifying {
		t.Errorf("Phase should remain %q after invalid transition", PhaseSpecifying)
	}
}

func TestFeatureGetPhase(t *testing.T) {
	f := NewFeature("1", "Test", "Test")
	f.Phase = PhasePlanning

	phase := f.GetPhase()
	if phase != PhasePlanning {
		t.Errorf("GetPhase() = %q, want %q", phase, PhasePlanning)
	}
}

func TestFeatureProgress(t *testing.T) {
	f := NewFeature("1", "Test", "Test")

	// No tasks
	progress := f.Progress()
	if !containsSubstring(progress, "Idle") {
		t.Errorf("Progress without tasks = %q, should contain 'Idle'", progress)
	}

	// With tasks
	f.Tasks = []*Task{
		{ID: "1", Status: StatusComplete},
		{ID: "2", Status: StatusRunning},
		{ID: "3", Status: StatusPending},
	}
	progress = f.Progress()
	if !containsSubstring(progress, "1/3") {
		t.Errorf("Progress with tasks = %q, should contain '1/3'", progress)
	}
}

func TestFeatureStatusReport(t *testing.T) {
	f := NewFeature("123", "Auth", "User authentication")
	f.Phase = PhaseImplementing
	f.Tasks = []*Task{
		{ID: "T1", Status: StatusComplete},
		{ID: "T2", Status: StatusRunning},
	}
	f.CurrentTask = f.Tasks[1]

	report := f.StatusReport()

	checks := []string{"Auth", "123", "feature/", "1/2", "T2"}
	for _, check := range checks {
		if !containsSubstring(report, check) {
			t.Errorf("StatusReport() = %q, should contain %q", report, check)
		}
	}
}

func TestFeatureNextTask(t *testing.T) {
	f := NewFeature("1", "Test", "Test")
	f.Tasks = []*Task{
		{ID: "T1"},
		{ID: "T2"},
		{ID: "T3"},
	}

	// First call
	task := f.NextTask()
	if task == nil || task.ID != "T1" {
		t.Errorf("First NextTask() should return T1, got %v", task)
	}
	if f.CurrentTask != task {
		t.Error("CurrentTask should be set to returned task")
	}

	// Second call
	task = f.NextTask()
	if task == nil || task.ID != "T2" {
		t.Errorf("Second NextTask() should return T2, got %v", task)
	}

	// Third call
	task = f.NextTask()
	if task == nil || task.ID != "T3" {
		t.Errorf("Third NextTask() should return T3, got %v", task)
	}

	// Fourth call (no more tasks)
	task = f.NextTask()
	if task != nil {
		t.Errorf("NextTask() after all tasks should return nil, got %v", task)
	}
}

func TestFeatureHasMoreTasks(t *testing.T) {
	f := NewFeature("1", "Test", "Test")
	f.Tasks = []*Task{{ID: "T1"}, {ID: "T2"}}

	if !f.HasMoreTasks() {
		t.Error("HasMoreTasks() should return true initially")
	}

	f.NextTask()
	if !f.HasMoreTasks() {
		t.Error("HasMoreTasks() should return true after first task")
	}

	f.NextTask()
	if f.HasMoreTasks() {
		t.Error("HasMoreTasks() should return false after all tasks")
	}
}

func TestSanitizeBranchName(t *testing.T) {
	tests := []struct {
		input    string
		contains string
	}{
		{"simple", "simple"},
		{"With Spaces", "With-Spaces"},
		{"special!@#chars", "specialchars"},
		{"MixedCase123", "MixedCase123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeBranchName(tt.input)
			if !containsSubstring(result, tt.contains) {
				t.Errorf("sanitizeBranchName(%q) = %q, should contain %q", tt.input, result, tt.contains)
			}
		})
	}
}
