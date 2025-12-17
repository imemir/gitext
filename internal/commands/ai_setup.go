package commands

import (
	"fmt"

	"github.com/imemir/gitext/pkg/ai"
	"github.com/imemir/gitext/pkg/aiconfig"
	"github.com/imemir/gitext/pkg/ui"
	"github.com/spf13/cobra"
)

func NewAISetupCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup AI provider for commit message generation",
		Long: `Interactive setup for configuring AI provider (OpenAI or OpenRouter) 
for automatic commit message generation. This will create a configuration file
at ~/.gitext/config.yaml with your API keys and model preferences.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			output := ui.NewOutput(opts.Verbose)
			aiOutput := ui.NewAIOutput(opts.Verbose)

			// Check if config already exists
			manager, err := aiconfig.NewManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			if manager.Exists() {
				output.Warning("AI configuration already exists")
				confirmed, err := ui.PromptConfirm("Do you want to overwrite it?", false)
				if err != nil {
					return err
				}
				if !confirmed {
					output.Info("Setup cancelled")
					return nil
				}
			}

			cfg := aiconfig.DefaultConfig()

			// Select provider
			output.Info("Select AI Provider:")
			providerOptions := []struct {
				Label       string
				Description string
			}{
				{"OpenAI", "Official OpenAI API (requires API key)"},
				{"OpenRouter", "OpenRouter API (supports free models)"},
			}

			providerIdx, err := ui.PromptSelectWithDescriptions("", providerOptions)
			if err != nil {
				return fmt.Errorf("failed to select provider: %w", err)
			}

			if providerIdx == 0 {
				cfg.Provider = "openai"
			} else {
				cfg.Provider = "openrouter"
			}

			// Get API key
			apiKeyPrompt := fmt.Sprintf("Enter your %s API key: ", cfg.Provider)
			apiKey, err := ui.PromptPassword(apiKeyPrompt)
			if err != nil {
				return fmt.Errorf("failed to read API key: %w", err)
			}
			if apiKey == "" {
				return fmt.Errorf("API key cannot be empty")
			}

			// Configure provider-specific settings
			if cfg.Provider == "openai" {
				cfg.OpenAI.APIKey = apiKey

				// Select model
				output.Info("Select OpenAI model:")
				modelOptions := []string{
					"gpt-4o (default - recommended)",
					"gpt-4o-mini (faster, cheaper)",
					"gpt-4-turbo",
					"gpt-3.5-turbo (cheapest)",
					"Custom model",
				}

				modelIdx, err := ui.PromptSelect("", modelOptions)
				if err != nil {
					return fmt.Errorf("failed to select model: %w", err)
				}

				switch modelIdx {
				case 0:
					cfg.OpenAI.Model = "gpt-4o"
				case 1:
					cfg.OpenAI.Model = "gpt-4o-mini"
				case 2:
					cfg.OpenAI.Model = "gpt-4-turbo"
				case 3:
					cfg.OpenAI.Model = "gpt-3.5-turbo"
				case 4:
					customModel, err := ui.PromptInput("Enter custom model name: ")
					if err != nil {
						return fmt.Errorf("failed to read model name: %w", err)
					}
					if customModel == "" {
						return fmt.Errorf("model name cannot be empty")
					}
					cfg.OpenAI.Model = customModel
				}
			} else {
				cfg.OpenRouter.APIKey = apiKey

				// Select model type
				output.Info("Select OpenRouter model option:")
				modelTypeOptions := []struct {
					Label       string
					Description string
				}{
					{"Use free model", "Select from free models (recommended)"},
					{"Use custom model", "Enter your own model name"},
				}

				modelTypeIdx, err := ui.PromptSelectWithDescriptions("", modelTypeOptions)
				if err != nil {
					return fmt.Errorf("failed to select model type: %w", err)
				}

				if modelTypeIdx == 0 {
					cfg.OpenRouter.UseFreeModel = true
					output.Info("Select free model:")
					freeModelOptions := make([]string, len(ai.FreeModels))
					for i, model := range ai.FreeModels {
						freeModelOptions[i] = fmt.Sprintf("%s - %s", model.Name, model.Description)
					}

					freeModelIdx, err := ui.PromptSelect("", freeModelOptions)
					if err != nil {
						return fmt.Errorf("failed to select free model: %w", err)
					}
					cfg.OpenRouter.Model = ai.FreeModels[freeModelIdx].ID
				} else {
					cfg.OpenRouter.UseFreeModel = false
					customModel, err := ui.PromptInput("Enter custom model name (e.g., anthropic/claude-3-opus): ")
					if err != nil {
						return fmt.Errorf("failed to read model name: %w", err)
					}
					if customModel == "" {
						return fmt.Errorf("model name cannot be empty")
					}
					cfg.OpenRouter.Model = customModel
				}
			}

			// Test connection
			output.Info("Testing connection...")
			if err := testAIConnection(cfg, aiOutput); err != nil {
				output.Error("Connection test failed: %v", err)
				confirmed, err := ui.PromptConfirm("Do you want to save the configuration anyway?", false)
				if err != nil {
					return err
				}
				if !confirmed {
					output.Info("Setup cancelled")
					return nil
				}
			}

			// Save configuration
			output.Doing("Saving configuration")
			if err := manager.Save(cfg); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			configPath, err := aiconfig.GetConfigPath()
			if err != nil {
				return err
			}

			output.Success("Configuration saved to %s", configPath)
			output.Next("run 'gitext commit' to generate commit messages with AI")

			return nil
		},
	}

	return cmd
}

// testAIConnection tests the AI connection with a simple prompt
func testAIConnection(cfg *aiconfig.Config, aiOutput *ui.AIOutput) error {
	aiOutput.TestingConnection(cfg.Provider)

	service, err := ai.NewService(cfg)
	if err != nil {
		return fmt.Errorf("failed to create AI service: %w", err)
	}

	// Test with a simple diff
	testDiff := `diff --git a/test.go b/test.go
index 1234567..abcdefg 100644
--- a/test.go
+++ b/test.go
@@ -1,3 +1,5 @@
 package main

+func newFunction() {
+}
`

	_, err = service.GenerateCommitMessage(testDiff)
	if err != nil {
		return err
	}

	aiOutput.ConnectionSuccess(cfg.Provider)
	return nil
}
