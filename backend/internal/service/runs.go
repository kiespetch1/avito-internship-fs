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

type RunRepo interface {
	CreatePendingForActiveAssistant(ctx context.Context, assistantID, userID uuid.UUID, userPrompt string) (domain.Assistant, domain.AssistantRun, error)
	MarkSuccess(ctx context.Context, id uuid.UUID, output string, tokensIn, tokensOut, latencyMs int, finishReason string) error
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string, latencyMs int) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.AssistantRun, error)
	UpsertFeedback(ctx context.Context, runID uuid.UUID, rating domain.RunFeedbackRating) (domain.AssistantRun, error)
	List(ctx context.Context, f repository.RunListFilter) ([]domain.AssistantRun, int, error)
}

type RunListInput struct {
	UserID      *uuid.UUID
	AssistantID *uuid.UUID
	Status      *domain.RunStatus
	Page        int
	PageSize    int
}

type RunStreamCallbacks struct {
	OnRunCreated func(domain.AssistantRun)
	OnDelta      func(string)
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

func (s *RunService) SetFeedback(ctx context.Context, runID, userID uuid.UUID, rating domain.RunFeedbackRating) (domain.AssistantRun, error) {
	run, err := s.runs.GetByID(ctx, runID)
	if err != nil {
		return domain.AssistantRun{}, err
	}
	if run.UserID != userID {
		return domain.AssistantRun{}, domain.ErrRunForbidden
	}

	return s.runs.UpsertFeedback(ctx, runID, rating)
}

type RunService struct {
	runs     RunRepo
	provider llm.Provider
	timeout  time.Duration
}

func NewRunService(runs RunRepo, provider llm.Provider, timeout time.Duration) *RunService {
	return &RunService{
		runs:     runs,
		provider: provider,
		timeout:  timeout,
	}
}

// Run запускает ассистента через LLM-провайдер. Возвращает run в финальном состоянии (success или failed),
// а при ошибке провайдера — обёрнутую llm.ErrProviderFailed, чтобы handler выбрал нужный HTTP-статус
func (s *RunService) Run(ctx context.Context, assistantID, userID uuid.UUID, userPrompt string) (domain.AssistantRun, error) {
	assistant, run, err := s.prepareRun(ctx, assistantID, userID, userPrompt)
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

	return s.completeRun(run, resp, llmErr, latencyMs)
}

func (s *RunService) RunStream(ctx context.Context, assistantID, userID uuid.UUID, userPrompt string, callbacks RunStreamCallbacks) (domain.AssistantRun, error) {
	assistant, run, err := s.prepareRun(ctx, assistantID, userID, userPrompt)
	if err != nil {
		return domain.AssistantRun{}, err
	}
	if callbacks.OnRunCreated != nil {
		callbacks.OnRunCreated(run)
	}

	llmCtx, llmCancel := context.WithTimeout(ctx, s.timeout)
	defer llmCancel()

	start := time.Now()
	resp, llmErr := s.provider.GenerateStream(llmCtx, llm.Request{
		Model:        assistant.Model,
		SystemPrompt: assistant.SystemPrompt,
		UserPrompt:   userPrompt,
	}, func(chunk llm.StreamChunk) {
		if callbacks.OnDelta != nil && chunk.Delta != "" {
			callbacks.OnDelta(chunk.Delta)
		}
	})
	latencyMs := int(time.Since(start) / time.Millisecond)

	return s.completeRun(run, resp, llmErr, latencyMs)
}

func (s *RunService) prepareRun(ctx context.Context, assistantID, userID uuid.UUID, userPrompt string) (domain.Assistant, domain.AssistantRun, error) {
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dbCancel()

	assistant, run, err := s.runs.CreatePendingForActiveAssistant(dbCtx, assistantID, userID, userPrompt)
	if err != nil {
		return domain.Assistant{}, domain.AssistantRun{}, err
	}

	return assistant, run, nil
}

func (s *RunService) completeRun(run domain.AssistantRun, resp llm.Response, llmErr error, latencyMs int) (domain.AssistantRun, error) {
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

	if err := s.runs.MarkSuccess(
		finalCtx, run.ID, resp.Output, resp.TokensIn, resp.TokensOut, latencyMs, resp.FinishReason); err != nil {
		return run, err
	}
	updated, err := s.runs.GetByID(finalCtx, run.ID)
	if err != nil {
		return run, err
	}

	return updated, nil
}
