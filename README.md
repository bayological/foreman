# Foreman

A Go-based multi-agent orchestration system that automates feature development through a structured, human-in-the-loop workflow. Foreman combines AI coding agents (Claude Code, OpenAI Codex) with specification-driven development, code review automation, and real-time human feedback through Telegram.

## Overview

Foreman manages the entire feature lifecycle from specification to implementation and deployment:

1. **Specification** - Generate feature specs with SpecKit
2. **Planning** - Create implementation plans with AI assistance
3. **Tasking** - Break down plans into actionable tasks
4. **Implementation** - Execute tasks using AI coding agents
5. **Review** - Automated code review with linters, tests, and AI
6. **Approval** - Human oversight at each critical phase

## Features

- **Multi-Agent Support** - Use Claude Code or OpenAI Codex for implementation
- **Specification-Driven Development** - Structured specs, plans, and tasks via SpecKit
- **Human-in-the-Loop** - Telegram integration for approvals and feedback
- **Parallel Task Execution** - Git worktrees enable safe concurrent development
- **Automated Code Review** - CodeRabbit, linters, tests, and LLM synthesis
- **Persistence** - Optional JSON storage for feature state across restarts
- **Graceful Shutdown** - Clean handling of interrupts and cancellation

## Getting Started

### Prerequisites

- **Go 1.21+** - Required to build Foreman
- **Claude Code CLI** - Primary AI coding agent (`claude` command)
- **SpecKit CLI** - Specification management (`specify` command)
- **Telegram Bot** - For human interaction and approvals
- **Git** - Version control for your project repository

Optional tools:
- **CodeRabbit CLI** - AI-powered code review
- **Linters** - eslint, golangci-lint, ruff, flake8, pylint (as needed)
- **OpenAI Codex CLI** - Alternative coding agent

### Installation

1. Clone the repository:

```bash
git clone https://github.com/bayological/foreman.git
cd foreman
```

2. Build the application:

```bash
go build -o foreman .
```

3. Create your configuration file:

```bash
cp configs/foreman.yaml.example configs/foreman.yaml
```

### Configuration

Edit `configs/foreman.yaml` to customize Foreman for your project:

```yaml
# Repository configuration
repo:
  path: /path/to/your/project
  remote: origin
  main_branch: main

# Telegram bot configuration
telegram:
  token: ${TELEGRAM_BOT_TOKEN}
  chat_id: ${TELEGRAM_CHAT_ID}

# Agent configuration
agents:
  claude-code:
    enabled: true
    timeout: 30m
    priority: 1
  codex:
    enabled: false
    timeout: 30m
    priority: 2

# Code review configuration
review:
  tools:
    coderabbit: false
    linters:
      - eslint
    test_command: "npm test"
  use_llm: true
  max_retries: 2

# Concurrency settings
concurrency:
  max_tasks: 3
  task_timeout: 30m

# Storage for feature persistence (optional)
storage:
  path: ""

# Defaults
default_agent: claude-code
default_tech_stack: ""
```

### Setting Up Telegram

1. Create a Telegram bot via [@BotFather](https://t.me/BotFather)
2. Get your bot token from BotFather
3. Get your chat ID (send a message to the bot and check the API)
4. Set environment variables:

```bash
export TELEGRAM_BOT_TOKEN="your-bot-token"
export TELEGRAM_CHAT_ID="your-chat-id"
```

### Running Foreman

Start Foreman with the default config:

```bash
./foreman
```

Or specify a custom config path:

```bash
./foreman -config /path/to/config.yaml
```

## Usage

### Telegram Commands

| Command | Description |
|---------|-------------|
| `/newfeature <name> \| <description>` | Start a new feature development |
| `/features` | List all active features |
| `/feature <id>` | View details of a specific feature |
| `/techstack <stack>` | Set the tech stack for the current feature |
| `/answer <text>` | Answer clarifying questions from SpecKit |
| `/constitution` | View the system's operating principles |
| `/assign <agent>` | Manually assign an agent to a task |
| `/cancel` | Cancel the current task |

### Workflow

1. **Create Feature**: Send `/newfeature Add user authentication | Implement JWT-based login and registration`

2. **Approve Specification**: Review the generated spec and tap "Approve" or "Request Changes"

3. **Answer Questions**: If SpecKit needs clarification, answer with `/answer <your response>`

4. **Approve Plan**: Review the implementation plan and approve

5. **Approve Tasks**: Review the task breakdown and approve

6. **Monitor Progress**: Watch as agents implement tasks with automatic code review

7. **Approve Code**: Review and approve completed task implementations

8. **Complete**: Feature is merged to main when all tasks are approved

### Task Execution

- Tasks marked with `[P]` execute in parallel using separate Git worktrees
- Sequential tasks queue and execute one at a time
- Failed tasks retry automatically (configurable max retries)
- Blocking issues escalate for human intervention

## Architecture

```
foreman/
├── main.go                 # Application entry point
├── configs/                # Configuration files
│   └── foreman.yaml.example
└── internal/
    ├── foreman/            # Core orchestration
    │   ├── foreman.go      # Main orchestrator
    │   ├── workflow.go     # Phase state machine
    │   ├── feature.go      # Feature management
    │   ├── task.go         # Task representation
    │   ├── handlers.go     # Telegram handlers
    │   └── config.go       # Configuration
    ├── agents/             # AI coding agents
    │   ├── agent.go        # Agent interface
    │   ├── claude.go       # Claude Code integration
    │   ├── codex.go        # OpenAI Codex integration
    │   └── reviewer.go     # Review orchestration
    ├── telegram/           # Telegram bot
    │   ├── bot.go          # Bot wrapper
    │   └── notifications.go
    ├── speckit/            # SpecKit integration
    │   ├── speckit.go      # CLI wrapper
    │   └── parser.go       # Spec/plan/task parsing
    ├── git/                # Git operations
    │   ├── repo.go         # Repository wrapper
    │   └── worktree.go     # Worktree management
    ├── storage/            # Feature persistence
    │   └── storage.go      # JSON file storage
    └── tools/              # Review tools
        ├── coderabbit.go   # CodeRabbit integration
        ├── linter.go       # Multi-linter support
        └── runner.go       # Command runner
```

## Feature Phases

| Phase | Description | Approval Required |
|-------|-------------|-------------------|
| Specifying | Generating feature specification | No |
| AwaitingSpecApproval | Waiting for spec approval | Yes |
| Clarifying | Answering questions | Yes |
| Planning | Creating implementation plan | No |
| AwaitingPlanApproval | Waiting for plan approval | Yes |
| Tasking | Breaking into tasks | No |
| AwaitingTaskApproval | Waiting for task approval | Yes |
| Implementing | Executing tasks | No |
| Reviewing | Running code review | No |
| AwaitingCodeApproval | Waiting for code approval | Yes |
| Complete | Feature finished | No |

## Dependencies

### Go Modules

- `github.com/go-telegram-bot-api/telegram-bot-api/v5` - Telegram Bot API
- `github.com/google/uuid` - UUID generation
- `gopkg.in/yaml.v3` - YAML configuration

### External Tools

- `claude` - Claude Code CLI
- `specify` - SpecKit CLI
- `codex` - OpenAI Codex CLI (optional)
- `coderabbit` - CodeRabbit CLI (optional)
- Linters as configured

## License

See [LICENSE](LICENSE) for details.
