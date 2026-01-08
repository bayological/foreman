# Foreman Implementation TODO

## Current Status
The codebase is **functional and complete** with all core features implemented.

## Completed
- [x] Core orchestration (internal/foreman/)
- [x] Agent implementations (claude-code, codex)
- [x] Git integration with worktrees
- [x] Telegram bot with approval flows
- [x] SpecKit integration
- [x] Review tools (CodeRabbit, linters)
- [x] State persistence
- [x] Unit tests for all packages

## In Progress
- [ ] Add integration tests for end-to-end workflow

## Cleanup Tasks
- [ ] Remove unused telegram/handlers.go (handlers are in foreman/handlers.go)
- [ ] Remove or update pkg/config/config.go (duplicate/unused)

## Future Enhancements
1. Metrics/Telemetry - Track task success rates, durations
2. Slack Integration - Alternative to Telegram
3. Dashboard - Web UI for monitoring
4. More integration tests
