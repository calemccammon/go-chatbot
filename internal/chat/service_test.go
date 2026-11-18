package chat

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"go-chatbot/internal/llm"
)

var errBoom = errors.New("boom")

// mockCompleter returns a canned reply or error and records the history it was given.
type mockCompleter struct {
	reply string
	err   error
	got   []llm.Message
}

func (m *mockCompleter) Chat(_ context.Context, history []llm.Message) (string, error) {
	m.got = history
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
				if !errors.Is(err, tt.wantErr) {
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
			completer: &mockCompleter{err: errBoom},
			wantErr:   errBoom,
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
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
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

func TestSendPassesHistoryToClient(t *testing.T) {
	completer := &mockCompleter{reply: "first reply"}
	svc, err := NewService(completer, "be brief")
	if err != nil {
		t.Fatalf("unexpected setup error: %v", err)
	}

	if _, err := svc.Send(context.Background(), " hello "); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []llm.Message{
		{Role: "system", Content: "be brief"},
		{Role: "user", Content: "hello"},
	}
	if !reflect.DeepEqual(completer.got, want) {
		t.Fatalf("first request: expected %+v, got %+v", want, completer.got)
	}

	if _, err := svc.Send(context.Background(), "again"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want = []llm.Message{
		{Role: "system", Content: "be brief"},
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "first reply"},
		{Role: "user", Content: "again"},
	}
	if !reflect.DeepEqual(completer.got, want) {
		t.Fatalf("second request: expected %+v, got %+v", want, completer.got)
	}
}

func TestSendFailureLeavesHistoryUnchanged(t *testing.T) {
	completer := &mockCompleter{err: errBoom}
	svc, err := NewService(completer, "be brief")
	if err != nil {
		t.Fatalf("unexpected setup error: %v", err)
	}

	if _, err := svc.Send(context.Background(), "hello"); !errors.Is(err, errBoom) {
		t.Fatalf("expected %v, got %v", errBoom, err)
	}
	history := svc.History()
	if len(history) != 1 || history[0].Role != "system" {
		t.Fatalf("expected only the system prompt after a failed send, got %+v", history)
	}

	completer.err = nil
	completer.reply = "ok"
	if _, err := svc.Send(context.Background(), "retry"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []llm.Message{
		{Role: "system", Content: "be brief"},
		{Role: "user", Content: "retry"},
	}
	if !reflect.DeepEqual(completer.got, want) {
		t.Fatalf("retry request: expected %+v, got %+v", want, completer.got)
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
