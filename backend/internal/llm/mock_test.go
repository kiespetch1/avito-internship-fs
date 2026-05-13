package llm

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestMockProviderDeterministic(t *testing.T) {
	p := NewStableMockProvider()
	req := Request{Model: "gpt-mock", SystemPrompt: "sys", UserPrompt: "hello"}

	r1, err := p.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	r2, err := p.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if r1.Output != r2.Output {
		t.Fatalf("non-deterministic output: %q vs %q", r1.Output, r2.Output)
	}

	if !strings.Contains(r1.Output, "gpt-mock") || !strings.Contains(r1.Output, "hello") {
		t.Fatalf("unexpected output: %q", r1.Output)
	}

	if r1.TokensIn == 0 || r1.TokensOut == 0 {
		t.Fatalf("token counts must be set: %+v", r1)
	}

	if r1.FinishReason != "stop" {
		t.Fatalf("finish reason: %q", r1.FinishReason)
	}
}

func TestMockProviderRespectsCancelledContext(t *testing.T) {
	p := NewMockProviderWithErrorRate(1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.Generate(ctx, Request{Model: "m", UserPrompt: "x"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestMockProviderCanReturnError(t *testing.T) {
	p := NewMockProviderWithErrorRate(1)

	_, err := p.Generate(context.Background(), Request{
		Model:      "m",
		UserPrompt: "x",
	})

	if !errors.Is(err, ErrMockProviderFailure) {
		t.Fatalf("expected ErrMockProviderFailure, got %v", err)
	}
}

func TestStableMockProviderDoesNotReturnRandomError(t *testing.T) {
	p := NewStableMockProvider()

	for i := 0; i < 100; i++ {
		_, err := p.Generate(context.Background(), Request{
			Model:      "m",
			UserPrompt: "x",
		})
		if err != nil {
			t.Fatalf("stable mock should not fail, got %v", err)
		}
	}
}
