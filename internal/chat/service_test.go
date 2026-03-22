package chat

import (
	"context"
	"errors"
	"testing"

	"go-chatbot/internal/llm"
)

type mockCompleter struct {
	reply string
	err   error
}

func (m *mockCompleter) Chat(_ context.Context, _ []llm.Message) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.reply, nil
}

func TestNewService(t *testing.T) {
	tests := []struct {
		name    string
		client  Completer
		prompt  string
		wantErr error
	}{
		{name: "nil client", client: nil, wantErr: ErrNilClient},
		{name: "valid client", client: &mockCompleter{reply: "hi"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewService(tt.client, tt.prompt)
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestSend(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		completer *mockCompleter
		wantReply string
		wantErr   error
	}{
		{
			name:      "empty message",
			input:     "   ",
			completer: &mockCompleter{reply: "ok"},
			wantErr:   ErrEmptyMessage,
		},
		{
			name:      "upstream error",
			input:     "hello",
			completer: &mockCompleter{err: errors.New("boom")},
			wantErr:   errors.New("boom"),
		},
		{
			name:      "success",
			input:     "hello",
			completer: &mockCompleter{reply: " world "},
			wantReply: "world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewService(tt.completer, "")
			if err != nil {
				t.Fatalf("unexpected setup error: %v", err)
			}

			reply, err := svc.Send(context.Background(), tt.input)
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if reply != tt.wantReply {
				t.Fatalf("expected reply %q, got %q", tt.wantReply, reply)
			}
		})
	}
}

func TestClear(t *testing.T) {
	svc, err := NewService(&mockCompleter{reply: "ok"}, "")
	if err != nil {
		t.Fatalf("unexpected setup error: %v", err)
	}

	_, _ = svc.Send(context.Background(), "hello")
	svc.Clear()

	history := svc.History()
	if len(history) != 1 {
		t.Fatalf("expected history length 1, got %d", len(history))
	}
	if history[0].Role != "system" {
		t.Fatalf("expected first message to be system, got %q", history[0].Role)
	}
}
