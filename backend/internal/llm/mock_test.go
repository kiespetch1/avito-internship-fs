package llm

import (
	"context"
	"strings"
	"testing"
)

func TestMockProviderDeterministic(t *testing.T) {
	p := NewMockProvider()
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
	p := NewMockProvider()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.Generate(ctx, Request{Model: "m", UserPrompt: "x"})
	if err == nil {
		t.Fatalf("expected error from cancelled context")
	}
}
