package assistants

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"avito-internship-fs/internal/api"
	"avito-internship-fs/internal/auth"
	"avito-internship-fs/internal/domain"
	"avito-internship-fs/internal/service"
)

type fakeService struct {
	getFn         func(ctx context.Context, userID, id uuid.UUID) (domain.Assistant, error)
	createFn      func(ctx context.Context, in service.AssistantCreateInput) (domain.Assistant, error)
	updateFn      func(ctx context.Context, in service.AssistantUpdateInput) (domain.Assistant, error)
	listFn        func(ctx context.Context, in service.AssistantListInput) ([]domain.Assistant, int, error)
	setFavoriteFn func(ctx context.Context, userID, assistantID uuid.UUID, favorite bool) error
}

func (f *fakeService) Get(ctx context.Context, userID, id uuid.UUID) (domain.Assistant, error) {
	return f.getFn(ctx, userID, id)
}

func (f *fakeService) Create(ctx context.Context, in service.AssistantCreateInput) (domain.Assistant, error) {
	return f.createFn(ctx, in)
}

func (f *fakeService) Update(ctx context.Context, in service.AssistantUpdateInput) (domain.Assistant, error) {
	return f.updateFn(ctx, in)
}

func (f *fakeService) List(ctx context.Context, in service.AssistantListInput) ([]domain.Assistant, int, error) {
	return f.listFn(ctx, in)
}

func (f *fakeService) SetFavorite(ctx context.Context, userID, assistantID uuid.UUID, favorite bool) error {
	return f.setFavoriteFn(ctx, userID, assistantID, favorite)
}

func withPrincipal(r *http.Request, role auth.Role) *http.Request {
	ctx := auth.WithPrincipalForTest(r.Context(), auth.Principal{UserID: uuid.New(), Role: role})
	return r.WithContext(ctx)
}

