# Foreman SpecKit Integration Implementation

You are implementing the SpecKit integration for Foreman, a multi-agent orchestration system written in Go.

## Current State

Foreman already has:
- Telegram bot integration (`internal/telegram/bot.go`)
- Git worktree management (`internal/git/`)
- Agent execution - Claude Code, Codex (`internal/agents/`)
- Review agent (`internal/agents/reviewer.go`)
- Basic task queue and execution (`internal/foreman/`)
- Configuration loading (`internal/foreman/config.go`)

## What You Must Implement

Add SpecKit integration to enable a structured workflow: spec → plan → tasks → implement

### 1. Create `internal/speckit/` directory with three files:

#### internal/speckit/commands.go
Define SpecKit command constants:
- `CmdConstitution`, `CmdSpecify`, `CmdClarify`, `CmdPlan`, `CmdTasks`, `CmdImplement`
- `CommandInfo` struct with Name, Description, NeedsArgs, ArgsHint
- `Commands` map

#### internal/speckit/speckit.go
SpecKit CLI wrapper:
- `SpecKit` struct with `repoPath` and `specifyPath` fields
- `New(repoPath string)` constructor
- `Initialize(ctx context.Context)` - runs `specify init` if needed
- `IsInitialized() bool`
- `RunClaudeCommand(ctx, command, args, workDir)` - executes slash commands via Claude Code CLI
- `Constitution(ctx, principles)`, `Specify(ctx, description, branch)`, `Clarify(ctx)`, `Plan(ctx, techStack)`, `Tasks(ctx)` - wrapper methods
- `GetSpecsDir()`, `GetLatestFeatureDir()`, `GetFeatureDir(prefix)` - path helpers

#### internal/speckit/parser.go
Parse SpecKit output files:
- `Spec` struct: Title, Description, UserStories, Requirements, RawContent, FilePath
- `UserStory` struct: ID, Title, Description, Acceptance
- `Plan` struct: Overview, TechStack, Architecture, Phases, RawContent, FilePath
- `PlanPhase` struct: Name, Description, Steps
- `TaskItem` struct: ID, Title, Description, UserStoryRef, Dependencies, FilePaths, IsParallel, IsTest, Order
- `Question` struct: ID, Question, Context, Answered, Answer
- `ParseSpec(featureDir)`, `ParsePlan(featureDir)`, `ParseTasks(featureDir)`, `ParseClarifications(output)` functions
- `Summary()` methods on Spec and Plan for Telegram display

### 2. Create `internal/foreman/workflow.go`

Phase state machine:
- `Phase` type (string)
- Constants: `PhaseIdle`, `PhaseSpecifying`, `PhaseAwaitingSpecApproval`, `PhaseClarifying`, `PhasePlanning`, `PhaseAwaitingPlanApproval`, `PhaseTasking`, `PhaseAwaitingTaskApproval`, `PhaseImplementing`, `PhaseReviewing`, `PhaseAwaitingCodeApproval`, `PhaseComplete`, `PhaseFailed`
- `validTransitions` map defining allowed state transitions
- `CanTransition(from, to Phase) bool`
- `PhaseInfo` struct: Emoji, Name, Description, NeedsHuman
- `phaseInfo` map with display info for each phase
- `Phase.Info()` and `Phase.String()` methods
- `WorkflowEvent` struct: Timestamp, FromPhase, ToPhase, Message, Actor

### 3. Create `internal/foreman/feature.go`

Feature lifecycle tracking:
- `Feature` struct with: ID, Name, Description, Branch, Phase, CurrentTask, Spec, Plan, Tasks, TaskIndex, PendingQuestions, Answers, Events, CreatedAt, UpdatedAt, TechStack, Constraints, mu (sync.RWMutex)
- `NewFeature(id, name, description)` constructor
- `Transition(to Phase, message, actor)` - validates and records state change
- `GetPhase()`, `SetSpec()`, `SetPlan()`, `SetTasks()` - thread-safe accessors
- `NextTask()` - returns next task and advances index
- `HasMoreTasks() bool`
- `Progress() string` - returns phase + task progress
- `StatusReport() string` - full status for Telegram
- `sanitizeBranchName(name)` helper

### 4. Modify `internal/foreman/task.go`

Add field to Task struct:
```go
FeatureID string
```

### 5. Modify `internal/foreman/config.go`

Add to Config struct:
```go
DefaultAgent     string `yaml:"default_agent"`
DefaultTechStack string `yaml:"default_tech_stack"`
```

Add defaults in LoadConfig:
```go
if cfg.DefaultAgent == "" {
    cfg.DefaultAgent = "claude-code"
}
```

