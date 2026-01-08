package speckit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseClarifications(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		firstQ   string
	}{
		{
			name: "multiple questions",
			input: `Here are some clarification questions:

1. What authentication method should we use?
2. Should the password have specific requirements?
3. Do we need email verification?`,
			expected: 3,
			firstQ:   "What authentication method should we use?",
		},
		{
			name:     "no questions",
			input:    "The specification is clear. No questions needed.",
			expected: 0,
		},
		{
			name: "single question",
			input: `Please clarify:

1. What is the expected response format?`,
			expected: 1,
			firstQ:   "What is the expected response format?",
		},
		{
			name:     "empty input",
			input:    "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			questions := ParseClarifications(tt.input)

			if len(questions) != tt.expected {
				t.Errorf("ParseClarifications() returned %d questions, want %d", len(questions), tt.expected)
			}

			if tt.expected > 0 && questions[0].Question != tt.firstQ {
				t.Errorf("First question = %q, want %q", questions[0].Question, tt.firstQ)
			}

			// Verify IDs are assigned
			for i, q := range questions {
				expectedID := "Q" + string(rune('1'+i))
				if q.ID != expectedID {
					t.Errorf("Question %d ID = %q, want %q", i, q.ID, expectedID)
				}
			}
		})
	}
}

func TestParseSpec(t *testing.T) {
	// Create a temp directory for test files
	tmpDir, err := os.MkdirTemp("", "speckit-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test spec.md file
	specContent := `# User Authentication Feature

## Overview
This feature adds user authentication to the application.

## User Story: Login Flow
As a user, I want to log in with my email and password.

## User Story: Password Reset
As a user, I want to reset my password if I forget it.

## Requirements
- Secure password storage
- Email validation
`

	specPath := filepath.Join(tmpDir, "spec.md")
	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		t.Fatal(err)
	}

	spec, err := ParseSpec(tmpDir)
	if err != nil {
		t.Fatalf("ParseSpec() error = %v", err)
	}

	if spec.Title != "User Authentication Feature" {
		t.Errorf("Spec.Title = %q, want %q", spec.Title, "User Authentication Feature")
	}

	if len(spec.UserStories) != 2 {
		t.Errorf("Spec.UserStories count = %d, want 2", len(spec.UserStories))
	}

	if spec.RawContent == "" {
		t.Error("Spec.RawContent should not be empty")
	}

	if spec.FilePath != specPath {
		t.Errorf("Spec.FilePath = %q, want %q", spec.FilePath, specPath)
	}
}

func TestParseSpec_NotFound(t *testing.T) {
	_, err := ParseSpec("/nonexistent/path")
	if err == nil {
		t.Error("ParseSpec() should return error for missing file")
	}
}

func TestParsePlan(t *testing.T) {
	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "speckit-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test plan.md file
	planContent := `# Implementation Plan

## Tech Stack
- Go 1.21
- PostgreSQL
- Redis for caching

## Architecture
RESTful API design with clean architecture.

## Phases
### Phase 1
Set up project structure.
`

	planPath := filepath.Join(tmpDir, "plan.md")
	if err := os.WriteFile(planPath, []byte(planContent), 0644); err != nil {
		t.Fatal(err)
	}

	plan, err := ParsePlan(tmpDir)
	if err != nil {
		t.Fatalf("ParsePlan() error = %v", err)
	}

	if len(plan.TechStack) != 3 {
		t.Errorf("Plan.TechStack count = %d, want 3", len(plan.TechStack))
	}

	if plan.RawContent == "" {
		t.Error("Plan.RawContent should not be empty")
	}
}

func TestParseTasks(t *testing.T) {
	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "speckit-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test tasks.md file
	tasksContent := `# Tasks

## User Story: Authentication

- [ ] Set up database schema for users
- [ ] [P] Implement password hashing
- [ ] Create login endpoint
- [ ] Write tests for authentication

## User Story: Authorization

- [ ] Implement role-based access
- [x] Add middleware for auth checks
`

	tasksPath := filepath.Join(tmpDir, "tasks.md")
	if err := os.WriteFile(tasksPath, []byte(tasksContent), 0644); err != nil {
		t.Fatal(err)
	}

	tasks, err := ParseTasks(tmpDir)
	if err != nil {
		t.Fatalf("ParseTasks() error = %v", err)
	}

	if len(tasks) != 6 {
		t.Errorf("ParseTasks() count = %d, want 6", len(tasks))
	}

	// Check first task
	if tasks[0].ID != "T-001" {
		t.Errorf("First task ID = %q, want %q", tasks[0].ID, "T-001")
	}
	if tasks[0].UserStoryRef != "Authentication" {
		t.Errorf("First task UserStoryRef = %q, want %q", tasks[0].UserStoryRef, "Authentication")
	}

	// Check parallel task
	var parallelFound bool
	for _, task := range tasks {
		if task.IsParallel {
			parallelFound = true
			break
		}
	}
	if !parallelFound {
		t.Error("Should find at least one parallel task marked with [P]")
	}

	// Check test task detection
	var testTaskFound bool
	for _, task := range tasks {
		if task.IsTest {
			testTaskFound = true
			break
		}
	}
	if !testTaskFound {
		t.Error("Should detect task containing 'test'")
	}
}

func TestSpecSummary(t *testing.T) {
	spec := &Spec{
		Title: "Test Feature",
		UserStories: []UserStory{
			{ID: "US-1", Title: "Story One"},
			{ID: "US-2", Title: "Story Two"},
		},
	}

	summary := spec.Summary()

	if !containsString(summary, "Test Feature") {
		t.Errorf("Summary should contain title, got %q", summary)
	}
	if !containsString(summary, "2") {
		t.Errorf("Summary should contain story count, got %q", summary)
	}
}

func TestPlanSummary(t *testing.T) {
	plan := &Plan{
		TechStack: []string{"Go", "PostgreSQL", "Redis"},
		Phases: []PlanPhase{
			{Name: "Phase 1"},
			{Name: "Phase 2"},
		},
	}

	summary := plan.Summary()

	if !containsString(summary, "Go") {
		t.Errorf("Summary should contain tech stack items, got %q", summary)
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
