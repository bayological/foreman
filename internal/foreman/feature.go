package foreman

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bayological/foreman/internal/speckit"
)

type Feature struct {
	ID          string
	Name        string
	Description string
	Branch      string

	Phase       Phase
	CurrentTask *Task

	Spec      *speckit.Spec
	Plan      *speckit.Plan
	Tasks     []*Task
	TaskIndex int

	PendingQuestions []speckit.Question
	Answers          map[string]string

	Events    []WorkflowEvent
	CreatedAt time.Time
	UpdatedAt time.Time

	TechStack   string
	Constraints string

	mu sync.RWMutex
}

func NewFeature(id, name, description string) *Feature {
	return &Feature{
		ID:          id,
		Name:        name,
		Description: description,
		Branch:      fmt.Sprintf("feature/%s-%s", id, sanitizeBranchName(name)),
		Phase:       PhaseIdle,
		Answers:     make(map[string]string),
		Events:      []WorkflowEvent{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (f *Feature) Transition(to Phase, message, actor string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !CanTransition(f.Phase, to) {
		return fmt.Errorf("invalid transition from %s to %s", f.Phase, to)
	}

	event := WorkflowEvent{
		Timestamp: time.Now(),
		FromPhase: f.Phase,
		ToPhase:   to,
		Message:   message,
		Actor:     actor,
	}

	f.Events = append(f.Events, event)
	f.Phase = to
	f.UpdatedAt = time.Now()

	return nil
}

func (f *Feature) GetPhase() Phase {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.Phase
}

func (f *Feature) SetSpec(spec *speckit.Spec) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Spec = spec
	f.UpdatedAt = time.Now()
}

func (f *Feature) SetPlan(plan *speckit.Plan) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Plan = plan
	f.UpdatedAt = time.Now()
}

func (f *Feature) SetTasks(tasks []*Task) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Tasks = tasks
	f.TaskIndex = 0
	f.UpdatedAt = time.Now()
}

func (f *Feature) NextTask() *Task {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.TaskIndex >= len(f.Tasks) {
		return nil
	}

	task := f.Tasks[f.TaskIndex]
	f.CurrentTask = task
	f.TaskIndex++
	return task
}

func (f *Feature) HasMoreTasks() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.TaskIndex < len(f.Tasks)
}

func (f *Feature) Progress() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if len(f.Tasks) == 0 {
		return f.Phase.String()
	}

	completed := 0
	for _, t := range f.Tasks {
		if t.Status == StatusComplete {
			completed++
		}
	}

	return fmt.Sprintf("%s (%d/%d tasks)", f.Phase.String(), completed, len(f.Tasks))
}

func (f *Feature) StatusReport() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	report := fmt.Sprintf("*Feature: %s*\n", f.Name)
	report += fmt.Sprintf("ID: `%s`\n", f.ID)
	report += fmt.Sprintf("Branch: `%s`\n", f.Branch)
	report += fmt.Sprintf("Phase: %s\n", f.Phase.String())

	if len(f.Tasks) > 0 {
		completed := 0
		for _, t := range f.Tasks {
			if t.Status == StatusComplete {
				completed++
			}
		}
		report += fmt.Sprintf("Progress: %d/%d tasks\n", completed, len(f.Tasks))
	}

	if f.CurrentTask != nil {
		report += fmt.Sprintf("Current Task: `%s`\n", f.CurrentTask.ID)
	}

	return report
}

func sanitizeBranchName(name string) string {
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		} else if r == ' ' {
			result.WriteRune('-')
		}
	}
	s := result.String()
	if len(s) > 30 {
		s = s[:30]
	}
	return s
}
