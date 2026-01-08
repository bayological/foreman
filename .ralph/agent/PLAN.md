# Foreman Implementation Plan

## Current State Analysis

The codebase is largely complete with the core architecture in place.

### Implemented Components:
- **Core Orchestration** (`internal/foreman/`)
- **Agents** (`internal/agents/`)
- **Git Integration** (`internal/git/`)
- **Telegram** (`internal/telegram/`)
- **SpecKit** (`internal/speckit/`)
- **Tools** (`internal/tools/`)
- **Validation** (`internal/validation/`)

## Issues to Fix

### 1. Missing Sample Config File
Need `configs/foreman.yaml.example` for users to copy.

### 2. Missing Git Repo Methods
Need methods: GetCurrentBranch, HasUncommittedChanges.

### 3. Hardcoded Test Runner
Reviewer uses `npm test` - should be configurable.

### 4. Missing Retry Callback Handler
Escalation shows "Retry" button but no handler registered.

### 5. Feature Task Sequencing
Tasks queued all at once without waiting for sequential completion.

### 6. Missing Feedback Handler for Phase Rejections
When user clicks "Request Changes" they're told to provide feedback but no handler captures it.

## Implementation Order

1. Create sample config file
2. Add missing Git operations
3. Make test runner configurable
4. Add retry callback handler
5. Fix task sequencing logic
6. Add feedback capture mechanism
7. Add feature persistence (optional)
