package chat

import (
	"context"
	"errors"
	"strings"

	"go-chatbot/internal/llm"
)

const defaultSystemPrompt = "You are a helpful assistant for a desktop chat app."

var (
	// ErrNilClient is returned when a nil completer is provided.
	ErrNilClient = errors.New("chat client is required")
	// ErrEmptyMessage is returned when a blank user message is submitted.
	ErrEmptyMessage = errors.New("message cannot be empty")
)

// Completer defines the behavior required from an LLM chat completion client.
type Completer interface {
	Chat(ctx context.Context, history []llm.Message) (string, error)
}

// Service manages chat history and sends requests to an LLM client.
type Service struct {
	client  Completer
	history []llm.Message
}

// NewService creates a chat service with an initialized system prompt.
func NewService(client Completer, systemPrompt string) (*Service, error) {
	if client == nil {
		return nil, ErrNilClient
	}

	prompt := strings.TrimSpace(systemPrompt)
	if prompt == "" {
		prompt = defaultSystemPrompt
	}

	return &Service{
		client: client,
		history: []llm.Message{{
			Role:    "system",
			Content: prompt,
		}},
	}, nil
}

// Send appends a user prompt, requests a response, and stores assistant output.
func (s *Service) Send(ctx context.Context, prompt string) (string, error) {
	trimmed := strings.TrimSpace(prompt)
	if trimmed == "" {
		return "", ErrEmptyMessage
	}

	s.history = append(s.history, llm.Message{Role: "user", Content: trimmed})
	historySnapshot := append([]llm.Message(nil), s.history...)

	reply, err := s.client.Chat(ctx, historySnapshot)
	if err != nil {
		return "", err
	}

	s.history = append(s.history, llm.Message{Role: "assistant", Content: strings.TrimSpace(reply)})
	return strings.TrimSpace(reply), nil
}

// Clear removes all messages except the initial system prompt.
func (s *Service) Clear() {
	if len(s.history) == 0 {
		return
	}
	s.history = s.history[:1]
}

// History returns a copy of the current message history.
func (s *Service) History() []llm.Message {
	return append([]llm.Message(nil), s.history...)
}
