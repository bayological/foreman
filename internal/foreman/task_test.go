package foreman

import (
	"testing"
	"time"
)

func TestNewTask(t *testing.T) {
	task := NewTask("Implement login", "claude-code", 30*time.Minute)

	if task.ID == "" {
		t.Error("Task.ID should not be empty")
	}
	if len(task.ID) != 8 {
		t.Errorf("Task.ID length = %d, want 8", len(task.ID))
	}
	if task.Spec != "Implement login" {
		t.Errorf("Task.Spec = %q, want %q", task.Spec, "Implement login")
	}
	if task.AgentName != "claude-code" {
		t.Errorf("Task.AgentName = %q, want %q", task.AgentName, "claude-code")
	}
	if task.Timeout != 30*time.Minute {
		t.Errorf("Task.Timeout = %v, want %v", task.Timeout, 30*time.Minute)
	}
	if task.Status != StatusPending {
		t.Errorf("Task.Status = %q, want %q", task.Status, StatusPending)
	}
	if task.Attempt != 0 {
		t.Errorf("Task.Attempt = %d, want 0", task.Attempt)
	}
	if task.Metadata == nil {
		t.Error("Task.Metadata should be initialized")
	}
	if !containsSubstring(task.Branch, "task/") {
		t.Errorf("Task.Branch = %q, should contain 'task/'", task.Branch)
	}
}

func TestTaskPRURL(t *testing.T) {
	task := NewTask("Test", "agent", time.Hour)
	task.Branch = "task/abc123"

	url := task.PRURL("https://github.com/owner/repo")

	expected := "https://github.com/owner/repo/compare/main...task/abc123"
	if url != expected {
		t.Errorf("PRURL() = %q, want %q", url, expected)
	}
}

func TestTaskAddContext(t *testing.T) {
	task := NewTask("Test", "agent", time.Hour)

	// First context
	task.AddContext("First context")
	if task.Context != "First context" {
		t.Errorf("Context = %q, want %q", task.Context, "First context")
	}

	// Second context (should be appended with separator)
	task.AddContext("Second context")
	if !containsSubstring(task.Context, "First context") {
		t.Error("Context should still contain first context")
	}
	if !containsSubstring(task.Context, "Second context") {
		t.Error("Context should contain second context")
	}
	if !containsSubstring(task.Context, "---") {
		t.Error("Context should contain separator")
	}
}

func TestTaskStatuses(t *testing.T) {
	statuses := []TaskStatus{
		StatusPending,
		StatusRunning,
		StatusReview,
		StatusApproval,
		StatusComplete,
		StatusFailed,
	}

	for _, status := range statuses {
		task := NewTask("Test", "agent", time.Hour)
		task.Status = status
		if task.Status != status {
			t.Errorf("Task status assignment failed for %q", status)
		}
	}
}

func TestTaskMetadata(t *testing.T) {
	task := NewTask("Test", "agent", time.Hour)

	task.Metadata["key1"] = "value1"
	task.Metadata["key2"] = "value2"

	if task.Metadata["key1"] != "value1" {
		t.Errorf("Metadata[key1] = %q, want %q", task.Metadata["key1"], "value1")
	}
	if task.Metadata["key2"] != "value2" {
		t.Errorf("Metadata[key2] = %q, want %q", task.Metadata["key2"], "value2")
	}
}

func TestTaskUniqueIDs(t *testing.T) {
	ids := make(map[string]bool)

	// Create multiple tasks and verify unique IDs
	for i := 0; i < 100; i++ {
		task := NewTask("Test", "agent", time.Hour)
		if ids[task.ID] {
			t.Errorf("Duplicate task ID generated: %s", task.ID)
		}
		ids[task.ID] = true
	}
}
