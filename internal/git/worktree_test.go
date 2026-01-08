package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateWorktreeInvalidBranch(t *testing.T) {
	tmpDir := setupGitRepo(t)
	defer os.RemoveAll(tmpDir)

	repo, err := NewRepo(tmpDir, "origin", "main")
	if err != nil {
		t.Fatalf("NewRepo failed: %v", err)
	}

	// Try to create worktree with invalid branch name (path traversal)
	_, err = repo.CreateWorktree("../../../etc/passwd")
	if err == nil {
		t.Error("Expected error for path traversal branch name")
	}

	// Try with double dots
	_, err = repo.CreateWorktree("feature/../test")
	if err == nil {
		t.Error("Expected error for branch name containing '..'")
	}

	// Try with empty branch name
	_, err = repo.CreateWorktree("")
	if err == nil {
		t.Error("Expected error for empty branch name")
	}
}

func TestWorktreeStruct(t *testing.T) {
	wt := &Worktree{
		Path:   "/test/path",
		Branch: "feature/test",
	}

	if wt.Path != "/test/path" {
		t.Errorf("Expected path '/test/path', got %s", wt.Path)
	}
	if wt.Branch != "feature/test" {
		t.Errorf("Expected branch 'feature/test', got %s", wt.Branch)
	}
}

func TestRemoveWorktreeNonExistent(t *testing.T) {
	tmpDir := setupGitRepo(t)
	defer os.RemoveAll(tmpDir)

	repo, err := NewRepo(tmpDir, "origin", "main")
	if err != nil {
		t.Fatalf("NewRepo failed: %v", err)
	}

	// Remove non-existent worktree should not error (idempotent)
	err = repo.RemoveWorktree("nonexistent-branch")
	if err != nil {
		t.Errorf("RemoveWorktree should not fail for non-existent branch: %v", err)
	}
}

func TestWorktreesDirectory(t *testing.T) {
	tmpDir := setupGitRepo(t)
	defer os.RemoveAll(tmpDir)

	_, err := NewRepo(tmpDir, "origin", "main")
	if err != nil {
		t.Fatalf("NewRepo failed: %v", err)
	}

	// Verify worktrees directory was created
	worktreesDir := filepath.Join(tmpDir, ".worktrees")
	stat, err := os.Stat(worktreesDir)
	if os.IsNotExist(err) {
		t.Fatal("Worktrees directory should be created by NewRepo")
	}
	if !stat.IsDir() {
		t.Error("Worktrees path should be a directory")
	}
}

// Note: Full CreateWorktree integration tests require a remote repository setup.
// The function relies on fetching from a remote and creating a branch from the
// remote's main branch. For comprehensive testing, use a real git server or
// mock the git commands.
//
// TestCreateWorktree tests have been removed because they require:
// 1. A configured remote (origin)
// 2. The remote to have the main branch
// 3. Network access to fetch from the remote
//
// In production, CreateWorktree is designed to work with GitHub/GitLab remotes
// where the repository already exists with proper branch structure.
