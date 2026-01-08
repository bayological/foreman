# Foreman Implementation Plan

## Current State Analysis

The codebase is complete with all core functionality in place.

### Implemented Components:
- **Core Orchestration** (`internal/foreman/`)
- **Agents** (`internal/agents/`)
- **Git Integration** (`internal/git/`)
- **Telegram** (`internal/telegram/`)
- **SpecKit** (`internal/speckit/`)
- **Tools** (`internal/tools/`)
- **Validation** (`internal/validation/`)

## Completed Tasks

### 1. Sample Config File [DONE]
Created `configs/foreman.yaml.example` with documentation.

### 2. Git Repo Methods [DONE]
Added: GetCurrentBranch, HasUncommittedChanges, CheckoutBranch, CreateBranch, Path, MainBranch.

### 3. Configurable Test Runner [DONE]
Added `test_command` to review config, allowing project-specific test commands.

### 4. Retry Callback Handler [DONE]
Added handler for "Retry" button in escalation dialogs.

### 5. Task Sequencing [DONE]
Fixed to properly handle parallel vs sequential tasks:
- Parallel tasks queue immediately
- Sequential tasks wait for approval before starting next

### 6. Feedback Handler [DONE]
Added complete feedback capture:
- Rejection handlers set pending feedback state
- Next text message is captured as feedback
- Re-runs appropriate phase with feedback as context

## Potential Future Enhancements

1. **Feature Persistence** - Store features to disk for recovery after restart
2. **Metrics/Telemetry** - Track task success rates, durations
3. **PR Creation** - Auto-create GitHub PRs after code approval
4. **Slack Integration** - Alternative to Telegram
5. **Dashboard** - Web UI for monitoring
