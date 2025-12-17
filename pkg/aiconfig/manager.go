package aiconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Manager handles loading and saving AI configuration
type Manager struct {
	configPath string
}

// NewManager creates a new Manager instance
func NewManager() (*Manager, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}
	return &Manager{configPath: configPath}, nil
}

// Load loads the AI configuration from ~/.gitext/config.yaml
func (m *Manager) Load() (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("AI configuration not found. Run 'gitext ai setup' to configure")
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults for missing values
	if cfg.Provider == "" {
		cfg.Provider = "openai"
	}
	if cfg.OpenAI.Model == "" {
		cfg.OpenAI.Model = "gpt-4o"
	}
	if cfg.OpenRouter.Model == "" {
		cfg.OpenRouter.Model = "google/gemini-flash-1.5-8b"
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Save saves the AI configuration to ~/.gitext/config.yaml
func (m *Manager) Save(cfg *Config) error {
	// Validate before saving
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Ensure config directory exists
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write config file with restricted permissions (0600)
	if err := os.WriteFile(m.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Exists checks if the config file exists
func (m *Manager) Exists() bool {
	_, err := os.Stat(m.configPath)
	return err == nil
}
