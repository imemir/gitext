package commands

import (
	"fmt"

	"github.com/imemir/gitext/pkg/aiconfig"
	"github.com/imemir/gitext/pkg/ui"
	"github.com/spf13/cobra"
)

func NewAIConfigCmd(opts *Options) *cobra.Command {
	var test bool

	cmd := &cobra.Command{
		Use:   "config",
		Short: "View or edit AI configuration",
		Long: `Display current AI configuration or test the connection.
Use 'gitext ai setup' to reconfigure.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			output := ui.NewOutput(opts.Verbose)
			aiOutput := ui.NewAIOutput(opts.Verbose)

			manager, err := aiconfig.NewManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			if !manager.Exists() {
				output.Warning("AI configuration not found")
				output.Next("run 'gitext ai setup' to configure")
				return nil
			}

			cfg, err := manager.Load()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Display configuration
			output.Info("Current AI Configuration:")
			fmt.Println()
			fmt.Printf("  Provider: %s\n", cfg.Provider)

			if cfg.Provider == "openai" {
				fmt.Printf("  API Key: %s\n", aiconfig.MaskAPIKey(cfg.OpenAI.APIKey))
				fmt.Printf("  Model: %s\n", cfg.OpenAI.Model)
			} else {
				fmt.Printf("  API Key: %s\n", aiconfig.MaskAPIKey(cfg.OpenRouter.APIKey))
				fmt.Printf("  Model: %s\n", cfg.OpenRouter.Model)
				if cfg.OpenRouter.UseFreeModel {
					fmt.Printf("  Using free model: Yes\n")
				} else {
					fmt.Printf("  Using free model: No\n")
				}
			}

			configPath, err := aiconfig.GetConfigPath()
			if err != nil {
				return err
			}
			fmt.Printf("  Config file: %s\n", configPath)
			fmt.Println()

			// Test connection if requested
			if test {
				output.Info("Testing connection...")
				if err := testAIConnection(cfg, aiOutput); err != nil {
					return fmt.Errorf("connection test failed: %w", err)
				}
			} else {
				output.Next("run 'gitext ai config --test' to test the connection")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&test, "test", false, "Test the AI connection")

	return cmd
}
