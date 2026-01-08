# Foreman Implementation Plan

## Current State Analysis

The codebase is complete with all core functionality and bug fixes in place.

### Implemented Components:
- **Core Orchestration** (`internal/foreman/`)
- **Agents** (`internal/agents/`)
- **Git Integration** (`internal/git/`)
- **Telegram** (`internal/telegram/`)
- **SpecKit** (`internal/speckit/`)
- **Tools** (`internal/tools/`)
- **Validation** (`internal/validation/`)
- **Storage** (`internal/storage/`)

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

### 7. Feature Persistence [DONE]
Added JSON file-based storage:
- New `internal/storage/` package
- Features saved to disk on create, complete, and error
- Automatic loading on startup
- Configurable via `storage.path` in config

### 8. CodeRabbit Integration [DONE]
Improved CodeRabbit and linter handling:
- Checks if CLI tools are installed before running
- Graceful skip when tools not available
- Added support for more linters: golangci-lint, flake8, pylint
- Better error messages and output formatting

### 9. Improved Error Handling - Tools [DONE]
Enhanced tools package:
- Added CommandResult type for detailed output
- Added CommandAvailable utility function
- Better stderr/stdout handling
- FormatCommandError for user-friendly messages

### 10. SpecKit Error Handling [DONE]
Fixed several issues:
- Added scanner error check in RunClaudeCommand
- Fixed git checkout error handling in Specify method
- Fixed GetLatestFeatureDir to sort by modification time (not just alphabetical)
- Removed orphaned CmdImplement constant

### 11. Agents Error Handling [DONE]
Fixed critical issues:
- Added scanner error checks in claude.go Execute and Review methods
- Added stderr collection to Review method for better error messages
- Handle diff errors in reviewer instead of ignoring them

## Potential Future Enhancements

1. **PR Creation** - Auto-create GitHub PRs after code approval
2. **Metrics/Telemetry** - Track task success rates, durations
3. **Slack Integration** - Alternative to Telegram
4. **Dashboard** - Web UI for monitoring
5. **Unit Tests** - Add test coverage for core functionality
6. **Graceful Shutdown** - Save in-progress features on shutdown
