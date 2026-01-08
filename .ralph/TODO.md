# Foreman SpecKit Integration Progress

## Implementation Checklist

### Phase 1: SpecKit Package
- [x] Create `internal/speckit/` directory
- [x] Create `internal/speckit/commands.go`
- [x] Create `internal/speckit/speckit.go`
- [x] Create `internal/speckit/parser.go`
- [x] Verify: `go build ./...`

### Phase 2: Workflow Types
- [x] Create `internal/foreman/workflow.go`
- [x] Create `internal/foreman/feature.go`
- [x] Verify: `go build ./...`

### Phase 3: Modifications
- [x] Update `internal/foreman/task.go` - add FeatureID field
- [x] Update `internal/foreman/config.go` - add DefaultAgent, DefaultTechStack
- [x] Update `internal/telegram/bot.go` - add RequestPhaseApproval method
- [x] Verify: `go build ./...`

### Phase 4: Core Integration
- [x] Update `internal/foreman/foreman.go` - add speckit field and all phase methods
- [x] Verify: `go build ./...`

### Phase 5: Handlers
- [x] Update `internal/foreman/handlers.go` - add all new commands and callbacks
- [x] Verify: `go build ./...`

### Phase 6: Config
- [x] Update `configs/foreman.yaml`
- [x] Final build: `go build ./...`

### Phase 7: Additional Improvements
- [x] Add sample config file `configs/foreman.yaml.example`
- [x] Add missing Git repo operations (GetCurrentBranch, HasUncommittedChanges, etc.)
- [x] Make test runner configurable in reviewer
- [x] Add retry callback handler for escalations
- [x] Fix task sequencing for parallel vs sequential tasks
- [x] Add feedback handler for rejected phases
- [x] Add /constitution command for project principles
- [x] Add proper phase transitions during task review

## Build Status

All components successfully implemented and verified with `go build ./...`

## Summary

The Foreman multi-agent orchestrator is complete. The following functionality is available:

### Telegram Commands:
- `/newfeature <name> | <description>` - Start new feature workflow
- `/features` - List all active features
- `/feature <id>` - Show feature status
- `/techstack <id> <stack>` - Set tech stack for feature
- `/answer <id> Q1: ans, Q2: ans` - Answer clarification questions
- `/constitution <principles>` - Set project governing principles
- `/assign <agent> <spec>` - Direct task assignment (legacy)
- `/cancel <id>` - Cancel task or feature
- `/status` - Show all active work
- `/agents` - List available agents
- `/help` - Show help message

### Workflow Phases:
1. **Specify** - Creates feature specification via SpecKit
2. **Clarify** - Gathers clarifications through Q&A
3. **Plan** - Creates technical implementation plan
4. **Task** - Generates actionable task breakdown
5. **Implement** - Executes tasks with agents
6. **Review** - Automated code review
7. **Approve** - Human approval gates

### Phase Approval Buttons:
- Spec: Approve/Request Changes
- Plan: Approve/Request Changes
- Tasks: Approve/Request Changes
- Code: Approve & Merge/Request Changes/Reject

### Task Types:
- **Sequential tasks** - Run one at a time, wait for approval between each
- **Parallel tasks** - Marked with `[P]`, run concurrently

### Feedback Handling:
- When a phase is rejected, user can type feedback
- Feedback is captured and used to re-run the phase with context

### Configuration:
- Configurable test command for reviews
- Configurable linters
- Default agent and tech stack settings
