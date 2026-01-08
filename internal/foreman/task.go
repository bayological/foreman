package foreman

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusReview    TaskStatus = "review"
	StatusApproval  TaskStatus = "awaiting_approval"
	StatusComplete  TaskStatus = "complete"
	StatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID           string
	Spec         string
	Context      string
	Branch       string
	WorktreePath string
	AgentName    string
	Timeout      time.Duration
	Attempt      int
	Status       TaskStatus
	CreatedAt    time.Time
	FeatureID    string
	Metadata     map[string]string
}

func NewTask(spec string, agentName string, timeout time.Duration) *Task {
	id := uuid.New().String()[:8]
	return &Task{
		ID:        id,
		Spec:      spec,
		Branch:    fmt.Sprintf("task/%s", id),
		AgentName: agentName,
		Timeout:   timeout,
		Attempt:   0,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]string),
	}
}

func (t *Task) PRURL(repoURL string) string {
	return fmt.Sprintf("%s/compare/main...%s", repoURL, t.Branch)
}

func (t *Task) AddContext(ctx string) {
	if t.Context != "" {
		t.Context += "\n\n---\n"
	}
	t.Context += ctx
}