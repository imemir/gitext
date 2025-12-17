package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "gitext-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check defaults
	if cfg.Branch.Production != DefaultProductionBranch {
		t.Errorf("Expected production branch %s, got %s", DefaultProductionBranch, cfg.Branch.Production)
	}
	if cfg.Branch.Stage != DefaultStageBranch {
		t.Errorf("Expected stage branch %s, got %s", DefaultStageBranch, cfg.Branch.Stage)
	}
	if cfg.Remote.Name != DefaultRemoteName {
		t.Errorf("Expected remote name %s, got %s", DefaultRemoteName, cfg.Remote.Name)
	}
}

func TestLoadWithConfig(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "gitext-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Create .gitext file
	configContent := `branch:
  production: "main"
  stage: "develop"
remote:
  name: "upstream"
`
	configPath := filepath.Join(tmpDir, ".gitext")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check loaded values
	if cfg.Branch.Production != "main" {
		t.Errorf("Expected production branch 'main', got %s", cfg.Branch.Production)
	}
	if cfg.Branch.Stage != "develop" {
		t.Errorf("Expected stage branch 'develop', got %s", cfg.Branch.Stage)
	}
	if cfg.Remote.Name != "upstream" {
		t.Errorf("Expected remote name 'upstream', got %s", cfg.Remote.Name)
	}
}

func TestValidate(t *testing.T) {
	cfg := &Config{}
	cfg.Branch.Production = "production"
	cfg.Branch.Stage = "stage"
	cfg.Remote.Name = "origin"

	if err := cfg.Validate(); err != nil {
		t.Errorf("Valid config should not error: %v", err)
	}

	// Test empty production branch
	cfg.Branch.Production = ""
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for empty production branch")
	}

	// Test same branches
	cfg.Branch.Production = "main"
	cfg.Branch.Stage = "main"
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for same production and stage branches")
	}
}

