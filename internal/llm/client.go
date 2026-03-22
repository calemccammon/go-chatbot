package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Message represents a single role/content item in a chat conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the request payload for chat completions endpoints.
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type openAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Client sends chat completion requests to an OpenAI-compatible endpoint.
type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	model      string
}

// NewClient builds an OpenAI-compatible chat client from validated settings.
func NewClient(apiKey, baseURL, model string) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("LLM_API_KEY is required")
	}
	if strings.TrimSpace(baseURL) == "" {
		return nil, errors.New("LLM_BASE_URL is required")
	}
	if strings.TrimSpace(model) == "" {
		return nil, errors.New("LLM_MODEL is required")
	}

	return &Client{
		httpClient: &http.Client{Timeout: 60 * time.Second},
		apiKey:     apiKey,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		model:      model,
	}, nil
}

// Chat sends the conversation history and returns the first assistant response.
func (c *Client) Chat(ctx context.Context, history []Message) (string, error) {
	payload := ChatRequest{
		Model:    c.model,
		Messages: history,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 300 {
		var apiErr openAIErrorResponse
		if jsonErr := json.Unmarshal(respBody, &apiErr); jsonErr == nil && apiErr.Error.Message != "" {
			return "", fmt.Errorf("llm error (%d): %s", resp.StatusCode, apiErr.Error.Message)
		}
		return "", fmt.Errorf("llm error (%d): %s", resp.StatusCode, string(respBody))
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("no choices returned by model")
	}

	return parsed.Choices[0].Message.Content, nil
}
