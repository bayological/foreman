Your job is to complete the Foreman implementation - a multi-agent orchestrator in Go that uses SpecKit for spec-driven development, Telegram for human-in-the-loop communication, and Git worktrees for parallel agent work.

Reference the design document at `.ralph/DESIGN.md` for architecture details.

The codebase is in `internal/`. Key packages: foreman, agents, git, telegram, speckit (needs creation).

Make a commit and push after every single file edit.

Use `.ralph/agent/` as your scratchpad. Store plans, todos, and notes there.

Run `go build ./...` after each change to verify compilation.