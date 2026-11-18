package llm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testTimeout = 5 * time.Second

func TestClientChatSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("expected path /v1/chat/completions, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer key" {
			t.Errorf("expected Authorization %q, got %q", "Bearer key", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", got)
		}

		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if req.Model != "model" {
			t.Errorf("expected model %q, got %q", "model", req.Model)
		}
		want := Message{Role: "user", Content: "hi"}
		if len(req.Messages) != 1 || req.Messages[0] != want {
			t.Errorf("expected messages [%+v], got %+v", want, req.Messages)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"hello"}}]}`))
	}))
	defer server.Close()

	// The trailing slash checks that the base URL is normalised before the path is appended.
	client, err := NewClient("key", server.URL+"/v1/", "model", testTimeout)
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"bad key"}}`))
	}))
	defer server.Close()

	client, err := NewClient("key", server.URL, "model", testTimeout)
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

func TestClientChatNonJSONError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("upstream down"))
	}))
	defer server.Close()

	client, err := NewClient("key", server.URL, "model", testTimeout)
	if err != nil {
		t.Fatalf("unexpected setup error: %v", err)
	}

	_, err = client.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); got != "llm error (502): upstream down" {
		t.Fatalf("expected raw body in error message, got %q", got)
	}
}

func TestClientChatNoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	defer server.Close()

	client, err := NewClient("key", server.URL, "model", testTimeout)
	if err != nil {
		t.Fatalf("unexpected setup error: %v", err)
	}

	_, err = client.Chat(context.Background(), []Message{{Role: "user", Content: "hi"}})
	if !errors.Is(err, ErrNoChoices) {
		t.Fatalf("expected %v, got %v", ErrNoChoices, err)
	}
}
