package ai

import (
	"fmt"

	"github.com/imemir/gitext/pkg/aiconfig"
)

// Service manages AI providers and generates commit messages
type Service struct {
	provider Provider
	config   *aiconfig.Config
}

// NewService creates a new AI service from configuration
func NewService(cfg *aiconfig.Config) (*Service, error) {
	var provider Provider

	switch cfg.Provider {
	case "openai":
		provider = NewOpenAIProvider(cfg.OpenAI.APIKey, cfg.OpenAI.Model)
	case "openrouter":
		model := cfg.OpenRouter.Model
		if cfg.OpenRouter.UseFreeModel && model == "" {
			model = FreeModels[0].ID
		}
		provider = NewOpenRouterProvider(cfg.OpenRouter.APIKey, model, cfg.OpenRouter.UseFreeModel)
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Provider)
	}

	return &Service{
		provider: provider,
		config:   cfg,
	}, nil
}

// GenerateCommitMessage generates a commit message from a git diff
func (s *Service) GenerateCommitMessage(diff string) (string, error) {
	if diff == "" {
		return "", fmt.Errorf("diff is empty")
	}

	message, err := s.provider.GenerateCommitMessage(diff)
	if err != nil {
		return "", fmt.Errorf("failed to generate commit message: %w", err)
	}

	return message, nil
}

// GetProviderName returns the name of the current provider
func (s *Service) GetProviderName() string {
	return s.provider.Name()
}
