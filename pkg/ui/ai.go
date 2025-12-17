package ui

import (
	"fmt"
	"strings"
)

// AIOutput provides AI-specific output formatting
type AIOutput struct {
	*Output
}

// NewAIOutput creates a new AIOutput instance
func NewAIOutput(verbose bool) *AIOutput {
	return &AIOutput{
		Output: NewOutput(verbose),
	}
}

// GeneratingCommitMessage shows that we're generating a commit message
func (o *AIOutput) GeneratingCommitMessage() {
	o.Doing("Generating commit message with AI...")
}

// CommitMessageGenerated displays the generated commit message
func (o *AIOutput) CommitMessageGenerated(message string) {
	o.Success("Generated commit message:")
	fmt.Println()
	fmt.Println("  " + message)
	fmt.Println()
}

// TestingConnection shows that we're testing the API connection
func (o *AIOutput) TestingConnection(provider string) {
	o.Doing("Testing connection to %s...", provider)
}

// ConnectionSuccess shows successful connection test
func (o *AIOutput) ConnectionSuccess(provider string) {
	o.Success("Successfully connected to %s", provider)
}

// ConnectionFailed shows failed connection test
func (o *AIOutput) ConnectionFailed(provider string, err error) {
	o.Error("Failed to connect to %s: %v", provider, err)
}

// FormatCommitMessage formats a commit message for display
func FormatCommitMessage(message string) string {
	lines := strings.Split(message, "\n")
	if len(lines) == 0 {
		return message
	}
	
	// Highlight the first line (header)
	header := lines[0]
	body := strings.Join(lines[1:], "\n")
	
	if body != "" {
		return fmt.Sprintf("%s\n%s", header, body)
	}
	return header
}
