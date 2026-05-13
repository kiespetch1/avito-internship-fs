package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
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
	ReapPending(ctx context.Context, olderThan time.Duration, reason string) (int64, error)
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
	runs      RunRepo
	provider  llm.Provider
	timeout   time.Duration
	lifecycle context.Context
	active    sync.WaitGroup
}

// NewRunService строит сервис запусков. lifecycle — серверный контекст, отменяемый при shutdown:
// он используется для синхронного Run, чтобы LLM-вызов пережил обрыв соединения клиента,
// но не висел вечно при остановке сервиса.
func NewRunService(lifecycle context.Context, runs RunRepo, provider llm.Provider, timeout time.Duration) *RunService {
	if lifecycle == nil {
		lifecycle = context.Background()
	}

	return &RunService{
		runs:      runs,
		provider:  provider,
		timeout:   timeout,
		lifecycle: lifecycle,
	}
}

// Run запускает ассистента через LLM-провайдер. Возвращает run в финальном состоянии (success или failed),
// а при ошибке провайдера — обёрнутую llm.ErrProviderFailed, чтобы handler выбрал нужный HTTP-статус.
// LLM-вызов отвязан от request ctx, чтобы запуск всегда дошёл до терминального статуса даже при обрыве клиента,
// но привязан к lifecycle, чтобы graceful shutdown корректно прерывал зависшие запросы к провайдеру.
func (s *RunService) Run(_ context.Context, assistantID, userID uuid.UUID, userPrompt string) (domain.AssistantRun, error) {
	s.active.Add(1)
	defer s.active.Done()

	assistant, run, err := s.prepareRun(assistantID, userID, userPrompt)
	if err != nil {
		return domain.AssistantRun{}, err
	}

	llmCtx, llmCancel := context.WithTimeout(s.lifecycle, s.timeout)
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

// RunStream запускает ассистента в стриминговом режиме. LLM-вызов привязан к lifecycle,
// а не к request ctx: даже если клиент отключится, провайдер дочитает поток до конца и run
// получит терминальный статус. Колбэки клиенту прекращаются по обрыву ctx, но запуск не страдает.
func (s *RunService) RunStream(ctx context.Context, assistantID, userID uuid.UUID, userPrompt string, callbacks RunStreamCallbacks) (domain.AssistantRun, error) {
	s.active.Add(1)
	defer s.active.Done()

	assistant, run, err := s.prepareRun(assistantID, userID, userPrompt)
	if err != nil {
		return domain.AssistantRun{}, err
	}
	if callbacks.OnRunCreated != nil {
		callbacks.OnRunCreated(run)
	}

	llmCtx, llmCancel := context.WithTimeout(s.lifecycle, s.timeout)
	defer llmCancel()

	start := time.Now()
	resp, llmErr := s.provider.GenerateStream(llmCtx, llm.Request{
		Model:        assistant.Model,
		SystemPrompt: assistant.SystemPrompt,
		UserPrompt:   userPrompt,
	}, func(chunk llm.StreamChunk) {
		if callbacks.OnDelta == nil || chunk.Delta == "" {
			return
		}
		if ctx.Err() != nil {
			return
		}
		callbacks.OnDelta(chunk.Delta)
	})
	latencyMs := int(time.Since(start) / time.Millisecond)

	return s.completeRun(run, resp, llmErr, latencyMs)
}

// Shutdown ждёт завершения активных запусков или истечения ctx, что наступит раньше.
// Должен вызываться после http.Server.Shutdown и до отмены lifecycle context'а,
// чтобы LLM-вызовы успели дописать терминальный статус в БД.
func (s *RunService) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		s.active.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ReapStalePending помечает зависшие pending-раны как failed. Запускается на старте сервиса,
// чтобы подчистить запуски, которые не дожили до терминала из-за crash/SIGKILL/OOM.
// olderThan должен быть строго больше LLM-таймаута.
func (s *RunService) ReapStalePending(ctx context.Context, olderThan time.Duration, reason string) (int64, error) {
	return s.runs.ReapPending(ctx, olderThan, reason)
}

func (s *RunService) prepareRun(assistantID, userID uuid.UUID, userPrompt string) (domain.Assistant, domain.AssistantRun, error) {
	dbCtx, dbCancel := context.WithTimeout(s.lifecycle, 5*time.Second)
	defer dbCancel()

	assistant, run, err := s.runs.CreatePendingForActiveAssistant(dbCtx, assistantID, userID, userPrompt)
	if err != nil {
		return domain.Assistant{}, domain.AssistantRun{}, err
	}

	return assistant, run, nil
}

func (s *RunService) completeRun(run domain.AssistantRun, resp llm.Response, llmErr error, latencyMs int) (domain.AssistantRun, error) {
	finalCtx, finalCancel := context.WithTimeout(s.lifecycle, 5*time.Second)
	defer finalCancel()

	if llmErr != nil {
		wrapped := fmt.Errorf("%w: %w", llm.ErrProviderFailed, llmErr)
		if err := s.runs.MarkFailed(finalCtx, run.ID, llmErr.Error(), latencyMs); err != nil {
			slog.Error("mark run failed", "run_id", run.ID, "error", err)
		}
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
