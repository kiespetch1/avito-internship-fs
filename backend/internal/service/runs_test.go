package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"avito-internship-fs/internal/domain"
	"avito-internship-fs/internal/llm"
	"avito-internship-fs/internal/repository"
)

type fakeRunRepo struct {
	pendingAssistant *domain.Assistant
	createPendingErr error
	created          []domain.AssistantRun
	successCall      *successArgs
	failedCall       *failedArgs
	feedbackCall     *feedbackArgs
	store            map[uuid.UUID]domain.AssistantRun
}

type successArgs struct {
	id                             uuid.UUID
	output, finishReason           string
	tokensIn, tokensOut, latencyMs int
}

type failedArgs struct {
	id        uuid.UUID
	errMsg    string
	latencyMs int
}

type feedbackArgs struct {
	runID  uuid.UUID
	rating domain.RunFeedbackRating
}

func newFakeRunRepo() *fakeRunRepo {
	return &fakeRunRepo{store: make(map[uuid.UUID]domain.AssistantRun)}
}

func (r *fakeRunRepo) CreatePendingForActiveAssistant(_ context.Context, assistantID, userID uuid.UUID, userPrompt string) (domain.Assistant, domain.AssistantRun, error) {
	if r.createPendingErr != nil {
		return domain.Assistant{}, domain.AssistantRun{}, r.createPendingErr
	}

	assistant := domain.Assistant{
		ID:           assistantID,
		Model:        "gpt-mock",
		SystemPrompt: "be nice",
		IsActive:     true,
	}
	if r.pendingAssistant != nil {
		assistant = *r.pendingAssistant
		if assistant.ID == uuid.Nil {
			assistant.ID = assistantID
		}
	}
	if !assistant.IsActive {
		return domain.Assistant{}, domain.AssistantRun{}, domain.ErrAssistantInactive
	}

	run := domain.AssistantRun{
		ID: uuid.New(), AssistantID: assistantID, UserID: userID, Model: assistant.Model,
		UserPrompt: userPrompt, Status: domain.RunPending, CreatedAt: time.Now(),
	}
	r.created = append(r.created, run)
	r.store[run.ID] = run

	return assistant, run, nil
}

func (r *fakeRunRepo) MarkSuccess(_ context.Context, id uuid.UUID, output string, tokensIn, tokensOut, latencyMs int, finishReason string) error {
	r.successCall = &successArgs{id, output, finishReason, tokensIn, tokensOut, latencyMs}
	run := r.store[id]
	run.Status = domain.RunSuccess
	run.Output = &output
	run.TokensIn = &tokensIn
	run.TokensOut = &tokensOut
	run.LatencyMs = &latencyMs
	run.FinishReason = &finishReason
	r.store[id] = run

	return nil
}

func (r *fakeRunRepo) MarkFailed(_ context.Context, id uuid.UUID, errMsg string, latencyMs int) error {
	r.failedCall = &failedArgs{id, errMsg, latencyMs}
	run := r.store[id]
	run.Status = domain.RunFailed
	run.Error = &errMsg
	run.LatencyMs = &latencyMs
	r.store[id] = run

	return nil
}

func (r *fakeRunRepo) List(_ context.Context, _ repository.RunListFilter) ([]domain.AssistantRun, int, error) {
	return nil, 0, nil
}

func (r *fakeRunRepo) GetByID(_ context.Context, id uuid.UUID) (domain.AssistantRun, error) {
	run, ok := r.store[id]
	if !ok {
		return domain.AssistantRun{}, domain.ErrRunNotFound
	}

	return run, nil
}

func (r *fakeRunRepo) ReapPending(_ context.Context, _ time.Duration, _ string) (int64, error) {
	return 0, nil
}

func (r *fakeRunRepo) UpsertFeedback(_ context.Context, runID uuid.UUID, rating domain.RunFeedbackRating) (domain.AssistantRun, error) {
	r.feedbackCall = &feedbackArgs{runID: runID, rating: rating}
	run, ok := r.store[runID]
	if !ok {
		return domain.AssistantRun{}, domain.ErrRunNotFound
	}
	run.FeedbackRating = &rating
	r.store[runID] = run

	return run, nil
}

type fakeProvider struct {
	resp         llm.Response
	err          error
	streamChunks []string
	streamFn     func(context.Context, llm.Request, func(llm.StreamChunk)) (llm.Response, error)
}