func sampleAssistant() domain.Assistant {
	return domain.Assistant{
		ID: uuid.New(), CategoryID: uuid.New(), Name: "Повар",
		Description: "рецепты", Model: "gpt-4o-mini", SystemPrompt: "be a chef",
		IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
}

// ----- Get -----

func TestGetReturnsAssistant(t *testing.T) {
	a := sampleAssistant()
	a.IsFavorite = true
	h := NewHandler(&fakeService{
		getFn: func(_ context.Context, _, _ uuid.UUID) (domain.Assistant, error) { return a, nil },
	})
	req := httptest.NewRequest(http.MethodGet, "/assistants/"+a.ID.String(), nil)
	req.SetPathValue("assistantId", a.ID.String())
	rr := httptest.NewRecorder()
	h.Get(rr, withPrincipal(req, auth.RoleAdmin))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	var got api.Assistant
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Name != "Повар" {
		t.Fatalf("name: %q", got.Name)
	}
	if got.SystemPrompt == nil || *got.SystemPrompt != "be a chef" {
		t.Fatalf("admin must see systemPrompt, got %+v", got.SystemPrompt)
	}
	if !got.IsFavorite {
		t.Fatalf("isFavorite must be returned")
	}
}

func TestGetHidesSystemPromptFromRegularUser(t *testing.T) {
	a := sampleAssistant()
	h := NewHandler(&fakeService{
		getFn: func(_ context.Context, _, _ uuid.UUID) (domain.Assistant, error) { return a, nil },
	})
	req := httptest.NewRequest(http.MethodGet, "/assistants/"+a.ID.String(), nil)
	req.SetPathValue("assistantId", a.ID.String())
	rr := httptest.NewRecorder()
	h.Get(rr, withPrincipal(req, auth.RoleUser))

	var got api.Assistant
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.SystemPrompt != nil {
		t.Fatalf("regular user must not see systemPrompt: %v", *got.SystemPrompt)
	}
}

func TestGetInvalidID(t *testing.T) {
	h := NewHandler(&fakeService{})
	req := httptest.NewRequest(http.MethodGet, "/assistants/not-a-uuid", nil)
	req.SetPathValue("assistantId", "not-a-uuid")
	rr := httptest.NewRecorder()
	h.Get(rr, withPrincipal(req, auth.RoleUser))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestGetNotFound(t *testing.T) {
	id := uuid.New()
	h := NewHandler(&fakeService{
		getFn: func(_ context.Context, _, _ uuid.UUID) (domain.Assistant, error) {
			return domain.Assistant{}, domain.ErrAssistantNotFound
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/assistants/"+id.String(), nil)
	req.SetPathValue("assistantId", id.String())
	rr := httptest.NewRecorder()
	h.Get(rr, withPrincipal(req, auth.RoleUser))

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status: %d", rr.Code)
	}
}

// ----- Create -----

func validCreateBody(categoryID uuid.UUID) string {
	return `{
		"categoryId":"` + categoryID.String() + `",
		"name":"Повар",
		"description":"рецепты",
		"model":"gpt-4o-mini",
		"systemPrompt":"be a chef"
	}`
}

func TestCreateSuccess(t *testing.T) {
	categoryID := uuid.New()
	var captured service.AssistantCreateInput
	h := NewHandler(&fakeService{
		createFn: func(_ context.Context, in service.AssistantCreateInput) (domain.Assistant, error) {
			captured = in
			a := sampleAssistant()
			a.CategoryID = in.CategoryID
			a.Name = in.Name

			return a, nil
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/assistants", strings.NewReader(validCreateBody(categoryID)))
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	if captured.CategoryID != categoryID || captured.Name != "Повар" || !captured.IsActive {
		t.Fatalf("captured: %+v", captured)
	}
}

func TestCreateRejectsMissingSystemPrompt(t *testing.T) {
	categoryID := uuid.New()
	h := NewHandler(&fakeService{
		createFn: func(_ context.Context, _ service.AssistantCreateInput) (domain.Assistant, error) {
			t.Fatal("service must not be called")
			return domain.Assistant{}, nil
		},
	})
	body := `{"categoryId":"` + categoryID.String() + `","name":"x","description":"d","model":"m","systemPrompt":""}`
	req := httptest.NewRequest(http.MethodPost, "/assistants", strings.NewReader(body))
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestCreateRejectsMissingName(t *testing.T) {
	categoryID := uuid.New()
	h := NewHandler(&fakeService{})
	body := `{"categoryId":"` + categoryID.String() + `","name":"  ","description":"d","model":"m","systemPrompt":"s"}`
	req := httptest.NewRequest(http.MethodPost, "/assistants", strings.NewReader(body))
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestCreateRejectsMalformedJSON(t *testing.T) {
	h := NewHandler(&fakeService{})
	req := httptest.NewRequest(http.MethodPost, "/assistants", strings.NewReader(`not json`))
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestCreateMapsCategoryNotFoundTo400(t *testing.T) {
	categoryID := uuid.New()
	h := NewHandler(&fakeService{
		createFn: func(_ context.Context, _ service.AssistantCreateInput) (domain.Assistant, error) {
			return domain.Assistant{}, domain.ErrCategoryNotFound
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/assistants", strings.NewReader(validCreateBody(categoryID)))
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "CATEGORY_NOT_FOUND") {
		t.Fatalf("expected CATEGORY_NOT_FOUND error code, got %s", rr.Body.String())
	}
}

func TestCreateRespectsExplicitIsActiveFalse(t *testing.T) {
	categoryID := uuid.New()
	var captured service.AssistantCreateInput
	h := NewHandler(&fakeService{
		createFn: func(_ context.Context, in service.AssistantCreateInput) (domain.Assistant, error) {
			captured = in
			return sampleAssistant(), nil
		},
	})
	body := `{"categoryId":"` + categoryID.String() + `","name":"x","description":"d","model":"m","systemPrompt":"s","isActive":false}`
	req := httptest.NewRequest(http.MethodPost, "/assistants", strings.NewReader(body))
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	if captured.IsActive {
		t.Fatalf("IsActive=false must propagate, got true")
	}
}

// ----- Update -----

func TestUpdateSuccess(t *testing.T) {
	id := uuid.New()
	categoryID := uuid.New()
	var captured service.AssistantUpdateInput
	h := NewHandler(&fakeService{
		updateFn: func(_ context.Context, in service.AssistantUpdateInput) (domain.Assistant, error) {
			captured = in
			a := sampleAssistant()
			a.ID = in.ID
			a.Name = in.Name

			return a, nil
		},
	})
	body := `{"categoryId":"` + categoryID.String() + `","name":"renamed","description":"d","model":"m","systemPrompt":"s","isActive":false}`
	req := httptest.NewRequest(http.MethodPut, "/assistants/"+id.String(), strings.NewReader(body))
	req.SetPathValue("assistantId", id.String())
	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	if captured.ID != id || captured.Name != "renamed" || captured.IsActive {
		t.Fatalf("captured: %+v", captured)
	}
}

func TestUpdateInvalidID(t *testing.T) {
	h := NewHandler(&fakeService{})
	req := httptest.NewRequest(http.MethodPut, "/assistants/x", strings.NewReader(`{}`))
	req.SetPathValue("assistantId", "x")
	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestUpdateRejectsMissingSystemPrompt(t *testing.T) {
	id := uuid.New()
	categoryID := uuid.New()
	h := NewHandler(&fakeService{})
	body := `{"categoryId":"` + categoryID.String() + `","name":"x","description":"d","model":"m","systemPrompt":"","isActive":true}`
	req := httptest.NewRequest(http.MethodPut, "/assistants/"+id.String(), strings.NewReader(body))
	req.SetPathValue("assistantId", id.String())
	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestUpdateMapsAssistantNotFoundTo404(t *testing.T) {
	id := uuid.New()
	categoryID := uuid.New()
	h := NewHandler(&fakeService{
		updateFn: func(_ context.Context, _ service.AssistantUpdateInput) (domain.Assistant, error) {
			return domain.Assistant{}, domain.ErrAssistantNotFound
		},
	})
	body := `{"categoryId":"` + categoryID.String() + `","name":"x","description":"d","model":"m","systemPrompt":"s","isActive":true}`
	req := httptest.NewRequest(http.MethodPut, "/assistants/"+id.String(), strings.NewReader(body))
	req.SetPathValue("assistantId", id.String())
	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
}

// ----- List -----

func TestListSuccess(t *testing.T) {
	a := sampleAssistant()
	a.IsFavorite = true
	var captured service.AssistantListInput
	h := NewHandler(&fakeService{
		listFn: func(_ context.Context, in service.AssistantListInput) ([]domain.Assistant, int, error) {
			captured = in
			return []domain.Assistant{a}, 1, nil
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/assistants?page=1&pageSize=5", nil)
	rr := httptest.NewRecorder()
	h.List(rr, withPrincipal(req, auth.RoleUser))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	if captured.Page != 1 || captured.PageSize != 5 {
		t.Fatalf("pagination: %+v", captured)
	}
	var got listResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Pagination.Total != 1 || len(got.Assistants) != 1 {
		t.Fatalf("body: %+v", got)
	}
	if got.Assistants[0].SystemPrompt != nil {
		t.Fatalf("regular user must not see systemPrompt in list")
	}
	if !got.Assistants[0].IsFavorite {
		t.Fatalf("isFavorite must be returned in list")
	}
}

func TestListParsesCategoryFilter(t *testing.T) {
	categoryID := uuid.New()
	var captured service.AssistantListInput
	h := NewHandler(&fakeService{
		listFn: func(_ context.Context, in service.AssistantListInput) ([]domain.Assistant, int, error) {
			captured = in
			return nil, 0, nil
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/assistants?categoryId="+categoryID.String(), nil)
	rr := httptest.NewRecorder()
	h.List(rr, withPrincipal(req, auth.RoleUser))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	if captured.CategoryID == nil || *captured.CategoryID != categoryID {
		t.Fatalf("categoryId not parsed: %+v", captured.CategoryID)
	}
}

func TestListParsesQuery(t *testing.T) {
	var captured service.AssistantListInput
	h := NewHandler(&fakeService{
		listFn: func(_ context.Context, in service.AssistantListInput) ([]domain.Assistant, int, error) {
			captured = in
			return nil, 0, nil
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/assistants?q=%20%20повар%20%20", nil)
	rr := httptest.NewRecorder()
	h.List(rr, withPrincipal(req, auth.RoleUser))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d", rr.Code)
	}
	if captured.Query == nil || *captured.Query != "повар" {
		t.Fatalf("query not trimmed/parsed: %+v", captured.Query)
	}
}

func TestListRejectsInvalidCategoryID(t *testing.T) {
	h := NewHandler(&fakeService{})
	req := httptest.NewRequest(http.MethodGet, "/assistants?categoryId=not-a-uuid", nil)
	rr := httptest.NewRecorder()
	h.List(rr, withPrincipal(req, auth.RoleUser))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestListRejectsIncludeInactiveForNonAdmin(t *testing.T) {
	h := NewHandler(&fakeService{})
	req := httptest.NewRequest(http.MethodGet, "/assistants?includeInactive=true", nil)
	rr := httptest.NewRecorder()
	h.List(rr, withPrincipal(req, auth.RoleUser))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestListAllowsIncludeInactiveForAdmin(t *testing.T) {
	var captured service.AssistantListInput
	h := NewHandler(&fakeService{
		listFn: func(_ context.Context, in service.AssistantListInput) ([]domain.Assistant, int, error) {
			captured = in
			return nil, 0, nil
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/assistants?includeInactive=true", nil)
	rr := httptest.NewRecorder()
	h.List(rr, withPrincipal(req, auth.RoleAdmin))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	if !captured.IncludeInactive {
		t.Fatalf("admin includeInactive=true must propagate")
	}
}

func TestListParsesFavoriteOnly(t *testing.T) {
	var captured service.AssistantListInput
	h := NewHandler(&fakeService{
		listFn: func(_ context.Context, in service.AssistantListInput) ([]domain.Assistant, int, error) {
			captured = in
			return nil, 0, nil
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/assistants?favoriteOnly=true", nil)
	rr := httptest.NewRecorder()
	h.List(rr, withPrincipal(req, auth.RoleUser))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	if !captured.FavoriteOnly {
		t.Fatalf("favoriteOnly=true must propagate")
	}
}

func TestListRejectsInvalidFavoriteOnly(t *testing.T) {
	h := NewHandler(&fakeService{})
	req := httptest.NewRequest(http.MethodGet, "/assistants?favoriteOnly=nope", nil)
	rr := httptest.NewRecorder()
	h.List(rr, withPrincipal(req, auth.RoleUser))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestListRejectsBadPagination(t *testing.T) {
	h := NewHandler(&fakeService{})
	req := httptest.NewRequest(http.MethodGet, "/assistants?page=abc", nil)
	rr := httptest.NewRecorder()
	h.List(rr, withPrincipal(req, auth.RoleUser))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestListRequiresAuth(t *testing.T) {
	h := NewHandler(&fakeService{})
	req := httptest.NewRequest(http.MethodGet, "/assistants", nil)
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestAddFavoriteSuccess(t *testing.T) {
	assistantID := uuid.New()
	var capturedFavorite bool
	h := NewHandler(&fakeService{
		setFavoriteFn: func(_ context.Context, _ uuid.UUID, gotAssistantID uuid.UUID, favorite bool) error {
			if gotAssistantID != assistantID {
				t.Fatalf("assistant id: got %v want %v", gotAssistantID, assistantID)
			}
			capturedFavorite = favorite

			return nil
		},
	})
	req := httptest.NewRequest(http.MethodPut, "/assistants/"+assistantID.String()+"/favorite", nil)
	req.SetPathValue("assistantId", assistantID.String())
	rr := httptest.NewRecorder()
	h.AddFavorite(rr, withPrincipal(req, auth.RoleUser))

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	if !capturedFavorite {
		t.Fatalf("favorite=true must propagate")
	}
}

func TestRemoveFavoriteSuccess(t *testing.T) {
	assistantID := uuid.New()
	var capturedFavorite bool
	h := NewHandler(&fakeService{
		setFavoriteFn: func(_ context.Context, _ uuid.UUID, gotAssistantID uuid.UUID, favorite bool) error {
			if gotAssistantID != assistantID {
				t.Fatalf("assistant id: got %v want %v", gotAssistantID, assistantID)
			}
			capturedFavorite = favorite

			return nil
		},
	})
	req := httptest.NewRequest(http.MethodDelete, "/assistants/"+assistantID.String()+"/favorite", nil)
	req.SetPathValue("assistantId", assistantID.String())
	rr := httptest.NewRecorder()
	h.RemoveFavorite(rr, withPrincipal(req, auth.RoleUser))

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
	if capturedFavorite {
		t.Fatalf("favorite=false must propagate")
	}
}

func TestSetFavoriteInvalidID(t *testing.T) {
	h := NewHandler(&fakeService{})
	req := httptest.NewRequest(http.MethodPut, "/assistants/not-a-uuid/favorite", nil)
	req.SetPathValue("assistantId", "not-a-uuid")
	rr := httptest.NewRecorder()
	h.AddFavorite(rr, withPrincipal(req, auth.RoleUser))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestSetFavoriteMapsAssistantNotFound(t *testing.T) {
	assistantID := uuid.New()
	h := NewHandler(&fakeService{
		setFavoriteFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ bool) error {
			return domain.ErrAssistantNotFound
		},
	})
	req := httptest.NewRequest(http.MethodPut, "/assistants/"+assistantID.String()+"/favorite", nil)
	req.SetPathValue("assistantId", assistantID.String())
	rr := httptest.NewRecorder()
	h.AddFavorite(rr, withPrincipal(req, auth.RoleUser))

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status: %d body=%s", rr.Code, rr.Body.String())
	}
}
