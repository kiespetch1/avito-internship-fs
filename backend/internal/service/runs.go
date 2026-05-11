package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"avito-internship-fs/internal/domain"
	"avito-internship-fs/internal/llm"
	"avito-internship-fs/internal/repository"
)

type AssistantRepo interface {
	Get(ctx context.Context, id uuid.UUID) (domain.Assistant, error)
}

type RunRepo interface {
	CreatePending(ctx context.Context, assistantID, userID uuid.UUID, model, userPrompt string) (domain.AssistantRun, error)
	MarkSuccess(ctx context.Context, id uuid.UUID, output string, tokensIn, tokensOut, latencyMs int, finishReason string) error
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string, latencyMs int) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.AssistantRun, error)
	List(ctx context.Context, f repository.RunListFilter) ([]domain.AssistantRun, int, error)
}

type RunListInput struct {
	UserID      *uuid.UUID
	AssistantID *uuid.UUID
	Status      *domain.RunStatus
	Page        int
	PageSize    int
}

func (s *RunService) List(ctx context.Context, in RunListInput) ([]domain.AssistantRun, int, error) {
	page, pageSize := normalizePage(in.Page, in.PageSize, defaultRunsPageSize, maxPageSize)

	return s.runs.List(ctx, repository.RunListFilter{
		UserID:      in.UserID,
		AssistantID: in.AssistantID,
		Status:      in.Status,
		Limit:       pageSize,
		Offset:      (page - 1) * pageSize,
	})
}

type RunService struct {
	assistants AssistantRepo
	runs       RunRepo
	provider   llm.Provider
	timeout    time.Duration
}

func NewRunService(assistants AssistantRepo, runs RunRepo, provider llm.Provider, timeout time.Duration) *RunService {
	return &RunService{
		assistants: assistants,
		runs:       runs,
		provider:   provider,
		timeout:    timeout,
	}
}

// Run запускает ассистента через LLM-провайдер. Возвращает run в финальном состоянии (success или failed),
// а при ошибке провайдера — обёрнутую llm.ErrProviderFailed, чтобы handler выбрал нужный HTTP-статус
func (s *RunService) Run(ctx context.Context, assistantID, userID uuid.UUID, userPrompt string) (domain.AssistantRun, error) {
	assistant, err := s.assistants.Get(ctx, assistantID)
	if err != nil {
		return domain.AssistantRun{}, err
	}
	if !assistant.IsActive {
		return domain.AssistantRun{}, domain.ErrAssistantInactive
	}

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dbCancel()

	run, err := s.runs.CreatePending(dbCtx, assistantID, userID, assistant.Model, userPrompt)
	if err != nil {
		return domain.AssistantRun{}, err
	}

	llmCtx, llmCancel := context.WithTimeout(context.Background(), s.timeout)
	defer llmCancel()

	start := time.Now()
	resp, llmErr := s.provider.Generate(llmCtx, llm.Request{
		Model:        assistant.Model,
		SystemPrompt: assistant.SystemPrompt,
		UserPrompt:   userPrompt,
	})
	latencyMs := int(time.Since(start) / time.Millisecond)

	finalCtx, finalCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer finalCancel()

	if llmErr != nil {
		wrapped := fmt.Errorf("%w: %w", llm.ErrProviderFailed, llmErr)
		_ = s.runs.MarkFailed(finalCtx, run.ID, llmErr.Error(), latencyMs)
		updated, getErr := s.runs.GetByID(finalCtx, run.ID)
		if getErr != nil {
			return run, wrapped
		}

		return updated, wrapped
	}

	err = s.runs.MarkSuccess(
		finalCtx, run.ID, resp.Output, resp.TokensIn, resp.TokensOut, latencyMs, resp.FinishReason)
	if err != nil {
		return run, err
	}
	updated, err := s.runs.GetByID(finalCtx, run.ID)
	if err != nil {
		return run, err
	}

	return updated, nil
}
