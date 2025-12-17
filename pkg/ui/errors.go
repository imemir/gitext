package ui

import (
	"fmt"
)

// ErrorWithSuggestion represents an error with a suggested fix
type ErrorWithSuggestion struct {
	Message    string
	Suggestion string
}

func (e *ErrorWithSuggestion) Error() string {
	if e.Suggestion != "" {
		return fmt.Sprintf("%s → %s", e.Message, e.Suggestion)
	}
	return e.Message
}

// NewError creates an error with a suggestion
func NewError(message, suggestion string) error {
	return &ErrorWithSuggestion{
		Message:    message,
		Suggestion: suggestion,
	}
}

// FormatError formats an error for display
func FormatError(err error) string {
	if errWithSuggestion, ok := err.(*ErrorWithSuggestion); ok {
		return errWithSuggestion.Error()
	}
	return err.Error()
}

