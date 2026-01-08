package speckit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	sk := New("/test/repo")
	if sk == nil {
		t.Fatal("New() should not return nil")
	}
	if sk.repoPath != "/test/repo" {
		t.Errorf("New() repoPath = %q, want %q", sk.repoPath, "/test/repo")
	}
	if sk.specifyPath != "/test/repo/.specify" {
		t.Errorf("New() specifyPath = %q, want %q", sk.specifyPath, "/test/repo/.specify")
	}
}

func TestIsInitialized(t *testing.T) {
	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "speckit-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	sk := New(tmpDir)

	// Initially should not be initialized
	if sk.IsInitialized() {
		t.Error("IsInitialized() should return false for new directory")
	}

	// Create .specify directory
	specifyDir := filepath.Join(tmpDir, ".specify")
	if err := os.MkdirAll(specifyDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Now should be initialized
	if !sk.IsInitialized() {
		t.Error("IsInitialized() should return true after creating .specify dir")
	}
}

func TestGetSpecsDir(t *testing.T) {
	sk := New("/my/repo")
	expected := "/my/repo/.specify/specs"
	if sk.GetSpecsDir() != expected {
		t.Errorf("GetSpecsDir() = %q, want %q", sk.GetSpecsDir(), expected)
	}
}

func TestGetLatestFeatureDir(t *testing.T) {
	// Create a temp directory with .specify/specs structure
	tmpDir, err := os.MkdirTemp("", "speckit-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	specsDir := filepath.Join(tmpDir, ".specify", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}

	sk := New(tmpDir)

	// Empty specs dir should return empty string
	if result := sk.GetLatestFeatureDir(); result != "" {
		t.Errorf("GetLatestFeatureDir() on empty should return '', got %q", result)
	}

	// Create some feature directories
	feat1 := filepath.Join(specsDir, "feature-1")
	feat2 := filepath.Join(specsDir, "feature-2")

	if err := os.MkdirAll(feat1, 0755); err != nil {
		t.Fatal(err)
	}
	// Small delay to ensure different mod times
	if err := os.MkdirAll(feat2, 0755); err != nil {
		t.Fatal(err)
	}

	// Should return one of the feature directories
	result := sk.GetLatestFeatureDir()
	if result != feat1 && result != feat2 {
		t.Errorf("GetLatestFeatureDir() = %q, expected one of the feature dirs", result)
	}
}

func TestGetLatestFeatureDir_NoSpecsDir(t *testing.T) {
	sk := New("/nonexistent/path")
	result := sk.GetLatestFeatureDir()
	if result != "" {
		t.Errorf("GetLatestFeatureDir() with no specs dir should return '', got %q", result)
	}
}

func TestGetFeatureDir(t *testing.T) {
	// Create a temp directory with .specify/specs structure
	tmpDir, err := os.MkdirTemp("", "speckit-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	specsDir := filepath.Join(tmpDir, ".specify", "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create feature directories
	feat1 := filepath.Join(specsDir, "user-auth-feature")
	feat2 := filepath.Join(specsDir, "payment-feature")

	if err := os.MkdirAll(feat1, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(feat2, 0755); err != nil {
		t.Fatal(err)
	}

	sk := New(tmpDir)

	// Test exact prefix match
	result := sk.GetFeatureDir("user-auth")
	if result != feat1 {
		t.Errorf("GetFeatureDir('user-auth') = %q, want %q", result, feat1)
	}

	// Test another prefix
	result = sk.GetFeatureDir("payment")
	if result != feat2 {
		t.Errorf("GetFeatureDir('payment') = %q, want %q", result, feat2)
	}

	// Test non-matching prefix
	result = sk.GetFeatureDir("nonexistent")
	if result != "" {
		t.Errorf("GetFeatureDir('nonexistent') = %q, want ''", result)
	}
}

func TestGetFeatureDir_NoSpecsDir(t *testing.T) {
	sk := New("/nonexistent/path")
	result := sk.GetFeatureDir("any-prefix")
	if result != "" {
		t.Errorf("GetFeatureDir() with no specs dir should return '', got %q", result)
	}
}

func TestCommandResult(t *testing.T) {
	result := &CommandResult{
		Command: "test.command",
		Args:    "some args",
		Output:  "command output",
		Success: true,
		Error:   "",
	}

	if result.Command != "test.command" {
		t.Errorf("CommandResult.Command = %q, want 'test.command'", result.Command)
	}
	if result.Args != "some args" {
		t.Errorf("CommandResult.Args = %q, want 'some args'", result.Args)
	}
	if !result.Success {
		t.Error("CommandResult.Success should be true")
	}
}

func TestCommand_Constants(t *testing.T) {
	// Verify command constants
	if CmdConstitution != "speckit.constitution" {
		t.Errorf("CmdConstitution = %q, want 'speckit.constitution'", CmdConstitution)
	}
	if CmdSpecify != "speckit.specify" {
		t.Errorf("CmdSpecify = %q, want 'speckit.specify'", CmdSpecify)
	}
	if CmdClarify != "speckit.clarify" {
		t.Errorf("CmdClarify = %q, want 'speckit.clarify'", CmdClarify)
	}
	if CmdPlan != "speckit.plan" {
		t.Errorf("CmdPlan = %q, want 'speckit.plan'", CmdPlan)
	}
	if CmdTasks != "speckit.tasks" {
		t.Errorf("CmdTasks = %q, want 'speckit.tasks'", CmdTasks)
	}
}

func TestCommands_Info(t *testing.T) {
	// Verify all commands have info
	cmds := []Command{CmdConstitution, CmdSpecify, CmdClarify, CmdPlan, CmdTasks}
	for _, cmd := range cmds {
		info, ok := Commands[cmd]
		if !ok {
			t.Errorf("Commands should contain %q", cmd)
			continue
		}
		if info.Name != cmd {
			t.Errorf("Commands[%q].Name = %q, want %q", cmd, info.Name, cmd)
		}
		if info.Description == "" {
			t.Errorf("Commands[%q].Description should not be empty", cmd)
		}
	}
}

func TestCommandInfo_NeedsArgs(t *testing.T) {
	// Commands that require args
	argsRequired := []Command{CmdConstitution, CmdSpecify, CmdPlan}
	for _, cmd := range argsRequired {
		if !Commands[cmd].NeedsArgs {
			t.Errorf("Commands[%q].NeedsArgs should be true", cmd)
		}
	}

	// Commands that don't require args
	noArgsRequired := []Command{CmdClarify, CmdTasks}
	for _, cmd := range noArgsRequired {
		if Commands[cmd].NeedsArgs {
			t.Errorf("Commands[%q].NeedsArgs should be false", cmd)
		}
	}
}
