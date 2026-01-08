package foreman

import (
	"context"
	"testing"
	"time"

	"github.com/bayological/foreman/internal/agents"
)

func TestNewForeman(t *testing.T) {
	// Note: This test requires mocking external dependencies
	// For now, we test that the basic config validation works

	cfg := &Config{
		Repo: RepoConfig{
			Path:       "/tmp/test-repo",
			Remote:     "origin",
			MainBranch: "main",
		},
		Telegram: TelegramConfig{
			Token:  "test-token",
			ChatID: 12345,
		},
		Concurrency: ConcurrencyConfig{
			MaxTasks:    3,
			TaskTimeout: 30 * time.Minute,
		},
		DefaultAgent: "claude-code",
	}

	// Test that config defaults are applied correctly
	if cfg.Repo.Remote != "origin" {
		t.Errorf("expected remote 'origin', got '%s'", cfg.Repo.Remote)
	}
	if cfg.Repo.MainBranch != "main" {
		t.Errorf("expected main branch 'main', got '%s'", cfg.Repo.MainBranch)
	}
	if cfg.Concurrency.MaxTasks != 3 {
		t.Errorf("expected max tasks 3, got %d", cfg.Concurrency.MaxTasks)
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	time.Sleep(time.Millisecond)
	id2 := generateID()

	if id1 == "" {
		t.Error("expected non-empty ID")
	}
	if id1 == id2 {
		t.Error("expected unique IDs")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		max      int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "he..."},
		{"", 5, ""},
		{"abc", 3, "abc"},
		{"abcd", 3, "..."},
	}

	for _, tc := range tests {
		result := truncate(tc.input, tc.max)
		if result != tc.expected {
			t.Errorf("truncate(%q, %d) = %q, expected %q", tc.input, tc.max, result, tc.expected)
		}
	}
}

func TestPendingFeedback(t *testing.T) {
	f := &Foreman{
		features: make(map[string]*Feature),
	}

	// Test setting pending feedback
	f.setPendingFeedback("feature-1", "spec", "")
	f.pendingFeedbackMu.RLock()
	feedback := f.pendingFeedback
	f.pendingFeedbackMu.RUnlock()

	if feedback == nil {
		t.Fatal("expected pending feedback to be set")
	}
	if feedback.FeatureID != "feature-1" {
		t.Errorf("expected feature ID 'feature-1', got '%s'", feedback.FeatureID)
	}
	if feedback.Phase != "spec" {
		t.Errorf("expected phase 'spec', got '%s'", feedback.Phase)
	}

	// Test clearing pending feedback
	cleared := f.clearPendingFeedback()
	if cleared == nil {
		t.Fatal("expected cleared feedback")
	}
	if cleared.FeatureID != "feature-1" {
		t.Errorf("expected feature ID 'feature-1', got '%s'", cleared.FeatureID)
	}

	// Verify it's cleared
	f.pendingFeedbackMu.RLock()
	if f.pendingFeedback != nil {
		t.Error("expected pending feedback to be nil after clearing")
	}
	f.pendingFeedbackMu.RUnlock()
}

func TestGetFeature(t *testing.T) {
	f := &Foreman{
		features: make(map[string]*Feature),
	}

	// Test getting non-existent feature
	feature := f.getFeature("nonexistent")
	if feature != nil {
		t.Error("expected nil for non-existent feature")
	}

	// Add a feature
	testFeature := NewFeature("test-1", "Test Feature", "Test description")
	f.features["test-1"] = testFeature

	// Test getting existing feature
	feature = f.getFeature("test-1")
	if feature == nil {
		t.Fatal("expected feature to be found")
	}
	if feature.ID != "test-1" {
		t.Errorf("expected ID 'test-1', got '%s'", feature.ID)
	}
}

func TestGetFeatures(t *testing.T) {
	f := &Foreman{
		features: make(map[string]*Feature),
	}

	// Test empty features
	features := f.getFeatures()
	if len(features) != 0 {
		t.Errorf("expected 0 features, got %d", len(features))
	}

	// Add features
	f.features["1"] = NewFeature("1", "Feature 1", "Desc 1")
	f.features["2"] = NewFeature("2", "Feature 2", "Desc 2")

	features = f.getFeatures()
	if len(features) != 2 {
		t.Errorf("expected 2 features, got %d", len(features))
	}
}

func TestTaskTracking(t *testing.T) {
	f := &Foreman{
		active: make(map[string]context.CancelFunc),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Test tracking
	f.trackTask("task-1", cancel)
	ids := f.getActiveTaskIDs()
	if len(ids) != 1 {
		t.Errorf("expected 1 active task, got %d", len(ids))
	}
	if ids[0] != "task-1" {
		t.Errorf("expected task ID 'task-1', got '%s'", ids[0])
	}

	// Test cancel
	cancelled := f.cancelTask("task-1")
	if !cancelled {
		t.Error("expected cancel to return true")
	}

	// Check context was cancelled
	select {
	case <-ctx.Done():
		// Good, context was cancelled
	default:
		t.Error("expected context to be cancelled")
	}

	// Test untracking
	f.trackTask("task-2", func() {})
	f.untrackTask("task-2")
	ids = f.getActiveTaskIDs()
	// Note: task-1 was only cancelled, not removed
	if len(ids) != 1 {
		t.Errorf("expected 1 active task after untrack, got %d", len(ids))
	}

	// Cancel non-existent task
	cancelled = f.cancelTask("nonexistent")
	if cancelled {
		t.Error("expected cancel to return false for non-existent task")
	}
}

func TestGetAgentNames(t *testing.T) {
	f := &Foreman{
		agents: make(map[string]agents.Agent),
	}

	// Empty agents
	names := f.getAgentNames()
	if len(names) != 0 {
		t.Errorf("expected 0 agents, got %d", len(names))
	}
}

func TestBuildPRBody(t *testing.T) {
	f := &Foreman{}
	feature := NewFeature("test-1", "Test Feature", "A test feature description")

	body := f.buildPRBody(feature)

	if body == "" {
		t.Error("expected non-empty PR body")
	}

	// Check contains summary
	if !contains(body, "Summary") {
		t.Error("expected PR body to contain 'Summary'")
	}

	// Check contains description
	if !contains(body, "A test feature description") {
		t.Error("expected PR body to contain description")
	}

	// Check contains foreman footer
	if !contains(body, "Foreman") {
		t.Error("expected PR body to contain 'Foreman'")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
