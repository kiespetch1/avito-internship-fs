package llm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAICompatibleProviderSendsChatCompletionRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method: %s", r.Method)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("authorization: %q", got)
		}

		var body chatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if body.Model != "openai/gpt-4o-mini" {
			t.Fatalf("model: %q", body.Model)
		}
		if body.Stream {
			t.Fatalf("stream must be false")
		}
		if len(body.Messages) != 2 {
			t.Fatalf("messages: %+v", body.Messages)
		}
		if body.Messages[0] != (chatMessage{Role: "system", Content: "sys"}) {
			t.Fatalf("system message: %+v", body.Messages[0])
		}
		if body.Messages[1] != (chatMessage{Role: "user", Content: "hello"}) {
			t.Fatalf("user message: %+v", body.Messages[1])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices":[{"finish_reason":"stop","message":{"content":"answer"}}],
			"usage":{"prompt_tokens":7,"completion_tokens":3}
		}`))
	}))
	defer server.Close()

	provider, err := NewOpenAICompatibleProvider(OpenAICompatibleConfig{
		BaseURL: server.URL + "/v1",
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("provider: %v", err)
	}

	resp, err := provider.Generate(context.Background(), Request{
		Model:        "openai/gpt-4o-mini",
		SystemPrompt: "sys",
		UserPrompt:   "hello",
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if resp.Output != "answer" || resp.TokensIn != 7 || resp.TokensOut != 3 || resp.FinishReason != "stop" {
		t.Fatalf("response: %+v", resp)
	}
}

func TestOpenAICompatibleProviderRequiresModel(t *testing.T) {
	provider, err := NewOpenAICompatibleProvider(OpenAICompatibleConfig{
		BaseURL: "https://example.test",
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("provider: %v", err)
	}
	_, err = provider.Generate(context.Background(), Request{UserPrompt: "hello"})
	if !errors.Is(err, ErrProviderFailed) {
		t.Fatalf("expected ErrProviderFailed, got %v", err)
	}
	if !strings.Contains(err.Error(), "model is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenAICompatibleProviderStreamsChatCompletion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "text/event-stream" {
			t.Fatalf("accept: %q", got)
		}
		var body chatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !body.Stream {
			t.Fatal("stream must be true")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"hel\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"lo\"},\"finish_reason\":\"stop\"}]}\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	provider, err := NewOpenAICompatibleProvider(OpenAICompatibleConfig{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("provider: %v", err)
	}

	chunks := make([]string, 0, 2)
	resp, err := provider.GenerateStream(context.Background(), Request{Model: "m", UserPrompt: "hello"}, func(chunk StreamChunk) {
		chunks = append(chunks, chunk.Delta)
	})
	if err != nil {
		t.Fatalf("generate stream: %v", err)
	}
	if strings.Join(chunks, "") != "hello" {
		t.Fatalf("chunks: %+v", chunks)
	}
	if resp.Output != "hello" || resp.FinishReason != "stop" {
		t.Fatalf("response: %+v", resp)
	}
}

func TestOpenAICompatibleProviderRejectsPrematureStreamClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"partial\"}}]}\n\n"))
	}))
	defer server.Close()

	provider, err := NewOpenAICompatibleProvider(OpenAICompatibleConfig{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("provider: %v", err)
	}

	_, err = provider.GenerateStream(context.Background(), Request{Model: "m", UserPrompt: "hello"}, func(StreamChunk) {})
	if !errors.Is(err, ErrProviderFailed) {
		t.Fatalf("expected provider error, got %v", err)
	}
}

func TestOpenAICompatibleProviderReturnsProviderErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"code":"invalid_key","message":"bad key"}}`))
	}))
	defer server.Close()

	provider, err := NewOpenAICompatibleProvider(OpenAICompatibleConfig{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("provider: %v", err)
	}
	_, err = provider.Generate(context.Background(), Request{Model: "m", UserPrompt: "hello"})
	if !errors.Is(err, ErrProviderFailed) {
		t.Fatalf("expected ErrProviderFailed, got %v", err)
	}
	if !strings.Contains(err.Error(), "status 401") || !strings.Contains(err.Error(), "invalid_key: bad key") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenAICompatibleProviderRejectsInvalidConfig(t *testing.T) {
	if _, err := NewOpenAICompatibleProvider(OpenAICompatibleConfig{BaseURL: ":", APIKey: "key"}); err == nil {
		t.Fatal("expected invalid base url error")
	}
	if _, err := NewOpenAICompatibleProvider(OpenAICompatibleConfig{BaseURL: "https://example.test"}); err == nil {
		t.Fatal("expected missing api key error")
	}
}

func TestOpenAICompatibleProviderRejectsInvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	defer server.Close()

	provider, err := NewOpenAICompatibleProvider(OpenAICompatibleConfig{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})
	if err != nil {
		t.Fatalf("provider: %v", err)
	}
	_, err = provider.Generate(context.Background(), Request{Model: "m", UserPrompt: "hello"})
	if !errors.Is(err, ErrProviderFailed) {
		t.Fatalf("expected ErrProviderFailed, got %v", err)
	}
}
