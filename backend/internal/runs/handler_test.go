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
	fn func(ctx context.Context, assistantID, userID uuid.UUID, userPrompt string) (domain.AssistantRun, error)
}

func (f *fakeService) Run(ctx context.Context, a, u uuid.UUID, p string) (domain.AssistantRun, error) {
	return f.fn(ctx, a, u, p)
}

func (f *fakeService) List(_ context.Context, _ service.RunListInput) ([]domain.AssistantRun, int, error) {
	return nil, 0, nil
}

func authedRequest(method, path, body string, userID uuid.UUID) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.SetPathValue("assistantId", strings.TrimPrefix(strings.TrimSuffix(path, "/run"), "/assistants/"))
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
