package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setupGitRepo creates a temporary git repo with initial commit
func setupGitRepo(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "foreman-git-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	runGit(t, tmpDir, "init")
	runGit(t, tmpDir, "config", "user.email", "test@test.com")
	runGit(t, tmpDir, "config", "user.name", "Test User")
	runGit(t, tmpDir, "config", "commit.gpgsign", "false")

	// Create initial commit
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to create test file: %v", err)
	}

	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "Initial commit")

	return tmpDir
}

// runGit executes a git command and fails the test on error
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\nOutput: %s", args, err, output)
	}
	return string(output)
}

func TestNewRepo(t *testing.T) {
	// Test with non-git directory (should fail)
	tmpDir, err := os.MkdirTemp("", "foreman-git-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = NewRepo(tmpDir, "origin", "main")
	if err == nil {
		t.Error("Expected error for non-git directory")
	}

	// Initialize git repo
	runGit(t, tmpDir, "init")

	// Test with git directory (should succeed)
	repo, err := NewRepo(tmpDir, "origin", "main")
	if err != nil {
		t.Fatalf("NewRepo failed: %v", err)
	}

	if repo.path != tmpDir {
		t.Errorf("Expected path %s, got %s", tmpDir, repo.path)
	}
	if repo.remote != "origin" {
		t.Errorf("Expected remote 'origin', got %s", repo.remote)
	}
	if repo.mainBranch != "main" {
		t.Errorf("Expected main branch 'main', got %s", repo.mainBranch)
	}

	// Verify worktrees directory was created
	worktreesDir := filepath.Join(tmpDir, ".worktrees")
	if _, err := os.Stat(worktreesDir); os.IsNotExist(err) {
		t.Error("Expected .worktrees directory to be created")
	}
}

func TestRepoPath(t *testing.T) {
	repo := &Repo{path: "/test/path"}
	if repo.Path() != "/test/path" {
		t.Errorf("Expected '/test/path', got %s", repo.Path())
	}
}

func TestRepoMainBranch(t *testing.T) {
	repo := &Repo{mainBranch: "master"}
	if repo.MainBranch() != "master" {
		t.Errorf("Expected 'master', got %s", repo.MainBranch())
	}
}

func TestGetCurrentBranch(t *testing.T) {
	tmpDir := setupGitRepo(t)
	defer os.RemoveAll(tmpDir)

	repo, err := NewRepo(tmpDir, "origin", "main")
	if err != nil {
		t.Fatalf("NewRepo failed: %v", err)
	}

	branch, err := repo.GetCurrentBranch()
	if err != nil {
		t.Fatalf("GetCurrentBranch failed: %v", err)
	}

	// Git init creates 'master' or 'main' depending on git version
	if branch != "master" && branch != "main" {
		t.Errorf("Expected 'master' or 'main', got %s", branch)
	}
}

func TestHasUncommittedChanges(t *testing.T) {
	tmpDir := setupGitRepo(t)
	defer os.RemoveAll(tmpDir)

	repo, err := NewRepo(tmpDir, "origin", "main")
	if err != nil {
		t.Fatalf("NewRepo failed: %v", err)
	}

	// Initially no uncommitted changes (just untracked .worktrees)
	hasChanges, err := repo.HasUncommittedChanges()
	if err != nil {
		t.Fatalf("HasUncommittedChanges failed: %v", err)
	}

	// Create a new file
	testFile := filepath.Join(tmpDir, "new-file.txt")
	if err := os.WriteFile(testFile, []byte("new content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hasChanges, err = repo.HasUncommittedChanges()
	if err != nil {
		t.Fatalf("HasUncommittedChanges failed: %v", err)
	}
	if !hasChanges {
		t.Error("Expected uncommitted changes after creating file")
	}
}

func TestCreateBranch(t *testing.T) {
	tmpDir := setupGitRepo(t)
	defer os.RemoveAll(tmpDir)

	repo, err := NewRepo(tmpDir, "origin", "main")
	if err != nil {
		t.Fatalf("NewRepo failed: %v", err)
	}

	// Create a new branch
	err = repo.CreateBranch("test-branch")
	if err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}

	// Verify we're on the new branch
	branch, err := repo.GetCurrentBranch()
	if err != nil {
		t.Fatalf("GetCurrentBranch failed: %v", err)
	}
	if branch != "test-branch" {
		t.Errorf("Expected 'test-branch', got %s", branch)
	}
}

func TestCheckoutBranch(t *testing.T) {
	tmpDir := setupGitRepo(t)
	defer os.RemoveAll(tmpDir)

	repo, err := NewRepo(tmpDir, "origin", "main")
	if err != nil {
		t.Fatalf("NewRepo failed: %v", err)
	}

	// Get current branch name
	originalBranch, _ := repo.GetCurrentBranch()

	// Create a new branch
	err = repo.CreateBranch("test-branch")
	if err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}

	// Checkout original branch
	err = repo.CheckoutBranch(originalBranch)
	if err != nil {
		t.Fatalf("CheckoutBranch failed: %v", err)
	}

	branch, err := repo.GetCurrentBranch()
	if err != nil {
		t.Fatalf("GetCurrentBranch failed: %v", err)
	}
	if branch != originalBranch {
		t.Errorf("Expected '%s', got %s", originalBranch, branch)
	}
}

func TestDeleteBranch(t *testing.T) {
	tmpDir := setupGitRepo(t)
	defer os.RemoveAll(tmpDir)

	repo, err := NewRepo(tmpDir, "origin", "main")
	if err != nil {
		t.Fatalf("NewRepo failed: %v", err)
	}

	// Get current branch name
	originalBranch, _ := repo.GetCurrentBranch()

	// Create a new branch
	err = repo.CreateBranch("branch-to-delete")
	if err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}

	// Go back to original branch
	err = repo.CheckoutBranch(originalBranch)
	if err != nil {
		t.Fatalf("CheckoutBranch failed: %v", err)
	}

	// Delete the branch
	err = repo.DeleteBranch("branch-to-delete")
	if err != nil {
		t.Fatalf("DeleteBranch failed: %v", err)
	}

	// Verify branch is deleted by trying to check it out
	err = repo.CheckoutBranch("branch-to-delete")
	if err == nil {
		t.Error("Expected error when checking out deleted branch")
	}
}
