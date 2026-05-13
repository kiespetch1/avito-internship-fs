package llm

import (
	"context"
	"errors"
)

var ErrProviderFailed = errors.New("llm provider failed")

type Request struct {
	Model        string
	SystemPrompt string
	UserPrompt   string
}

type Response struct {
	Output       string
	TokensIn     int
	TokensOut    int
	LatencyMs    int
	FinishReason string
}

type StreamChunk struct {
	Delta string
}

type Provider interface {
	Generate(ctx context.Context, req Request) (Response, error)
	GenerateStream(ctx context.Context, req Request, onChunk func(StreamChunk)) (Response, error)
}
