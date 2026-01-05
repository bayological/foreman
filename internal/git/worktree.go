package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bayological/foreman/internal/validation"
)

type Worktree struct {
	Path   string
	Branch string
}

func (r *Repo) CreateWorktree(branch string) (*Worktree, error) {
	// Validate branch name to prevent path traversal attacks
	if !validation.IsValidBranchName(branch) {
		return nil, fmt.Errorf("invalid branch name: %s", branch)
	}

	wtPath := filepath.Join(r.worktrees, branch)

	// Create branch from main if it doesn't exist
	r.git("fetch", r.remote, r.mainBranch)
	r.git("branch", branch, fmt.Sprintf("%s/%s", r.remote, r.mainBranch))

	// Remove existing worktree if present
	r.git("worktree", "remove", wtPath, "--force")
	os.RemoveAll(wtPath)

	// Create worktree
	cmd := exec.Command("git", "worktree", "add", wtPath, branch)
	cmd.Dir = r.path
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to create worktree: %s: %w", out, err)
	}

	return &Worktree{Path: wtPath, Branch: branch}, nil
}

func (r *Repo) RemoveWorktree(branch string) error {
	wtPath := filepath.Join(r.worktrees, branch)

	cmd := exec.Command("git", "worktree", "remove", wtPath, "--force")
	cmd.Dir = r.path
	cmd.Run() // Ignore errors

	os.RemoveAll(wtPath)
	return nil
}

func (r *Repo) CommitAndPush(wt *Worktree, message string) error {
	// Stage all changes
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = wt.Path
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %s: %w", out, err)
	}

	// Check if there are changes to commit
	cmd = exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = wt.Path
	if err := cmd.Run(); err == nil {
		// No changes to commit
		return nil
	}

	// Commit
	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = wt.Path
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %s: %w", out, err)
	}

	// Push (use --force-with-lease for safer force push)
	cmd = exec.Command("git", "push", "-u", r.remote, wt.Branch, "--force-with-lease")
	cmd.Dir = wt.Path
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push failed: %s: %w", out, err)
	}

	return nil
}

func (r *Repo) MergeBranch(branch string) error {
	// Fetch latest
	if _, err := r.git("fetch", r.remote); err != nil {
		return fmt.Errorf("fetch failed: %w", err)
	}

	// Checkout main
	if _, err := r.git("checkout", r.mainBranch); err != nil {
		return fmt.Errorf("checkout main failed: %w", err)
	}

	// Pull latest
	if _, err := r.git("pull", r.remote, r.mainBranch); err != nil {
		return fmt.Errorf("pull failed: %w", err)
	}

	// Merge branch
	if _, err := r.git("merge", branch, "--no-ff", "-m", fmt.Sprintf("Merge %s", branch)); err != nil {
		return fmt.Errorf("merge failed: %w", err)
	}

	// Push
	if _, err := r.git("push", r.remote, r.mainBranch); err != nil {
		return fmt.Errorf("push failed: %w", err)
	}

	// Delete branch
	r.git("branch", "-d", branch)
	r.git("push", r.remote, "--delete", branch)

	return nil
}

func (r *Repo) DeleteBranch(branch string) error {
	r.git("branch", "-D", branch)
	r.git("push", r.remote, "--delete", branch)
	return nil
}