func (f *fakeProvider) Generate(_ context.Context, _ llm.Request) (llm.Response, error) {
	return f.resp, f.err
}

func (f *fakeProvider) GenerateStream(ctx context.Context, req llm.Request, onChunk func(llm.StreamChunk)) (llm.Response, error) {
	if f.streamFn != nil {
		return f.streamFn(ctx, req, onChunk)
	}
	for _, chunk := range f.streamChunks {
		onChunk(llm.StreamChunk{Delta: chunk})
	}

	return f.resp, f.err
}

func activeAssistant() domain.Assistant {
	return domain.Assistant{
		ID: uuid.New(), CategoryID: uuid.New(), Name: "x", Model: "gpt-mock",
		SystemPrompt: "be nice", IsActive: true,
	}
}

func TestRunHappyPath(t *testing.T) {
	a := activeAssistant()
	runs := newFakeRunRepo()
	svc := NewRunService(
		context.Background(),
		runs,
		&fakeProvider{resp: llm.Response{Output: "hi", TokensIn: 1, TokensOut: 2, LatencyMs: 0, FinishReason: "stop"}},
		time.Second,
	)

	run, err := svc.Run(context.Background(), a.ID, uuid.New(), "hello")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if run.Status != domain.RunSuccess {
		t.Fatalf("status: %s", run.Status)
	}
	if runs.successCall == nil {
		t.Fatal("MarkSuccess not called")
	}
	if len(runs.created) != 1 {
		t.Fatalf("expected exactly one pending row, got %d", len(runs.created))
	}
}

func TestRunPersistsRowEvenOnProviderError(t *testing.T) {
	a := activeAssistant()
	runs := newFakeRunRepo()
	svc := NewRunService(
		context.Background(),
		runs,
		&fakeProvider{err: errors.New("boom")},
		time.Second,
	)

	run, err := svc.Run(context.Background(), a.ID, uuid.New(), "hello")
	if err == nil {
		t.Fatal("expected provider error")
	}
	if run.Status != domain.RunFailed {
		t.Fatalf("status: %s", run.Status)
	}
	if runs.failedCall == nil {
		t.Fatal("MarkFailed not called")
	}
	if len(runs.created) != 1 {
		t.Fatalf("pending row must exist on provider error, got %d", len(runs.created))
	}
}

func TestRunInactiveDoesNotCreateRow(t *testing.T) {
	a := activeAssistant()
	a.IsActive = false
	runs := newFakeRunRepo()
	runs.pendingAssistant = &a
	svc := NewRunService(
		context.Background(),
		runs,
		&fakeProvider{},
		time.Second,
	)

	_, err := svc.Run(context.Background(), a.ID, uuid.New(), "hi")
	if !errors.Is(err, domain.ErrAssistantInactive) {
		t.Fatalf("err: %v", err)
	}
	if len(runs.created) != 0 {
		t.Fatalf("must not create row for inactive assistant, got %d", len(runs.created))
	}
}

func TestRunNotFound(t *testing.T) {
	runs := newFakeRunRepo()
	runs.createPendingErr = domain.ErrAssistantNotFound
	svc := NewRunService(
		context.Background(),
		runs,
		&fakeProvider{},
		time.Second,
	)
	_, err := svc.Run(context.Background(), uuid.New(), uuid.New(), "hi")
	if !errors.Is(err, domain.ErrAssistantNotFound) {
		t.Fatalf("err: %v", err)
	}
}

func TestRunStreamHappyPath(t *testing.T) {
	a := activeAssistant()
	runs := newFakeRunRepo()
	chunks := make([]string, 0, 2)
	var created *domain.AssistantRun
	svc := NewRunService(
		context.Background(),
		runs,
		&fakeProvider{
			resp:         llm.Response{Output: "hello world", TokensIn: 3, TokensOut: 4, LatencyMs: 0, FinishReason: "stop"},
			streamChunks: []string{"hello", " world"},
		},
		time.Second,
	)

	run, err := svc.RunStream(context.Background(), a.ID, uuid.New(), "hello", RunStreamCallbacks{
		OnRunCreated: func(run domain.AssistantRun) {
			created = &run
		},
		OnDelta: func(delta string) {
			chunks = append(chunks, delta)
		},
	})
	if err != nil {
		t.Fatalf("run stream: %v", err)
	}
	if created == nil || created.Status != domain.RunPending {
		t.Fatalf("pending callback: %+v", created)
	}
	if run.Status != domain.RunSuccess || run.Output == nil || *run.Output != "hello world" {
		t.Fatalf("final run: %+v", run)
	}
	if got := strings.Join(chunks, ""); got != "hello world" {
		t.Fatalf("chunks: %q", got)
	}
}

