package tools

import (
	"context"
	"fmt"
	"strings"
)

type Linter struct {
	linters []string
}

func NewLinter(linters ...string) *Linter {
	if len(linters) == 0 {
		linters = []string{"eslint", "ruff"}
	}
	return &Linter{linters: linters}
}

// linterConfig holds linter-specific configuration
type linterConfig struct {
	command string
	args    []string
	check   string // command to check availability
}

var linterConfigs = map[string]linterConfig{
	"eslint": {
		command: "npx",
		args:    []string{"eslint", ".", "--format", "compact"},
		check:   "npx",
	},
	"ruff": {
		command: "ruff",
		args:    []string{"check", "."},
		check:   "ruff",
	},
	"golangci-lint": {
		command: "golangci-lint",
		args:    []string{"run", "./..."},
		check:   "golangci-lint",
	},
	"flake8": {
		command: "flake8",
		args:    []string{"."},
		check:   "flake8",
	},
	"pylint": {
		command: "pylint",
		args:    []string{"."},
		check:   "pylint",
	},
}

func (l *Linter) Run(ctx context.Context, workDir string) (string, error) {
	var results []string

	for _, linter := range l.linters {
		var output string
		var err error

		cfg, known := linterConfigs[linter]
		if known {
			// Check if linter is available
			if !CommandAvailable(cfg.check) {
				results = append(results, fmt.Sprintf("%s: not installed (skipped)", linter))
				continue
			}
			output, err = RunCommand(ctx, workDir, cfg.command, cfg.args...)
		} else {
			// Unknown linter - try to run it directly
			if !CommandAvailable(linter) {
				results = append(results, fmt.Sprintf("%s: not installed (skipped)", linter))
				continue
			}
			output, err = RunCommand(ctx, workDir, linter)
		}

		if err != nil {
			// Linter found issues (exit code != 0 is normal)
			if output != "" {
				results = append(results, fmt.Sprintf("%s:\n%s", linter, output))
			} else {
				results = append(results, fmt.Sprintf("%s error: %v", linter, err))
			}
		} else if output != "" {
			results = append(results, fmt.Sprintf("%s:\n%s", linter, output))
		} else {
			results = append(results, fmt.Sprintf("%s: no issues found", linter))
		}
	}

	if len(results) == 0 {
		return "No linters configured or available", nil
	}

	return strings.Join(results, "\n\n"), nil
}