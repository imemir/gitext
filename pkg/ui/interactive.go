package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// PromptInput prompts the user for input and returns the entered string
func PromptInput(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// PromptPassword prompts the user for a password (hidden input)
func PromptPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	
	// Read password with hidden input
	fd := int(os.Stdin.Fd())
	bytePassword, err := term.ReadPassword(fd)
	if err != nil {
		return "", err
	}
	
	fmt.Println() // New line after hidden input
	return string(bytePassword), nil
}

// PromptSelect prompts the user to select from a list of options
func PromptSelect(prompt string, options []string) (int, error) {
	if len(options) == 0 {
		return -1, fmt.Errorf("no options provided")
	}

	fmt.Println(prompt)
	for i, option := range options {
		fmt.Printf("  %d) %s\n", i+1, option)
	}
	fmt.Print("Select (1-" + fmt.Sprintf("%d", len(options)) + "): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return -1, err
	}

	input = strings.TrimSpace(input)
	var choice int
	if _, err := fmt.Sscanf(input, "%d", &choice); err != nil {
		return -1, fmt.Errorf("invalid selection")
	}

	if choice < 1 || choice > len(options) {
		return -1, fmt.Errorf("selection out of range")
	}

	return choice - 1, nil
}

// PromptConfirm prompts the user for yes/no confirmation
func PromptConfirm(prompt string, defaultValue bool) (bool, error) {
	defaultText := "y/N"
	if defaultValue {
		defaultText = "Y/n"
	}

	fmt.Printf("%s [%s]: ", prompt, defaultText)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return defaultValue, nil
	}

	return input == "y" || input == "yes", nil
}

// PromptSelectWithDescriptions prompts the user to select from options with descriptions
func PromptSelectWithDescriptions(prompt string, options []struct {
	Label       string
	Description string
}) (int, error) {
	if len(options) == 0 {
		return -1, fmt.Errorf("no options provided")
	}

	fmt.Println(prompt)
	for i, option := range options {
		fmt.Printf("  %d) %s\n", i+1, option.Label)
		if option.Description != "" {
			fmt.Printf("     %s\n", option.Description)
		}
	}
	fmt.Print("Select (1-" + fmt.Sprintf("%d", len(options)) + "): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return -1, err
	}

	input = strings.TrimSpace(input)
	var choice int
	if _, err := fmt.Sscanf(input, "%d", &choice); err != nil {
		return -1, fmt.Errorf("invalid selection")
	}

	if choice < 1 || choice > len(options) {
		return -1, fmt.Errorf("selection out of range")
	}

	return choice - 1, nil
}
