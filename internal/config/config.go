package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	OpenRouter OpenRouterConfig `mapstructure:"openrouter"`
	Agent      AgentConfig      `mapstructure:"agent"`
}

type OpenRouterConfig struct {
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
	BaseURL string `mapstructure:"base_url"`
}

type AgentConfig struct {
	ConfirmDestructive bool    `mapstructure:"confirm_destructive"`
	MaxTokens          int     `mapstructure:"max_tokens"`
	Temperature        float64 `mapstructure:"temperature"`
}

func Load() (*Config, error) {
	viper.SetConfigName(".agent_go")
	viper.SetConfigType("yaml")

	home, err := os.UserHomeDir()
	if err == nil {
		viper.AddConfigPath(home)
	}
	viper.AddConfigPath(".")

	viper.SetDefault("openrouter.model", "anthropic/claude-3.5-sonnet")
	viper.SetDefault("openrouter.base_url", "https://openrouter.ai/api/v1")
	viper.SetDefault("agent.confirm_destructive", true)
	viper.SetDefault("agent.max_tokens", 4000)
	viper.SetDefault("agent.temperature", 0.7)

	// env variables
	viper.SetEnvPrefix("GOAGENT")
	viper.AutomaticEnv()

	// bind specific env vars
	viper.BindEnv("openrouter.api_key", "OPENROUTER_API_KEY")

	// read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// validate required fields
	if config.OpenRouter.APIKey == "" {
		return nil, fmt.Errorf("OpenRouter API key is required. Set OPENROUTER_API_KEY environment variable or add to config file")
	}

	return &config, nil
}

func (c *Config) CreateDefaultConfigFile() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(home, ".agent_go.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return nil // File already exists
	}

	defaultConfig := `# Agent_Go Configuration
openrouter:
  api_key: "your-openrouter-api-key-here"
  model: "anthropic/claude-3.5-sonnet"
  base_url: "https://openrouter.ai/api/v1"

agent:
  confirm_destructive: true
  max_tokens: 4000
  temperature: 0.7
`

	return os.WriteFile(configPath, []byte(defaultConfig), 0644)
}
