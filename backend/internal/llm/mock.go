package llm

import (
	"context"
	"fmt"
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
