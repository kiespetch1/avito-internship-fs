package categories

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
	"avito-internship-fs/internal/domain"
)

type fakeService struct {
	listFn   func(ctx context.Context) ([]domain.Category, error)
	createFn func(ctx context.Context, name string, description *string) (domain.Category, error)
}

func (f *fakeService) List(ctx context.Context) ([]domain.Category, error) {
	return f.listFn(ctx)
}

func (f *fakeService) Create(ctx context.Context, name string, description *string) (domain.Category, error) {
	return f.createFn(ctx, name, description)
}

func TestListReturnsCategories(t *testing.T) {
	id := uuid.New()
	desc := "tasty things"
	svc := &fakeService{
		listFn: func(_ context.Context) ([]domain.Category, error) {
			return []domain.Category{
				{ID: id, Name: "Food", Description: &desc, CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
			}, nil
		},
	}
	h := NewHandler(svc)

	rr := httptest.NewRecorder()
	h.List(rr, httptest.NewRequest(http.MethodGet, "/categories", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
	var resp listResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Categories) != 1 || resp.Categories[0].Name != "Food" {
		t.Fatalf("unexpected: %+v", resp)
	}
	if resp.Categories[0].Id != id {
		t.Fatalf("id mismatch: %v", resp.Categories[0].Id)
	}
}

func TestListReturnsEmptyArrayNotNull(t *testing.T) {
	svc := &fakeService{
		listFn: func(_ context.Context) ([]domain.Category, error) { return nil, nil },
	}
	h := NewHandler(svc)
	rr := httptest.NewRecorder()
	h.List(rr, httptest.NewRequest(http.MethodGet, "/categories", nil))

	if !strings.Contains(rr.Body.String(), `"categories":[]`) {
		t.Fatalf("expected empty array in body, got %s", rr.Body.String())
	}
}

func TestCreateRejectsEmptyName(t *testing.T) {
	h := NewHandler(&fakeService{})
	req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{"name":""}`))
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got %d want 400", rr.Code)
	}
}

func TestCreateRejectsWhitespaceName(t *testing.T) {
	h := NewHandler(&fakeService{
		createFn: func(_ context.Context, _ string, _ *string) (domain.Category, error) {
			t.Fatal("service must not be called for whitespace-only name")
			return domain.Category{}, nil
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{"name":"   "}`))
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got %d want 400", rr.Code)
	}
}

func TestCreateTrimsName(t *testing.T) {
	var got string
	svc := &fakeService{
		createFn: func(_ context.Context, name string, _ *string) (domain.Category, error) {
			got = name
			return domain.Category{ID: uuid.New(), Name: name, CreatedAt: time.Now()}, nil
		},
	}
	h := NewHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{"name":"  Food  "}`))
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("got %d want 201", rr.Code)
	}
	if got != "Food" {
		t.Fatalf("name not trimmed: %q", got)
	}
}

func TestCreateRejectsTooLongName(t *testing.T) {
	h := NewHandler(&fakeService{
		createFn: func(_ context.Context, _ string, _ *string) (domain.Category, error) {
			t.Fatal("service must not be called for over-long name")
			return domain.Category{}, nil
		},
	})
	long := strings.Repeat("x", maxCategoryNameLen+1)
	body := `{"name":"` + long + `"}`
	req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(body))
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got %d want 400", rr.Code)
	}
}

func TestCreatedAtSerializedInUTC(t *testing.T) {
	loc := time.FixedZone("MSK", 3*60*60)
	svc := &fakeService{
		createFn: func(_ context.Context, name string, _ *string) (domain.Category, error) {
			return domain.Category{ID: uuid.New(), Name: name, CreatedAt: time.Date(2026, 1, 2, 12, 0, 0, 0, loc)}, nil
		},
	}
	h := NewHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{"name":"Food"}`))
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if !strings.Contains(rr.Body.String(), `"2026-01-02T09:00:00Z"`) {
		t.Fatalf("expected UTC timestamp in body, got %s", rr.Body.String())
	}
}

func TestCreateRejectsMalformedJSON(t *testing.T) {
	h := NewHandler(&fakeService{})
	req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`not json`))
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got %d want 400", rr.Code)
	}
}

func TestCreateMapsNameTakenTo400(t *testing.T) {
	svc := &fakeService{
		createFn: func(_ context.Context, _ string, _ *string) (domain.Category, error) {
			return domain.Category{}, domain.ErrCategoryNameTaken
		},
	}
	h := NewHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{"name":"Food"}`))
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("got %d want 400, body=%s", rr.Code, rr.Body.String())
	}
}

func TestCreateReturns201WithBody(t *testing.T) {
	id := uuid.New()
	svc := &fakeService{
		createFn: func(_ context.Context, name string, _ *string) (domain.Category, error) {
			return domain.Category{ID: id, Name: name, CreatedAt: time.Now()}, nil
		},
	}
	h := NewHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/categories", strings.NewReader(`{"name":"Food"}`))
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("got %d want 201", rr.Code)
	}
	var got api.Category
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Name != "Food" || got.Id != id {
		t.Fatalf("unexpected: %+v", got)
	}
}
