package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"avito-internship-fs/internal/domain"
	"avito-internship-fs/internal/llm"
	"avito-internship-fs/internal/repository"
)

type fakeAssistantRepo struct {
	getFn func(ctx context.Context, id uuid.UUID) (domain.Assistant, error)
}

func (f *fakeAssistantRepo) Get(ctx context.Context, id uuid.UUID) (domain.Assistant, error) {
	return f.getFn(ctx, id)
}

type fakeRunRepo struct {
	created     []domain.AssistantRun
	successCall *successArgs
	failedCall  *failedArgs
	store       map[uuid.UUID]domain.AssistantRun
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

func newFakeRunRepo() *fakeRunRepo {
	return &fakeRunRepo{store: make(map[uuid.UUID]domain.AssistantRun)}
}

func (r *fakeRunRepo) CreatePending(_ context.Context, assistantID, userID uuid.UUID, model, userPrompt string) (domain.AssistantRun, error) {
	run := domain.AssistantRun{
		ID: uuid.New(), AssistantID: assistantID, UserID: userID, Model: model,
		UserPrompt: userPrompt, Status: domain.RunPending, CreatedAt: time.Now(),
	}
	r.created = append(r.created, run)
	r.store[run.ID] = run

	return run, nil
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
		return domain.AssistantRun{}, errors.New("not found")
	}

	return run, nil
}

type fakeProvider struct {
	resp llm.Response
	err  error
}

func (f *fakeProvider) Generate(_ context.Context, _ llm.Request) (llm.Response, error) {
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
		&fakeAssistantRepo{getFn: func(_ context.Context, _ uuid.UUID) (domain.Assistant, error) { return a, nil }},
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
		&fakeAssistantRepo{getFn: func(_ context.Context, _ uuid.UUID) (domain.Assistant, error) { return a, nil }},
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
	svc := NewRunService(
		&fakeAssistantRepo{getFn: func(_ context.Context, _ uuid.UUID) (domain.Assistant, error) { return a, nil }},
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
	svc := NewRunService(
		&fakeAssistantRepo{getFn: func(_ context.Context, _ uuid.UUID) (domain.Assistant, error) {
			return domain.Assistant{}, domain.ErrAssistantNotFound
		}},
		newFakeRunRepo(),
		&fakeProvider{},
		time.Second,
	)
	_, err := svc.Run(context.Background(), uuid.New(), uuid.New(), "hi")
	if !errors.Is(err, domain.ErrAssistantNotFound) {
		t.Fatalf("err: %v", err)
	}
}
