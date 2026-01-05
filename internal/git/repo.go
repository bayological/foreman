package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Repo struct {
	path       string
	remote     string
	mainBranch string
	worktrees  string
}

func NewRepo(path, remote, mainBranch string) (*Repo, error) {
	// Verify it's a git repo
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return nil, fmt.Errorf("not a git repository: %s", path)
	}

	worktreesDir := filepath.Join(path, ".worktrees")
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create worktrees dir: %w", err)
	}

	return &Repo{
		path:       path,
		remote:     remote,
		mainBranch: mainBranch,
		worktrees:  worktreesDir,
	}, nil
}

func (r *Repo) git(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}