### 6. Modify `internal/telegram/bot.go`

Add `RequestPhaseApproval(featureID, phase, summary, extra string) error` method:
- Creates phase-specific inline keyboard buttons
- Phases: "spec", "plan", "tasks", "code"
- Each phase has Approve and Request Changes buttons
- Code phase also has Reject button

### 7. Modify `internal/foreman/foreman.go`

Add fields to Foreman struct:
```go
speckit    *speckit.SpecKit
features   map[string]*Feature
featuresMu sync.RWMutex
```

Add import: `"foreman/internal/speckit"`

In New():
```go
f.speckit = speckit.New(cfg.Repo.Path)
f.features = make(map[string]*Feature)
```

In Run(), after telegram setup:
```go
if err := f.speckit.Initialize(ctx); err != nil {
    log.Printf("Warning: SpecKit init failed: %v", err)
}
```

Add methods:
- `StartFeature(ctx, name, description)` - creates feature, starts spec phase
- `runSpecificationPhase(ctx, feature)` - runs speckit.specify, requests approval
- `requestSpecApproval(feature)` - sends Telegram approval request
- `ApproveSpec(ctx, featureID)` - transitions to clarification
- `runClarificationPhase(ctx, feature)` - runs speckit.clarify, sends questions
- `sendClarificationQuestions(feature)` - formats questions for Telegram
- `AnswerClarification(ctx, featureID, answers)` - stores answers, proceeds if complete
- `runPlanningPhase(ctx, feature)` - runs speckit.plan, requests approval
- `requestPlanApproval(feature)` - sends Telegram approval request
- `ApprovePlan(ctx, featureID)` - transitions to tasking
- `runTaskingPhase(ctx, feature)` - runs speckit.tasks, parses into Task objects
- `requestTaskApproval(feature, taskItems)` - sends task list for approval
- `ApproveTasks(ctx, featureID)` - starts implementation
- `runImplementationPhase(ctx, feature)` - queues all tasks
- `completeFeature(feature)` - marks complete, sends celebration message
- `handlePhaseError(feature, err)` - transitions to failed, notifies
- `getFeature(id)` - thread-safe feature lookup
- `generateID()` - creates unique ID

### 8. Update `internal/foreman/handlers.go`

Register new commands in registerHandlers():
- `/newfeature` - handleNewFeature (starts feature workflow)
- `/features` - handleListFeatures (lists all features)
- `/feature` - handleFeatureStatus (single feature status)
- `/techstack` - handleSetTechStack (sets tech stack for feature)
- `/answer` - handleAnswer (answers clarification questions)
- `/cancel` - handleCancel (cancels feature)

Register callbacks:
- `approve_spec`, `reject_spec`
- `approve_plan`, `reject_plan`
- `approve_tasks`, `reject_tasks`
- `approve_code`, `reject_code`, `request_changes`

Implement all handlers:
- Parse command arguments
- Call appropriate Foreman methods
- Send feedback via Telegram

### 9. Update `configs/foreman.yaml`

Add:
```yaml
default_agent: claude-code
default_tech_stack: ""
```

## Implementation Rules

1. **COMMIT AFTER EVERY FILE** - `git add -A && git commit -m "description of change"`
2. **RUN `go build`** after each change to verify compilation
3. **ONE FILE AT A TIME** - complete one file before moving to the next
4. **FOLLOW THE ORDER** - implement in the order listed above (dependencies matter)
5. **USE EXISTING PATTERNS** - look at existing code for style guidance
6. **TEST IMPORTS** - ensure all imports resolve correctly

## Error Prevention Signs

- DO NOT create duplicate type definitions
- DO NOT forget to add new imports when using new packages
- DO NOT modify files outside the scope of this task
- DO NOT skip the `go build` check after each file
- DO NOT forget sync.RWMutex for thread-safe maps
- DO NOT forget to handle errors - always check err != nil
- DO NOT use `log.Fatal` in library code - return errors instead
- ALWAYS use `context.Context` for cancellable operations
- ALWAYS close resources (files, etc.) with defer

## Current Progress Tracking

Maintain your progress in `.ralph/TODO.md`:
- Mark completed files with [x]
- Note any issues encountered
- Track which step you're on

## Start

Begin by reading the existing codebase structure:
1. `ls -la internal/`
2. Read `internal/foreman/foreman.go` to understand current structure
3. Read `internal/telegram/bot.go` to understand bot patterns
4. Create `.ralph/TODO.md` with your implementation plan
5. Start implementing from step 1

Remember: commit after every single file change.