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

// GetCurrentBranch returns the name of the current branch.
func (r *Repo) GetCurrentBranch() (string, error) {
	return r.git("rev-parse", "--abbrev-ref", "HEAD")
}

// HasUncommittedChanges returns true if there are uncommitted changes in the repo.
func (r *Repo) HasUncommittedChanges() (bool, error) {
	output, err := r.git("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return output != "", nil
}

// CheckoutBranch switches to the specified branch.
func (r *Repo) CheckoutBranch(branch string) error {
	_, err := r.git("checkout", branch)
	return err
}

// CreateBranch creates a new branch from the current HEAD.
func (r *Repo) CreateBranch(branch string) error {
	_, err := r.git("checkout", "-b", branch)
	return err
}

// Path returns the repository path.
func (r *Repo) Path() string {
	return r.path
}

// MainBranch returns the configured main branch name.
func (r *Repo) MainBranch() string {
	return r.mainBranch
}

// PRResult contains information about a created pull request.
type PRResult struct {
	URL    string
	Number int
}

// CreatePullRequest creates a GitHub PR using the gh CLI.
func (r *Repo) CreatePullRequest(branch, title, body string) (*PRResult, error) {
	// Check if gh is available
	if _, err := exec.LookPath("gh"); err != nil {
		return nil, fmt.Errorf("gh CLI not found: %w", err)
	}

	// Create the PR
	cmd := exec.Command("gh", "pr", "create",
		"--head", branch,
		"--base", r.mainBranch,
		"--title", title,
		"--body", body,
	)
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gh pr create failed: %s: %w", output, err)
	}

	// Output contains the PR URL
	url := strings.TrimSpace(string(output))

	return &PRResult{URL: url}, nil
}

// GetPullRequestURL returns the URL for an existing PR on the given branch.
func (r *Repo) GetPullRequestURL(branch string) (string, error) {
	cmd := exec.Command("gh", "pr", "view", branch, "--json", "url", "-q", ".url")
	cmd.Dir = r.path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gh pr view failed: %s: %w", output, err)
	}
	return strings.TrimSpace(string(output)), nil
}