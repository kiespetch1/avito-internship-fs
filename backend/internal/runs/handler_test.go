package runs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"avito-internship-fs/internal/api"
	"avito-internship-fs/internal/auth"
	"avito-internship-fs/internal/domain"
	"avito-internship-fs/internal/llm"
	"avito-internship-fs/internal/service"
)

type fakeService struct {
	fn         func(ctx context.Context, assistantID, userID uuid.UUID, userPrompt string) (domain.AssistantRun, error)
	streamFn   func(ctx context.Context, assistantID, userID uuid.UUID, userPrompt string, callbacks service.RunStreamCallbacks) (domain.AssistantRun, error)
	feedbackFn func(ctx context.Context, runID, userID uuid.UUID, rating domain.RunFeedbackRating) (domain.AssistantRun, error)
}

func (f *fakeService) Run(ctx context.Context, a, u uuid.UUID, p string) (domain.AssistantRun, error) {
	return f.fn(ctx, a, u, p)
}

func (f *fakeService) RunStream(ctx context.Context, a, u uuid.UUID, p string, callbacks service.RunStreamCallbacks) (domain.AssistantRun, error) {
	if f.streamFn != nil {
		return f.streamFn(ctx, a, u, p, callbacks)
	}

	return f.Run(ctx, a, u, p)
}

func (f *fakeService) List(_ context.Context, _ service.RunListInput) ([]domain.AssistantRun, int, error) {
	return nil, 0, nil
}

func (f *fakeService) SetFeedback(ctx context.Context, runID, userID uuid.UUID, rating domain.RunFeedbackRating) (domain.AssistantRun, error) {
	return f.feedbackFn(ctx, runID, userID, rating)
}

type failingStreamWriter struct {
	header http.Header
	status int
}

func newFailingStreamWriter() *failingStreamWriter {
	return &failingStreamWriter{header: make(http.Header)}
}

func (w *failingStreamWriter) Header() http.Header {
	return w.header
}

func (w *failingStreamWriter) WriteHeader(status int) {
	w.status = status
}

func (w *failingStreamWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("write failed")
}

func (w *failingStreamWriter) Flush() {}

func authedRequest(method, path, body string, userID uuid.UUID) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.SetPathValue("assistantId", strings.TrimPrefix(strings.TrimSuffix(path, "/run"), "/assistants/"))
	ctx := auth.WithPrincipalForTest(req.Context(), auth.Principal{UserID: userID, Role: auth.RoleUser})

	return req.WithContext(ctx)
}

func authedStreamRequest(method, path, body string, userID uuid.UUID) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.SetPathValue("assistantId", strings.TrimPrefix(strings.TrimSuffix(path, "/run/stream"), "/assistants/"))
	ctx := auth.WithPrincipalForTest(req.Context(), auth.Principal{UserID: userID, Role: auth.RoleUser})

	return req.WithContext(ctx)
}

