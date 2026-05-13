package llm

import (
	"bufio"
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
	httpResp, err := p.doChatCompletionRequest(ctx, req, false, "application/json")
	if err != nil {
		return Response{}, err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(httpResp.Body, 1<<20))
	if err != nil {
		return Response{}, fmt.Errorf("%w: read response: %v", ErrProviderFailed, err)
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

func (p *OpenAICompatibleProvider) GenerateStream(ctx context.Context, req Request, onChunk func(StreamChunk)) (Response, error) {
	start := time.Now()
	httpResp, err := p.doChatCompletionRequest(ctx, req, true, "text/event-stream")
	if err != nil {
		return Response{}, err
	}
	defer httpResp.Body.Close()

	var output strings.Builder
	finishReason := "unknown"
	seenDone := false
	scanner := bufio.NewScanner(httpResp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	dataLines := make([]string, 0, 1)

	flushEvent := func() (bool, error) {
		if len(dataLines) == 0 {
			return false, nil
		}
		data := strings.Join(dataLines, "\n")
		dataLines = dataLines[:0]
		if data == "[DONE]" {
			seenDone = true
			return true, nil
		}

		var parsed chatCompletionStreamResponse
		if err := json.Unmarshal([]byte(data), &parsed); err != nil {
			return false, fmt.Errorf("%w: decode stream chunk: %v", ErrProviderFailed, err)
		}
		for _, choice := range parsed.Choices {
			if choice.Delta.Content != "" {
				output.WriteString(choice.Delta.Content)
				onChunk(StreamChunk{Delta: choice.Delta.Content})
			}
			if choice.FinishReason != "" {
				finishReason = choice.FinishReason
			}
		}

		return false, nil
	}

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			done, err := flushEvent()
			if err != nil {
				return Response{}, err
			}
			if done {
				break
			}

			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		if data, ok := strings.CutPrefix(line, "data:"); ok {
			dataLines = append(dataLines, strings.TrimPrefix(data, " "))
		}
	}
	if err := scanner.Err(); err != nil {
		return Response{}, fmt.Errorf("%w: read stream: %v", ErrProviderFailed, err)
	}
	if len(dataLines) > 0 {
		if _, err := flushEvent(); err != nil {
			return Response{}, err
		}
	}
	if !seenDone {
		return Response{}, fmt.Errorf("%w: stream ended before done event", ErrProviderFailed)
	}

	text := strings.TrimSpace(output.String())
	if text == "" {
		return Response{}, fmt.Errorf("%w: response content is empty", ErrProviderFailed)
	}

	return Response{
		Output:       text,
		TokensIn:     approximateTokens(req.SystemPrompt) + approximateTokens(req.UserPrompt),
		TokensOut:    approximateTokens(text),
		LatencyMs:    int(time.Since(start) / time.Millisecond),
		FinishReason: finishReason,
	}, nil
}

func (p *OpenAICompatibleProvider) doChatCompletionRequest(ctx context.Context, req Request, stream bool, accept string) (*http.Response, error) {
	body, err := p.chatCompletionRequestBody(req, stream)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrProviderFailed, err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", accept)

	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: request failed: %v", ErrProviderFailed, err)
	}
	if httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
		return httpResp, nil
	}

	defer httpResp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(httpResp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("%w: read response: %v", ErrProviderFailed, err)
	}

	return nil, fmt.Errorf("%w: status %d: %s", ErrProviderFailed, httpResp.StatusCode, providerErrorMessage(respBody))
}

func (p *OpenAICompatibleProvider) chatCompletionRequestBody(req Request, stream bool) ([]byte, error) {
	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = p.defaultModel
	}
	if model == "" {
		return nil, fmt.Errorf("%w: model is required", ErrProviderFailed)
	}

	body, err := json.Marshal(chatCompletionRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: req.SystemPrompt},
			{Role: "user", Content: req.UserPrompt},
		},
		Stream: stream,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: encode request: %v", ErrProviderFailed, err)
	}

	return body, nil
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

type chatCompletionStreamResponse struct {
	Choices []struct {
		FinishReason string `json:"finish_reason"`
		Delta        struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}
