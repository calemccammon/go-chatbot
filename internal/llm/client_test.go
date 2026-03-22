package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientChatSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"hello"}}]}`))
	}))
	defer server.Close()

	client, err := NewClient("key", server.URL, "model")
	if err != nil {
		t.Fatalf("unexpected setup error: %v", err)
	}

	reply, err := client.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply != "hello" {
		t.Fatalf("expected hello, got %q", reply)
	}
}

func TestClientChatAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error":{"message":"bad key"}}`))
	}))
	defer server.Close()

	client, err := NewClient("key", server.URL, "model")
	if err != nil {
		t.Fatalf("unexpected setup error: %v", err)
	}

	_, err = client.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "llm error (401): bad key" {
		t.Fatalf("expected API error message, got %q", got)
	}
}

func TestClientChatNoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	defer server.Close()

	client, err := NewClient("key", server.URL, "model")
	if err != nil {
		t.Fatalf("unexpected setup error: %v", err)
	}

	_, err = client.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "no choices returned by model" {
		t.Fatalf("expected no choices error, got %q", got)
	}
}
