package speckit

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type SpecKit struct {
	repoPath    string
	specifyPath string
}

func New(repoPath string) *SpecKit {
	return &SpecKit{
		repoPath:    repoPath,
		specifyPath: filepath.Join(repoPath, ".specify"),
	}
}

type CommandResult struct {
	Command   string
	Args      string
	Output    string
	Success   bool
	Error     string
	Timestamp time.Time
}

// Initialize runs 'specify init' if not already initialized
func (s *SpecKit) Initialize(ctx context.Context) error {
	if _, err := os.Stat(s.specifyPath); err == nil {
		return nil
	}

	cmd := exec.CommandContext(ctx, "specify", "init", ".", "--ai", "claude", "--force")
	cmd.Dir = s.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("specify init failed: %w\noutput: %s", err, output)
	}

	return nil
}

// IsInitialized checks if SpecKit is set up in the repo
func (s *SpecKit) IsInitialized() bool {
	_, err := os.Stat(s.specifyPath)
	return err == nil
}

// RunClaudeCommand executes a SpecKit slash command via Claude Code
func (s *SpecKit) RunClaudeCommand(ctx context.Context, command string, args string, workDir string) (*CommandResult, error) {
	prompt := fmt.Sprintf("/%s %s", command, args)

	claudeArgs := []string{
		"--print",
		"--output-format", "stream-json",
		prompt,
	}

	cmd := exec.CommandContext(ctx, "claude", claudeArgs...)
	cmd.Dir = workDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start claude: %w", err)
	}

	var output strings.Builder
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		var msg struct {
			Type    string `json:"type"`
			Content string `json:"content,omitempty"`
		}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		if msg.Type == "assistant" && msg.Content != "" {
			output.WriteString(msg.Content)
		}
	}

	cmdErr := cmd.Wait()

	result := &CommandResult{
		Command:   command,
		Args:      args,
		Output:    output.String(),
		Timestamp: time.Now(),
	}

	if cmdErr != nil {
		result.Success = false
		result.Error = cmdErr.Error()
	} else {
		result.Success = true
	}

	return result, nil
}

// Constitution creates or updates project principles
func (s *SpecKit) Constitution(ctx context.Context, principles string) (*CommandResult, error) {
	return s.RunClaudeCommand(ctx, "speckit.constitution", principles, s.repoPath)
}

// Specify creates feature specification
func (s *SpecKit) Specify(ctx context.Context, description string, featureBranch string) (*CommandResult, error) {
	// Checkout the feature branch first
	exec.CommandContext(ctx, "git", "-C", s.repoPath, "checkout", "-B", featureBranch).Run()

	return s.RunClaudeCommand(ctx, "speckit.specify", description, s.repoPath)
}

// Clarify runs the clarification workflow
func (s *SpecKit) Clarify(ctx context.Context) (*CommandResult, error) {
	return s.RunClaudeCommand(ctx, "speckit.clarify", "", s.repoPath)
}

// Plan creates implementation plan with tech stack
func (s *SpecKit) Plan(ctx context.Context, techStack string) (*CommandResult, error) {
	return s.RunClaudeCommand(ctx, "speckit.plan", techStack, s.repoPath)
}

// Tasks generates task breakdown
func (s *SpecKit) Tasks(ctx context.Context) (*CommandResult, error) {
	return s.RunClaudeCommand(ctx, "speckit.tasks", "", s.repoPath)
}

// GetSpecsDir returns the specs directory path
func (s *SpecKit) GetSpecsDir() string {
	return filepath.Join(s.specifyPath, "specs")
}

// GetLatestFeatureDir returns the most recent feature directory
func (s *SpecKit) GetLatestFeatureDir() string {
	specsDir := s.GetSpecsDir()

	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return ""
	}

	var latest string
	for _, entry := range entries {
		if entry.IsDir() {
			latest = filepath.Join(specsDir, entry.Name())
		}
	}

	return latest
}

// GetFeatureDir returns the directory for a specific feature ID prefix
func (s *SpecKit) GetFeatureDir(featurePrefix string) string {
	specsDir := s.GetSpecsDir()

	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), featurePrefix) {
			return filepath.Join(specsDir, entry.Name())
		}
	}

	return ""
}
