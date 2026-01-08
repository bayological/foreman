# Foreman Design Document

## Overview

Foreman orchestrates AI coding agents through a structured workflow:
1. User describes feature via Telegram
2. SpecKit generates spec → plan → tasks
3. Human approves at each phase
4. Agents implement tasks in Git worktrees
5. Review agent checks code
6. Human approves PRs

## Workflow Phases

idle → specifying → awaiting_spec_approval → clarifying → planning → awaiting_plan_approval → tasking → awaiting_task_approval → implementing → reviewing → awaiting_code_approval → complete

## Packages Needed

### internal/speckit/
- Wraps SpecKit CLI (`specify init`, slash commands via `claude`)
- Parses spec.md, plan.md, tasks.md
- Commands: constitution, specify, clarify, plan, tasks

### internal/foreman/
- workflow.go: Phase type, transitions, PhaseInfo
- feature.go: Feature struct tracking full lifecycle
- Update foreman.go: Add speckit integration, phase methods
- Update handlers.go: /newfeature, /features, /techstack, /answer, approval callbacks

### internal/telegram/
- Add RequestPhaseApproval(featureID, phase, summary, extra) for approval buttons

### Config additions
- default_agent: claude-code
- default_tech_stack: ""

## Telegram Commands

/newfeature <name> | <description> - Start feature
/features - List features
/feature <id> - Status
/techstack <id> <stack> - Set tech stack
/answer <id> Q1: answer - Answer clarifications
/cancel <id> - Cancel feature

## Key Patterns

- Thread-safe maps with sync.RWMutex
- Context for cancellation
- Commit after every file
- go build after every change