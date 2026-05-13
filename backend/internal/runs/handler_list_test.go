package runs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"avito-internship-fs/internal/auth"
	"avito-internship-fs/internal/domain"
	"avito-internship-fs/internal/service"
)

type fakeListService struct {
	captured service.RunListInput
	items    []domain.AssistantRun
	total    int
	runFn    func(ctx context.Context, a, u uuid.UUID, p string) (domain.AssistantRun, error)
}

func (f *fakeListService) Run(ctx context.Context, a, u uuid.UUID, p string) (domain.AssistantRun, error) {
	if f.runFn == nil {
		return domain.AssistantRun{}, nil
	}

	return f.runFn(ctx, a, u, p)
}

func (f *fakeListService) RunStream(ctx context.Context, a, u uuid.UUID, p string, _ service.RunStreamCallbacks) (domain.AssistantRun, error) {
	return f.Run(ctx, a, u, p)
}

func (f *fakeListService) List(_ context.Context, in service.RunListInput) ([]domain.AssistantRun, int, error) {
	f.captured = in
	return f.items, f.total, nil
}

func (f *fakeListService) SetFeedback(context.Context, uuid.UUID, uuid.UUID, domain.RunFeedbackRating) (domain.AssistantRun, error) {
	return domain.AssistantRun{}, nil
}

func authedListRequest(method, path string, principal auth.Principal) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	ctx := auth.WithPrincipalForTest(req.Context(), principal)

	return req.WithContext(ctx)
}

func TestMyRunsScopedToCurrentUser(t *testing.T) {
	userID := uuid.New()
	svc := &fakeListService{
		items: []domain.AssistantRun{
			{ID: uuid.New(), UserID: userID, Status: domain.RunSuccess, CreatedAt: time.Now()},
		},
		total: 1,
	}
	h := NewHandler(svc)
	rr := httptest.NewRecorder()
	h.MyRuns(rr, authedListRequest(http.MethodGet, "/runs/my",
		auth.Principal{UserID: userID, Role: auth.RoleUser}))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	if svc.captured.UserID == nil || *svc.captured.UserID != userID {
		t.Fatalf("MyRuns must scope to current user: %+v", svc.captured.UserID)
	}
}

func TestMyRunsRejectsAssistantIDQueryParam(t *testing.T) {
	// MyRuns does not parse assistantId from query - this user-provided value
	// must be ignored, not used to widen the scope.
	userID := uuid.New()
	svc := &fakeListService{}
	h := NewHandler(svc)
	otherAssistant := uuid.New()
	rr := httptest.NewRecorder()
	h.MyRuns(rr, authedListRequest(http.MethodGet, "/runs/my?assistantId="+otherAssistant.String(),
		auth.Principal{UserID: userID, Role: auth.RoleUser}))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d", rr.Code)
	}
	if svc.captured.AssistantID != nil {
		t.Fatalf("MyRuns must ignore assistantId query param, got %+v", svc.captured.AssistantID)
	}
}

func TestMyRunsParsesStatusFilter(t *testing.T) {
	svc := &fakeListService{}
	h := NewHandler(svc)
	rr := httptest.NewRecorder()
	h.MyRuns(rr, authedListRequest(http.MethodGet, "/runs/my?status=failed",
		auth.Principal{UserID: uuid.New(), Role: auth.RoleUser}))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d", rr.Code)
	}
	if svc.captured.Status == nil || *svc.captured.Status != domain.RunFailed {
		t.Fatalf("status filter not parsed: %+v", svc.captured.Status)
	}
}

func TestMyRunsRejectsInvalidStatus(t *testing.T) {
	h := NewHandler(&fakeListService{})
	rr := httptest.NewRecorder()
	h.MyRuns(rr, authedListRequest(http.MethodGet, "/runs/my?status=unknown",
		auth.Principal{UserID: uuid.New(), Role: auth.RoleUser}))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestMyRunsRequiresAuth(t *testing.T) {
	h := NewHandler(&fakeListService{})
	rr := httptest.NewRecorder()
	h.MyRuns(rr, httptest.NewRequest(http.MethodGet, "/runs/my", nil))

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestAdminRunsParsesAllFilters(t *testing.T) {
	assistantID := uuid.New()
	svc := &fakeListService{}
	h := NewHandler(svc)
	rr := httptest.NewRecorder()
	h.AdminRuns(rr, authedListRequest(http.MethodGet,
		"/admin/runs?status=success&assistantId="+assistantID.String()+"&page=2&pageSize=15",
		auth.Principal{UserID: uuid.New(), Role: auth.RoleAdmin}))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	if svc.captured.AssistantID == nil || *svc.captured.AssistantID != assistantID {
		t.Fatalf("assistantId not parsed: %+v", svc.captured.AssistantID)
	}
	if svc.captured.Status == nil || *svc.captured.Status != domain.RunSuccess {
		t.Fatalf("status not parsed: %+v", svc.captured.Status)
	}
	if svc.captured.UserID != nil {
		t.Fatalf("AdminRuns must not scope by user, got %+v", svc.captured.UserID)
	}
	if svc.captured.Page != 2 || svc.captured.PageSize != 15 {
		t.Fatalf("pagination: %+v", svc.captured)
	}
}

func TestAdminRunsRejectsInvalidAssistantID(t *testing.T) {
	h := NewHandler(&fakeListService{})
	rr := httptest.NewRecorder()
	h.AdminRuns(rr, authedListRequest(http.MethodGet, "/admin/runs?assistantId=not-a-uuid",
		auth.Principal{UserID: uuid.New(), Role: auth.RoleAdmin}))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestAdminRunsReturnsEmptyArrayNotNull(t *testing.T) {
	svc := &fakeListService{items: nil, total: 0}
	h := NewHandler(svc)
	rr := httptest.NewRecorder()
	h.AdminRuns(rr, authedListRequest(http.MethodGet, "/admin/runs",
		auth.Principal{UserID: uuid.New(), Role: auth.RoleAdmin}))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d", rr.Code)
	}
	var got listResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Runs == nil {
		t.Fatalf("Runs must be empty slice, not nil")
	}
}
