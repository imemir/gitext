package ui

import (
	"fmt"
	"os"
)

// Output provides consistent output formatting
type Output struct {
	verbose bool
}

// NewOutput creates a new Output instance
func NewOutput(verbose bool) *Output {
	return &Output{verbose: verbose}
}

// Info prints an info message
func (o *Output) Info(format string, args ...interface{}) {
	fmt.Printf("ℹ  %s\n", fmt.Sprintf(format, args...))
}

// Success prints a success message
func (o *Output) Success(format string, args ...interface{}) {
	fmt.Printf("✓  %s\n", fmt.Sprintf(format, args...))
}

// Warning prints a warning message
func (o *Output) Warning(format string, args ...interface{}) {
	fmt.Printf("⚠  %s\n", fmt.Sprintf(format, args...))
}

// Error prints an error message
func (o *Output) Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "✗  %s\n", fmt.Sprintf(format, args...))
}

// Doing prints what is about to be done
func (o *Output) Doing(format string, args ...interface{}) {
	fmt.Printf("→  %s\n", fmt.Sprintf(format, args...))
}

// Did prints what was done
func (o *Output) Did(format string, args ...interface{}) {
	fmt.Printf("✓  %s\n", fmt.Sprintf(format, args...))
}

// Next prints the next recommended command
func (o *Output) Next(format string, args ...interface{}) {
	fmt.Printf("→  Next: %s\n", fmt.Sprintf(format, args...))
}

// Verbose prints a message only if verbose mode is enabled
func (o *Output) Verbose(format string, args ...interface{}) {
	if o.verbose {
		fmt.Printf("   %s\n", fmt.Sprintf(format, args...))
	}
}

// Print prints a plain message
func (o *Output) Print(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

