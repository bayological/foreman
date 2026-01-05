package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration.
type Config struct {
	Agents   AgentsConfig   `yaml:"agents"`
	Git      GitConfig      `yaml:"git"`
	Telegram TelegramConfig `yaml:"telegram"`
	Tools    ToolsConfig    `yaml:"tools"`
}

// AgentsConfig holds agent configuration.
type AgentsConfig struct {
	Claude ClaudeConfig `yaml:"claude"`
	Codex  CodexConfig  `yaml:"codex"`
}

// ClaudeConfig holds Claude agent configuration.
type ClaudeConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

// CodexConfig holds Codex agent configuration.
type CodexConfig struct {
	APIKey string `yaml:"api_key"`
	Model  string `yaml:"model"`
}

// GitConfig holds git configuration.
type GitConfig struct {
	DefaultRemote string `yaml:"default_remote"`
	WorktreesPath string `yaml:"worktrees_path"`
}

// TelegramConfig holds Telegram bot configuration.
type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
	ChatID   string `yaml:"chat_id"`
}

// ToolsConfig holds tools configuration.
type ToolsConfig struct {
	CodeRabbit CodeRabbitConfig `yaml:"coderabbit"`
	Linter     LinterConfig     `yaml:"linter"`
}

// CodeRabbitConfig holds CodeRabbit configuration.
type CodeRabbitConfig struct {
	APIKey string `yaml:"api_key"`
}

// LinterConfig holds linter configuration.
type LinterConfig struct {
	Enabled bool     `yaml:"enabled"`
	Rules   []string `yaml:"rules"`
}

// Load loads the configuration from a YAML file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadFromEnv loads configuration from environment variables.
func LoadFromEnv() *Config {
	return &Config{
		Agents: AgentsConfig{
			Claude: ClaudeConfig{
				APIKey: os.Getenv("CLAUDE_API_KEY"),
				Model:  os.Getenv("CLAUDE_MODEL"),
			},
			Codex: CodexConfig{
				APIKey: os.Getenv("OPENAI_API_KEY"),
				Model:  os.Getenv("CODEX_MODEL"),
			},
		},
		Telegram: TelegramConfig{
			BotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
			ChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
		},
	}
}
