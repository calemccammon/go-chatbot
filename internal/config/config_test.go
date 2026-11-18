package config

import (
	"errors"
	"testing"
)

func TestLoadFromEnv(t *testing.T) {
	tests := []struct {
		name      string
		env       map[string]string
		wantModel string
		wantURL   string
		wantErr   error
	}{
		{
			name: "uses defaults",
			env: map[string]string{
				"LLM_API_KEY": "key",
			},
			wantModel: "gpt-4o-mini",
			wantURL:   "https://api.openai.com/v1",
		},
		{
			name: "uses provided values",
			env: map[string]string{
				"LLM_API_KEY":        "key",
				"LLM_MODEL":          "custom-model",
				"LLM_BASE_URL":       "https://example.com/v1",
				"CHAT_SYSTEM_PROMPT": "Prompt",
			},
			wantModel: "custom-model",
			wantURL:   "https://example.com/v1",
		},
		{
			name: "missing api key",
			env: map[string]string{
				"LLM_API_KEY": "",
			},
			wantErr: ErrMissingAPIKey,
		},
		{
			name: "invalid base url",
			env: map[string]string{
				"LLM_API_KEY":  "key",
				"LLM_BASE_URL": "://bad",
			},
			wantErr: ErrInvalidBaseURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LLM_API_KEY", "")
			t.Setenv("LLM_MODEL", "")
			t.Setenv("LLM_BASE_URL", "")
			t.Setenv("CHAT_SYSTEM_PROMPT", "")

			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg, err := LoadFromEnv()
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Model != tt.wantModel {
				t.Fatalf("expected model %q, got %q", tt.wantModel, cfg.Model)
			}
			if cfg.BaseURL != tt.wantURL {
				t.Fatalf("expected base URL %q, got %q", tt.wantURL, cfg.BaseURL)
			}
		})
	}
}
