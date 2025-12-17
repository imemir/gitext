package ai

// Provider defines the interface for AI providers
type Provider interface {
	// GenerateCommitMessage generates a commit message based on the git diff
	// The message should follow Conventional Commits format: type(scope): description
	GenerateCommitMessage(diff string) (string, error)
	
	// Name returns the name of the provider
	Name() string
}

// Model represents an AI model configuration
type Model struct {
	ID          string
	Name        string
	Description string
	IsFree      bool
}
