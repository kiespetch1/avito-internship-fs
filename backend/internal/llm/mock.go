package llm

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const defaultMockProviderErrorRate = 0.05

var ErrMockProviderFailure = errors.New("mock provider failure")

type MockProvider struct {
	errorRate float64
	rand      *rand.Rand
}

func NewMockProvider() *MockProvider {
	return NewMockProviderWithErrorRate(defaultMockProviderErrorRate)
}

func NewStableMockProvider() *MockProvider {
	return NewMockProviderWithErrorRate(0)
}

func NewMockProviderWithErrorRate(errorRate float64) *MockProvider {
	return NewMockProviderWithRand(
		errorRate,
		rand.New(rand.NewSource(time.Now().UnixNano())),
	)
}

func NewMockProviderWithRand(errorRate float64, r *rand.Rand) *MockProvider {
	if r == nil {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	return &MockProvider{
		errorRate: errorRate,
		rand:      r,
	}
}

func (m *MockProvider) Generate(ctx context.Context, req Request) (Response, error) {
	start := time.Now()

	select {
	case <-ctx.Done():
		return Response{}, ctx.Err()
	default:
	}

	if m.shouldFail() {
		return Response{}, ErrMockProviderFailure
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

	if m.shouldFail() {
		return Response{}, ErrMockProviderFailure
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

func (m *MockProvider) shouldFail() bool {
	if m.errorRate <= 0 {
		return false
	}
	if m.errorRate >= 1 {
		return true
	}

	return m.rand.Float64() < m.errorRate
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
