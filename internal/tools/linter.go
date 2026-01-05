package tools

import (
	"context"
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

func (l *Linter) Run(ctx context.Context, workDir string) (string, error) {
	var results []string

	for _, linter := range l.linters {
		var output string
		var err error

		switch linter {
		case "eslint":
			output, err = RunCommand(ctx, workDir, "npx", "eslint", ".", "--format", "compact")
		case "ruff":
			output, err = RunCommand(ctx, workDir, "ruff", "check", ".")
		default:
			output, err = RunCommand(ctx, workDir, linter)
		}

		if output != "" {
			results = append(results, linter+":\n"+output)
		}
		if err != nil {
			results = append(results, linter+" error: "+err.Error())
		}
	}

	return strings.Join(results, "\n\n"), nil
}