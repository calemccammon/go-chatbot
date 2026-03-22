package config

import (
	"errors"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	defaultBaseURL       = "https://api.openai.com/v1"
	defaultModel         = "gpt-4o-mini"
	defaultSystemPrompt  = "You are a helpful assistant for a desktop chat app."
	defaultTimeoutSecond = 90
)

var (
	// ErrMissingAPIKey is returned when LLM_API_KEY is not set.
	ErrMissingAPIKey = errors.New("LLM_API_KEY is required")
)

// Config stores runtime values used to initialize the chatbot application.
type Config struct {
	APIKey         string
	BaseURL        string
	Model          string
	SystemPrompt   string
	RequestTimeout time.Duration
}

// LoadFromEnv reads configuration from environment variables and validates values.
func LoadFromEnv() (Config, error) {
	cfg := Config{
		APIKey:         strings.TrimSpace(os.Getenv("LLM_API_KEY")),
		BaseURL:        getenvDefault("LLM_BASE_URL", defaultBaseURL),
		Model:          getenvDefault("LLM_MODEL", defaultModel),
		SystemPrompt:   getenvDefault("CHAT_SYSTEM_PROMPT", defaultSystemPrompt),
		RequestTimeout: defaultTimeoutSecond * time.Second,
	}

	if cfg.APIKey == "" {
		return Config{}, ErrMissingAPIKey
	}
	if _, err := url.ParseRequestURI(cfg.BaseURL); err != nil {
		return Config{}, errors.New("LLM_BASE_URL must be a valid URL")
	}

	return cfg, nil
}

func getenvDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
