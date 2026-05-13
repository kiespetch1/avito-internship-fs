package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"avito-internship-fs/internal/config"
	"avito-internship-fs/internal/llm"
)

func TestResolveLLMProviderMock(t *testing.T) {
	p, err := resolveLLMProvider(config.LLMConfig{Provider: "mock"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if _, ok := p.(*llm.MockProvider); !ok {
		t.Fatalf("expected *llm.MockProvider, got %T", p)
	}
}

func TestResolveLLMProviderEmptyDefaultsToMock(t *testing.T) {
	p, err := resolveLLMProvider(config.LLMConfig{Provider: ""})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if _, ok := p.(*llm.MockProvider); !ok {
		t.Fatalf("empty provider must default to mock, got %T", p)
	}
}

func TestResolveLLMProviderRejectsUnknown(t *testing.T) {
	_, err := resolveLLMProvider(config.LLMConfig{Provider: "unknown"})
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestResolveLLMProviderOpenAICompatible(t *testing.T) {
	p, err := resolveLLMProvider(config.LLMConfig{
		Provider: "openai-compatible",
		APIKey:   "test-key",
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if _, ok := p.(*llm.OpenAICompatibleProvider); !ok {
		t.Fatalf("expected *llm.OpenAICompatibleProvider, got %T", p)
	}
}

func TestResolveLLMProviderOpenAICompatibleRequiresAPIKey(t *testing.T) {
	_, err := resolveLLMProvider(config.LLMConfig{Provider: "openai-compatible"})
	if err == nil {
		t.Fatal("expected error without api key")
	}
}

func TestHealthcheckReturns200JSON(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_info", nil)
	healthcheck(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type: %q", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status field: %q", body["status"])
	}
	if body["time"] == "" {
		t.Fatalf("expected time field")
	}
}

func TestOpenAPISpecReturnsRootAPIYaml(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/docs/openapi.yaml", nil)
	openAPISpec(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/yaml; charset=utf-8" {
		t.Fatalf("content-type: %q", ct)
	}
	if body := rr.Body.String(); body == "" || !strings.Contains(body, "openapi: 3.0.3") {
		t.Fatalf("unexpected openapi body prefix: %.80q", body)
	}
}

func TestSwaggerUIReferencesOpenAPISpec(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	swaggerUI(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("content-type: %q", ct)
	}
	if body := rr.Body.String(); !strings.Contains(body, "/docs/openapi.yaml") || !strings.Contains(body, "SwaggerUIBundle") {
		t.Fatalf("swagger ui does not reference spec: %.120q", body)
	}
}

func TestWithCORSHandlesPreflight(t *testing.T) {
	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("inner handler must not run for OPTIONS preflight")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/anything", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Headers", "X-Custom-Header")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status: %d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("allow-origin: %q", rr.Header().Get("Access-Control-Allow-Origin"))
	}
	if rr.Header().Get("Access-Control-Allow-Headers") != "X-Custom-Header" {
		t.Fatalf("allow-headers: %q", rr.Header().Get("Access-Control-Allow-Headers"))
	}
}

func TestWithCORSFallsBackToDefaultRequestHeaders(t *testing.T) {
	handler := withCORS(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	req := httptest.NewRequest(http.MethodOptions, "/x", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Headers"); got != "Authorization, Content-Type" {
		t.Fatalf("default allow-headers: %q", got)
	}
}

func TestWithCORSPassesThroughNonOptions(t *testing.T) {
	called := false
	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	}))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("inner handler must run for non-OPTIONS")
	}
	if rr.Code != http.StatusTeapot {
		t.Fatalf("status: %d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("allow-origin must be set on non-preflight too")
	}
}

func TestWithCORSWithoutOriginSetsNoHeaders(t *testing.T) {
	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("must not set CORS headers without Origin")
	}
}
