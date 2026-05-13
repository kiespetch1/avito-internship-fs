package llm

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (m *MockProvider) Generate(ctx context.Context, req Request) (Response, error) {
	start := time.Now()

	select {
	case <-ctx.Done():
		return Response{}, ctx.Err()
	default:
	}

	output := fmt.Sprintf("[mock:%s] %s", req.Model, req.UserPrompt)

	return Response{
		Output:       output,
		TokensIn:     approximateTokens(req.SystemPrompt) + approximateTokens(req.UserPrompt),
		TokensOut:    approximateTokens(output),
		LatencyMs:    int(time.Since(start) / time.Millisecond),
		FinishReason: "stop",
	}, nil
}

func (m *MockProvider) GenerateStream(ctx context.Context, req Request, onChunk func(StreamChunk)) (Response, error) {
	start := time.Now()

	select {
	case <-ctx.Done():
		return Response{}, ctx.Err()
	default:
	}

	output := fmt.Sprintf("[mock:%s] %s", req.Model, req.UserPrompt)
	parts := strings.Fields(output)
	if len(parts) == 0 {
		parts = []string{output}
	}
	for i, part := range parts {
		select {
		case <-ctx.Done():
			return Response{}, ctx.Err()
		default:
		}
		if i > 0 {
			part = " " + part
		}
		onChunk(StreamChunk{Delta: part})
	}

	return Response{
		Output:       output,
		TokensIn:     approximateTokens(req.SystemPrompt) + approximateTokens(req.UserPrompt),
		TokensOut:    approximateTokens(output),
		LatencyMs:    int(time.Since(start) / time.Millisecond),
		FinishReason: "stop",
	}, nil
}

func approximateTokens(s string) int {
	if s == "" {
		return 0
	}
	n := len(s) / 4
	if n == 0 {
		return 1
	}

	return n
}