func TestRunHandlerSuccess(t *testing.T) {
	aid := uuid.New()
	uid := uuid.New()
	out := "[mock:gpt-mock] hi"
	svc := &fakeService{
		fn: func(_ context.Context, gotA, gotU uuid.UUID, p string) (domain.AssistantRun, error) {
			if gotA != aid || gotU != uid || p != "hi" {
				t.Fatalf("args: %v %v %q", gotA, gotU, p)
			}

			return domain.AssistantRun{ID: uuid.New(), AssistantID: aid, UserID: uid, Model: "gpt-mock", UserPrompt: "hi", Output: &out, Status: domain.RunSuccess, CreatedAt: time.Now()}, nil
		},
	}
	h := NewHandler(svc)
	rr := httptest.NewRecorder()
	h.Run(rr, authedRequest(http.MethodPost, "/assistants/"+aid.String()+"/run", `{"userPrompt":"hi"}`, uid))

	if rr.Code != http.StatusCreated {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	var got api.AssistantRun
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Status != api.Success {
		t.Fatalf("status: %s", got.Status)
	}
}

func TestRunHandlerEmptyPrompt(t *testing.T) {
	aid := uuid.New()
	h := NewHandler(&fakeService{fn: func(_ context.Context, _, _ uuid.UUID, _ string) (domain.AssistantRun, error) {
		t.Fatal("service must not be called")
		return domain.AssistantRun{}, nil
	}})
	rr := httptest.NewRecorder()
	h.Run(rr, authedRequest(http.MethodPost, "/assistants/"+aid.String()+"/run", `{"userPrompt":"   "}`, uuid.New()))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestRunHandlerNotFound(t *testing.T) {
	aid := uuid.New()
	h := NewHandler(&fakeService{fn: func(_ context.Context, _, _ uuid.UUID, _ string) (domain.AssistantRun, error) {
		return domain.AssistantRun{}, domain.ErrAssistantNotFound
	}})
	rr := httptest.NewRecorder()
	h.Run(rr, authedRequest(http.MethodPost, "/assistants/"+aid.String()+"/run", `{"userPrompt":"hi"}`, uuid.New()))

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestRunHandlerInactive(t *testing.T) {
	aid := uuid.New()
	h := NewHandler(&fakeService{fn: func(_ context.Context, _, _ uuid.UUID, _ string) (domain.AssistantRun, error) {
		return domain.AssistantRun{}, domain.ErrAssistantInactive
	}})
	rr := httptest.NewRecorder()
	h.Run(rr, authedRequest(http.MethodPost, "/assistants/"+aid.String()+"/run", `{"userPrompt":"hi"}`, uuid.New()))

	if rr.Code != http.StatusConflict {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestRunHandlerProviderError(t *testing.T) {
	aid := uuid.New()
	runID := uuid.New()
	errMsg := "boom"
	h := NewHandler(&fakeService{fn: func(_ context.Context, a, u uuid.UUID, _ string) (domain.AssistantRun, error) {
		return domain.AssistantRun{ID: runID, AssistantID: a, UserID: u, Model: "m", UserPrompt: "hi", Status: domain.RunFailed, Error: &errMsg, CreatedAt: time.Now()}, fmt.Errorf("%w: %s", llm.ErrProviderFailed, errMsg)
	}})
	rr := httptest.NewRecorder()
	h.Run(rr, authedRequest(http.MethodPost, "/assistants/"+aid.String()+"/run", `{"userPrompt":"hi"}`, uuid.New()))

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestRunHandlerInvalidAssistantID(t *testing.T) {
	h := NewHandler(&fakeService{})
	req := httptest.NewRequest(http.MethodPost, "/assistants/not-a-uuid/run", strings.NewReader(`{"userPrompt":"hi"}`))
	req.SetPathValue("assistantId", "not-a-uuid")
	ctx := auth.WithPrincipalForTest(req.Context(), auth.Principal{UserID: uuid.New(), Role: auth.RoleUser})
	rr := httptest.NewRecorder()
	h.Run(rr, req.WithContext(ctx))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestSetFeedbackHandlerSuccess(t *testing.T) {
	runID := uuid.New()
	userID := uuid.New()
	rating := domain.RunFeedbackLike
	h := NewHandler(&fakeService{
		feedbackFn: func(_ context.Context, gotRunID, gotUserID uuid.UUID, gotRating domain.RunFeedbackRating) (domain.AssistantRun, error) {
			if gotRunID != runID || gotUserID != userID || gotRating != rating {
				t.Fatalf("args: %v %v %d", gotRunID, gotUserID, gotRating)
			}

			return domain.AssistantRun{
				ID: runID, UserID: userID, Status: domain.RunSuccess, FeedbackRating: &rating, CreatedAt: time.Now(),
			}, nil
		},
	})
	req := httptest.NewRequest(http.MethodPut, "/runs/"+runID.String()+"/feedback", strings.NewReader(`{"rating":1}`))
	req.SetPathValue("runId", runID.String())
	ctx := auth.WithPrincipalForTest(req.Context(), auth.Principal{UserID: userID, Role: auth.RoleUser})
	rr := httptest.NewRecorder()
	h.SetFeedback(rr, req.WithContext(ctx))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	var got api.AssistantRun
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.FeedbackRating == nil || *got.FeedbackRating != api.N1 {
		t.Fatalf("feedbackRating: %+v", got.FeedbackRating)
	}
}

func TestSetFeedbackHandlerRejectsInvalidRating(t *testing.T) {
	h := NewHandler(&fakeService{feedbackFn: func(context.Context, uuid.UUID, uuid.UUID, domain.RunFeedbackRating) (domain.AssistantRun, error) {
		t.Fatal("service must not be called")
		return domain.AssistantRun{}, nil
	}})
	runID := uuid.New()
	req := httptest.NewRequest(http.MethodPut, "/runs/"+runID.String()+"/feedback", strings.NewReader(`{"rating":0}`))
	req.SetPathValue("runId", runID.String())
	ctx := auth.WithPrincipalForTest(req.Context(), auth.Principal{UserID: uuid.New(), Role: auth.RoleUser})
	rr := httptest.NewRecorder()
	h.SetFeedback(rr, req.WithContext(ctx))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestSetFeedbackHandlerRejectsOtherUserRun(t *testing.T) {
	h := NewHandler(&fakeService{feedbackFn: func(context.Context, uuid.UUID, uuid.UUID, domain.RunFeedbackRating) (domain.AssistantRun, error) {
		return domain.AssistantRun{}, domain.ErrRunForbidden
	}})
	runID := uuid.New()
	req := httptest.NewRequest(http.MethodPut, "/runs/"+runID.String()+"/feedback", strings.NewReader(`{"rating":1}`))
	req.SetPathValue("runId", runID.String())
	ctx := auth.WithPrincipalForTest(req.Context(), auth.Principal{UserID: uuid.New(), Role: auth.RoleUser})
	rr := httptest.NewRecorder()
	h.SetFeedback(rr, req.WithContext(ctx))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestRunStreamHandlerSuccess(t *testing.T) {
	aid := uuid.New()
	uid := uuid.New()
	out := "hello world"
	runID := uuid.New()
	h := NewHandler(&fakeService{
		streamFn: func(_ context.Context, gotA, gotU uuid.UUID, p string, callbacks service.RunStreamCallbacks) (domain.AssistantRun, error) {
			if gotA != aid || gotU != uid || p != "hi" {
				t.Fatalf("args: %v %v %q", gotA, gotU, p)
			}
			pending := domain.AssistantRun{ID: runID, AssistantID: aid, UserID: uid, Model: "m", UserPrompt: p, Status: domain.RunPending, CreatedAt: time.Now()}
			callbacks.OnRunCreated(pending)
			callbacks.OnDelta("hello")
			callbacks.OnDelta(" world")
			pending.Status = domain.RunSuccess
			pending.Output = &out

			return pending, nil
		},
	})

	rr := httptest.NewRecorder()
	h.RunStream(rr, authedStreamRequest(http.MethodPost, "/assistants/"+aid.String()+"/run/stream", `{"userPrompt":"hi"}`, uid))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	for _, want := range []string{"event: run", `"status":"pending"`, "event: delta", `"delta":"hello"`, "event: done", `"status":"success"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("stream does not contain %q: %s", want, body)
		}
	}
	if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/event-stream") {
		t.Fatalf("content-type: %q", got)
	}
}

func TestRunStreamHandlerProviderErrorAfterStart(t *testing.T) {
	aid := uuid.New()
	uid := uuid.New()
	errMsg := "boom"
	runID := uuid.New()
	h := NewHandler(&fakeService{
		streamFn: func(_ context.Context, _, _ uuid.UUID, p string, callbacks service.RunStreamCallbacks) (domain.AssistantRun, error) {
			failed := domain.AssistantRun{ID: runID, AssistantID: aid, UserID: uid, Model: "m", UserPrompt: p, Status: domain.RunPending, CreatedAt: time.Now()}
			callbacks.OnRunCreated(failed)
			failed.Status = domain.RunFailed
			failed.Error = &errMsg

			return failed, fmt.Errorf("%w: %s", llm.ErrProviderFailed, errMsg)
		},
	})

	rr := httptest.NewRecorder()
	h.RunStream(rr, authedStreamRequest(http.MethodPost, "/assistants/"+aid.String()+"/run/stream", `{"userPrompt":"hi"}`, uid))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	for _, want := range []string{"event: run", "event: failed", `"status":"failed"`, `"code":"LLM_PROVIDER_ERROR"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("stream does not contain %q: %s", want, body)
		}
	}
}

func TestRunStreamHandlerCancelsServiceWhenWriteFails(t *testing.T) {
	aid := uuid.New()
	uid := uuid.New()
	runID := uuid.New()
	contextCanceled := false
	h := NewHandler(&fakeService{
		streamFn: func(ctx context.Context, _, _ uuid.UUID, p string, callbacks service.RunStreamCallbacks) (domain.AssistantRun, error) {
			pending := domain.AssistantRun{ID: runID, AssistantID: aid, UserID: uid, Model: "m", UserPrompt: p, Status: domain.RunPending, CreatedAt: time.Now()}
			callbacks.OnRunCreated(pending)
			select {
			case <-ctx.Done():
				contextCanceled = true
				return pending, ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return pending, fmt.Errorf("context was not canceled")
			}
		},
	})

	w := newFailingStreamWriter()
	h.RunStream(w, authedStreamRequest(http.MethodPost, "/assistants/"+aid.String()+"/run/stream", `{"userPrompt":"hi"}`, uid))

	if !contextCanceled {
		t.Fatal("service context was not canceled after stream write failure")
	}
	if w.status != http.StatusOK {
		t.Fatalf("status: %d", w.status)
	}
}
