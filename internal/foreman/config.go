package foreman

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Repo             RepoConfig        `yaml:"repo"`
	Telegram         TelegramConfig    `yaml:"telegram"`
	Agents           AgentsConfig      `yaml:"agents"`
	Review           ReviewConfig      `yaml:"review"`
	Concurrency      ConcurrencyConfig `yaml:"concurrency"`
	Storage          StorageConfig     `yaml:"storage"`
	DefaultAgent     string            `yaml:"default_agent"`
	DefaultTechStack string            `yaml:"default_tech_stack"`
}

type StorageConfig struct {
	Path string `yaml:"path"` // Path to features.json file
}

type RepoConfig struct {
	Path       string `yaml:"path"`
	Remote     string `yaml:"remote"`
	MainBranch string `yaml:"main_branch"`
}

type TelegramConfig struct {
	Token  string `yaml:"token"`
	ChatID int64  `yaml:"chat_id"`
}

type AgentConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Timeout  time.Duration `yaml:"timeout"`
	Priority int           `yaml:"priority"`
}

type AgentsConfig struct {
	ClaudeCode AgentConfig `yaml:"claude-code"`
	Codex      AgentConfig `yaml:"codex"`
}

type ReviewConfig struct {
	Tools      ReviewToolsConfig `yaml:"tools"`
	UseLLM     bool              `yaml:"use_llm"`
	MaxRetries int               `yaml:"max_retries"`
}

type ReviewToolsConfig struct {
	CodeRabbit  bool     `yaml:"coderabbit"`
	Linters     []string `yaml:"linters"`
	TestCommand string   `yaml:"test_command"`
}

type ConcurrencyConfig struct {
	MaxTasks    int           `yaml:"max_tasks"`
	TaskTimeout time.Duration `yaml:"task_timeout"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Expand environment variables
	expanded := os.ExpandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Set defaults
	if cfg.Repo.Remote == "" {
		cfg.Repo.Remote = "origin"
	}
	if cfg.Repo.MainBranch == "" {
		cfg.Repo.MainBranch = "main"
	}
	if cfg.Concurrency.MaxTasks == 0 {
		cfg.Concurrency.MaxTasks = 3
	}
	if cfg.Concurrency.TaskTimeout == 0 {
		cfg.Concurrency.TaskTimeout = 30 * time.Minute
	}
	if cfg.Review.MaxRetries == 0 {
		cfg.Review.MaxRetries = 2
	}
	if cfg.DefaultAgent == "" {
		cfg.DefaultAgent = "claude-code"
	}

	return &cfg, nil
}