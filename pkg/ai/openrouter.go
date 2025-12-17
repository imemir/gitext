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
	openRouterAPIURL = "https://openrouter.ai/api/v1/chat/completions"
	openRouterTimeout = 30 * time.Second
)

// Free models available on OpenRouter
var FreeModels = []Model{
	{
		ID:          "google/gemini-flash-1.5-8b",
		Name:        "Gemini Flash 1.5 8B",
		Description: "Fast and efficient model from Google",
		IsFree:      true,
	},
	{
		ID:          "qwen/qwen-2.5-7b-instruct",
		Name:        "Qwen 2.5 7B Instruct",
		Description: "High-quality instruction-following model",
		IsFree:      true,
	},
	{
		ID:          "mistralai/mistral-7b-instruct-v0.2",
		Name:        "Mistral 7B Instruct",
		Description: "Balanced performance model from Mistral",
		IsFree:      true,
	},
}

// OpenRouterProvider implements the Provider interface for OpenRouter
type OpenRouterProvider struct {
	apiKey      string
	model       string
	useFreeModel bool
	client      *http.Client
}

// NewOpenRouterProvider creates a new OpenRouter provider
func NewOpenRouterProvider(apiKey, model string, useFreeModel bool) *OpenRouterProvider {
	if model == "" {
		model = FreeModels[0].ID
	}
	return &OpenRouterProvider{
		apiKey:       apiKey,
		model:        model,
		useFreeModel: useFreeModel,
		client: &http.Client{
			Timeout: openRouterTimeout,
		},
	}
}

// Name returns the provider name
func (p *OpenRouterProvider) Name() string {
	return "OpenRouter"
}

// GenerateCommitMessage generates a commit message using OpenRouter
func (p *OpenRouterProvider) GenerateCommitMessage(diff string) (string, error) {
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

	req, err := http.NewRequest("POST", openRouterAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/gitext/gitext")
	req.Header.Set("X-Title", "gitext")

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
			return "", fmt.Errorf("OpenRouter API error: %s", errorResp.Error.Message)
		}
		return "", fmt.Errorf("OpenRouter API error: status %d, body: %s", resp.StatusCode, string(body))
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
