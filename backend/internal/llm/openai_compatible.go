package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultOpenAIBaseURL = "https://api.openai.com/v1"
)

type OpenAICompatibleConfig struct {
	BaseURL      string
	APIKey       string
	DefaultModel string
}

type OpenAICompatibleProvider struct {
	endpoint     string
	apiKey       string
	defaultModel string
	client       *http.Client
}

func NewOpenAICompatibleProvider(cfg OpenAICompatibleConfig) (*OpenAICompatibleProvider, error) {
	endpoint, err := chatCompletionsEndpoint(cfg.BaseURL)
	if err != nil {
		return nil, err
	}
	apiKey := strings.TrimSpace(cfg.APIKey)
	if apiKey == "" {
		return nil, errors.New("llm api key is required for openai-compatible provider")
	}

	return &OpenAICompatibleProvider{
		endpoint:     endpoint,
		apiKey:       apiKey,
		defaultModel: strings.TrimSpace(cfg.DefaultModel),
		client:       http.DefaultClient,
	}, nil
}

func (p *OpenAICompatibleProvider) Generate(ctx context.Context, req Request) (Response, error) {
	start := time.Now()
	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = p.defaultModel
	}
	if model == "" {
		return Response{}, fmt.Errorf("%w: model is required", ErrProviderFailed)
	}

	body, err := json.Marshal(chatCompletionRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: req.SystemPrompt},
			{Role: "user", Content: req.UserPrompt},
		},
		Stream: false,
	})
	if err != nil {
		return Response{}, fmt.Errorf("%w: encode request: %v", ErrProviderFailed, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("%w: build request: %v", ErrProviderFailed, err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("%w: request failed: %v", ErrProviderFailed, err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(httpResp.Body, 1<<20))
	if err != nil {
		return Response{}, fmt.Errorf("%w: read response: %v", ErrProviderFailed, err)
	}
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return Response{}, fmt.Errorf("%w: status %d: %s", ErrProviderFailed, httpResp.StatusCode, providerErrorMessage(respBody))
	}

	var parsed chatCompletionResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return Response{}, fmt.Errorf("%w: decode response: %v", ErrProviderFailed, err)
	}
	if len(parsed.Choices) == 0 {
		return Response{}, fmt.Errorf("%w: response contains no choices", ErrProviderFailed)
	}

	output := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if output == "" {
		return Response{}, fmt.Errorf("%w: response content is empty", ErrProviderFailed)
	}

	tokensIn := approximateTokens(req.SystemPrompt) + approximateTokens(req.UserPrompt)
	tokensOut := approximateTokens(output)
	if parsed.Usage != nil {
		if parsed.Usage.PromptTokens > 0 {
			tokensIn = parsed.Usage.PromptTokens
		}
		if parsed.Usage.CompletionTokens > 0 {
			tokensOut = parsed.Usage.CompletionTokens
		}
	}

	finishReason := parsed.Choices[0].FinishReason
	if finishReason == "" {
		finishReason = "unknown"
	}

	return Response{
		Output:       output,
		TokensIn:     tokensIn,
		TokensOut:    tokensOut,
		LatencyMs:    int(time.Since(start) / time.Millisecond),
		FinishReason: finishReason,
	}, nil
}

func chatCompletionsEndpoint(baseURL string) (string, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return "", errors.New("llm base url is required for openai-compatible provider")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid llm base url: %q", baseURL)
	}
	if strings.HasSuffix(baseURL, "/chat/completions") {
		return baseURL, nil
	}

	return baseURL + "/chat/completions", nil
}

func providerErrorMessage(body []byte) string {
	var parsed struct {
		Error struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Error.Message != "" {
		if parsed.Error.Code != "" {
			return parsed.Error.Code + ": " + parsed.Error.Message
		}

		return parsed.Error.Message
	}
	msg := strings.TrimSpace(string(body))
	if msg == "" {
		return "empty error response"
	}
	if len(msg) > 500 {
		return msg[:500] + "..."
	}

	return msg
}

type chatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}
