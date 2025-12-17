package aiconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents the AI configuration stored in ~/.gitext/config.yaml
type Config struct {
	Provider string `yaml:"provider"` // "openai" or "openrouter"
	OpenAI   struct {
		APIKey string `yaml:"api_key"`
		Model  string `yaml:"model"` // default: "gpt-4o"
	} `yaml:"openai"`
	OpenRouter struct {
		APIKey     string `yaml:"api_key"`
		Model      string `yaml:"model"`
		UseFreeModel bool `yaml:"use_free_model"` // if true, use predefined free models
	} `yaml:"openrouter"`
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	cfg := &Config{}
	cfg.Provider = "openai"
	cfg.OpenAI.Model = "gpt-4o"
	cfg.OpenRouter.UseFreeModel = true
	cfg.OpenRouter.Model = "google/gemini-flash-1.5-8b"
	return cfg
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Provider != "openai" && c.Provider != "openrouter" {
		return fmt.Errorf("provider must be 'openai' or 'openrouter', got: %s", c.Provider)
	}

	if c.Provider == "openai" {
		if c.OpenAI.APIKey == "" {
			return fmt.Errorf("openai.api_key is required")
		}
		if !strings.HasPrefix(c.OpenAI.APIKey, "sk-") {
			return fmt.Errorf("openai.api_key must start with 'sk-'")
		}
		if c.OpenAI.Model == "" {
			c.OpenAI.Model = "gpt-4o"
		}
	}

	if c.Provider == "openrouter" {
		if c.OpenRouter.APIKey == "" {
			return fmt.Errorf("openrouter.api_key is required")
		}
		if !strings.HasPrefix(c.OpenRouter.APIKey, "sk-") {
			return fmt.Errorf("openrouter.api_key must start with 'sk-'")
		}
		if c.OpenRouter.Model == "" {
			c.OpenRouter.Model = "google/gemini-flash-1.5-8b"
		}
	}

	return nil
}

// GetConfigDir returns the directory where AI config is stored (~/.gitext)
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".gitext"), nil
}

// GetConfigPath returns the path to the AI config file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.yaml"), nil
}

// MaskAPIKey masks an API key for display (shows first 7 chars + ****)
func MaskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 7 {
		return "****"
	}
	return key[:7] + "****"
}