func TestRunStreamPersistsFailure(t *testing.T) {
	a := activeAssistant()
	runs := newFakeRunRepo()
	svc := NewRunService(
		context.Background(),
		runs,
		&fakeProvider{err: errors.New("boom"), streamChunks: []string{"partial"}},
		time.Second,
	)

	_, err := svc.RunStream(context.Background(), a.ID, uuid.New(), "hello", RunStreamCallbacks{})
	if err == nil {
		t.Fatal("expected provider error")
	}
	if len(runs.created) != 1 {
		t.Fatalf("created runs: %d", len(runs.created))
	}
	persisted, getErr := runs.GetByID(context.Background(), runs.created[0].ID)
	if getErr != nil {
		t.Fatalf("get persisted run: %v", getErr)
	}
	if persisted.Status != domain.RunFailed {
		t.Fatalf("status: %s", persisted.Status)
	}
	if runs.failedCall == nil {
		t.Fatal("MarkFailed not called")
	}
}

func TestRunStreamDecouplesProviderFromCallerContext(t *testing.T) {
	a := activeAssistant()
	runs := newFakeRunRepo()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc := NewRunService(
		context.Background(),
		runs,
		&fakeProvider{
			streamFn: func(ctx context.Context, _ llm.Request, onChunk func(llm.StreamChunk)) (llm.Response, error) {
				onChunk(llm.StreamChunk{Delta: "first"})
				select {
				case <-ctx.Done():
					return llm.Response{}, ctx.Err()
				case <-time.After(50 * time.Millisecond):
				}
				onChunk(llm.StreamChunk{Delta: "second"})

				return llm.Response{Output: "first second", FinishReason: "stop"}, nil
			},
		},
		time.Second,
	)

	var deltas []string
	_, err := svc.RunStream(ctx, a.ID, uuid.New(), "hello", RunStreamCallbacks{
		OnRunCreated: func(domain.AssistantRun) {
			cancel()
		},
		OnDelta: func(d string) {
			deltas = append(deltas, d)
		},
	})
	if err != nil {
		t.Fatalf("RunStream returned error after caller cancel: %v", err)
	}
	if runs.successCall == nil {
		t.Fatal("MarkSuccess not called — провайдер должен был дойти до конца после обрыва клиента")
	}
	for _, d := range deltas {
		if d == "second" {
			t.Fatal("delta after caller cancel должна быть отброшена")
		}
	}
}

func TestSetFeedbackAllowsRunOwner(t *testing.T) {
	runs := newFakeRunRepo()
	userID := uuid.New()
	run := domain.AssistantRun{
		ID: uuid.New(), UserID: userID, Status: domain.RunSuccess, CreatedAt: time.Now(),
	}
	runs.store[run.ID] = run
	svc := NewRunService(context.Background(), runs, &fakeProvider{}, time.Second)

	updated, err := svc.SetFeedback(context.Background(), run.ID, userID, domain.RunFeedbackLike)
	if err != nil {
		t.Fatalf("set feedback: %v", err)
	}
	if runs.feedbackCall == nil || runs.feedbackCall.rating != domain.RunFeedbackLike {
		t.Fatalf("feedback call: %+v", runs.feedbackCall)
	}
	if updated.FeedbackRating == nil || *updated.FeedbackRating != domain.RunFeedbackLike {
		t.Fatalf("feedback rating: %+v", updated.FeedbackRating)
	}
}

func TestSetFeedbackRejectsOtherUser(t *testing.T) {
	runs := newFakeRunRepo()
	run := domain.AssistantRun{
		ID: uuid.New(), UserID: uuid.New(), Status: domain.RunSuccess, CreatedAt: time.Now(),
	}
	runs.store[run.ID] = run
	svc := NewRunService(context.Background(), runs, &fakeProvider{}, time.Second)

	_, err := svc.SetFeedback(context.Background(), run.ID, uuid.New(), domain.RunFeedbackDislike)
	if !errors.Is(err, domain.ErrRunForbidden) {
		t.Fatalf("err: %v", err)
	}
	if runs.feedbackCall != nil {
		t.Fatalf("feedback must not be upserted for another user: %+v", runs.feedbackCall)
	}
}
