package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the .gitext configuration file
type Config struct {
	Branch struct {
		Production string `yaml:"production"`
		Stage      string `yaml:"stage"`
	} `yaml:"branch"`
	Naming struct {
		Feature string `yaml:"feature"`
		Hotfix  string `yaml:"hotfix"`
	} `yaml:"naming"`
	Merge struct {
		RequireRetargetForProdFromStage bool `yaml:"requireRetargetForProdFromStage"`
	} `yaml:"merge"`
	CI struct {
		Stage      []string `yaml:"stage"`
		Production []string `yaml:"production"`
	} `yaml:"ci"`
	PR struct {
		TemplatePath string `yaml:"templatePath"`
	} `yaml:"pr"`
	Remote struct {
		Name string `yaml:"name"`
	} `yaml:"remote"`
}

// Load loads the .gitext configuration file from the repository root
// It walks up from the current directory to find the git root
func Load() (*Config, error) {
	gitRoot, err := findGitRoot()
	if err != nil {
		return nil, fmt.Errorf("not in a git repository: %w", err)
	}

	configPath := filepath.Join(gitRoot, ".gitext")
	config := &Config{}

	// Set defaults
	config.Branch.Production = DefaultProductionBranch
	config.Branch.Stage = DefaultStageBranch
	config.Remote.Name = DefaultRemoteName
	config.Naming.Feature = DefaultFeaturePattern
	config.Naming.Hotfix = DefaultHotfixPattern

	// Load config file if it exists
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read .gitext: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse .gitext: %w", err)
		}
	}

	// Apply defaults for missing values
	if config.Branch.Production == "" {
		config.Branch.Production = DefaultProductionBranch
	}
	if config.Branch.Stage == "" {
		config.Branch.Stage = DefaultStageBranch
	}
	if config.Remote.Name == "" {
		config.Remote.Name = DefaultRemoteName
	}
	if config.Naming.Feature == "" {
		config.Naming.Feature = DefaultFeaturePattern
	}
	if config.Naming.Hotfix == "" {
		config.Naming.Hotfix = DefaultHotfixPattern
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate validates the configuration values
func (c *Config) Validate() error {
	if c.Branch.Production == "" {
		return fmt.Errorf("branch.production cannot be empty")
	}
	if c.Branch.Stage == "" {
		return fmt.Errorf("branch.stage cannot be empty")
	}
	if c.Branch.Production == c.Branch.Stage {
		return fmt.Errorf("branch.production and branch.stage must be different")
	}
	if c.Remote.Name == "" {
		return fmt.Errorf("remote.name cannot be empty")
	}
	return nil
}

// GetGitRoot returns the git repository root directory
func GetGitRoot() (string, error) {
	return findGitRoot()
}

// findGitRoot walks up from the current directory to find .git
func findGitRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not in a git repository")
		}
		dir = parent
	}
}

// Save writes the configuration to .gitext in the repository root
func (c *Config) Save() error {
	gitRoot, err := findGitRoot()
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	configPath := filepath.Join(gitRoot, ".gitext")
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configPath, data, 0644)
}

