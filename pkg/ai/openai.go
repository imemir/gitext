package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	openAIAPIURL = "https://api.openai.com/v1/chat/completions"
	openAITimeout = 30 * time.Second
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	apiKey string
	model  string
	client *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	if model == "" {
		model = "gpt-4o"
	}
	return &OpenAIProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{
			Timeout: openAITimeout,
		},
	}
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "OpenAI"
}

// GenerateCommitMessage generates a commit message using OpenAI
func (p *OpenAIProvider) GenerateCommitMessage(diff string) (string, error) {
	prompt := `You are a git commit message generator. Analyze the following git diff and generate a commit message following the Conventional Commits specification (https://www.conventionalcommits.org/en/v1.0.0/).

The commit message format should be:
type(scope): description

Where:
- type: feat, fix, docs, style, refactor, perf, test, chore, etc.
- scope: optional, the area affected (e.g., auth, api, ui)
- description: brief summary in imperative mood

Rules:
- Use lowercase for the type
- Use imperative mood for description (e.g., "add feature" not "added feature")
- Keep description concise (max 72 characters)
- If there are breaking changes, add "!" after type or "BREAKING CHANGE:" in body

Git diff:
` + diff + `

Generate ONLY the commit message header (type(scope): description), nothing else.`

	requestBody := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
		"max_tokens": 100,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", openAIAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return "", fmt.Errorf("OpenAI API error: %s", errorResp.Error.Message)
		}
		return "", fmt.Errorf("OpenAI API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	message := response.Choices[0].Message.Content
	// Clean up the message (remove quotes, trim whitespace)
	message = trimMessage(message)

	return message, nil
}
