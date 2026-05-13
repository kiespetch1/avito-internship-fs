package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"avito-internship-fs/internal/domain"
	"avito-internship-fs/internal/repository"
)

type fakeAssistantWriteRepo struct {
	getFn         func(ctx context.Context, id uuid.UUID) (domain.Assistant, error)
	getForUserFn  func(ctx context.Context, userID, id uuid.UUID) (domain.Assistant, error)
	createFn      func(ctx context.Context, a domain.Assistant) (domain.Assistant, error)
	updateFn      func(ctx context.Context, a domain.Assistant) (domain.Assistant, error)
	listFn        func(ctx context.Context, f repository.AssistantListFilter) ([]domain.Assistant, int, error)
	setFavoriteFn func(ctx context.Context, userID, assistantID uuid.UUID, favorite bool) error
}

func (f *fakeAssistantWriteRepo) Get(ctx context.Context, id uuid.UUID) (domain.Assistant, error) {
	return f.getFn(ctx, id)
}

func (f *fakeAssistantWriteRepo) GetForUser(ctx context.Context, userID, id uuid.UUID) (domain.Assistant, error) {
	return f.getForUserFn(ctx, userID, id)
}

func (f *fakeAssistantWriteRepo) Create(ctx context.Context, a domain.Assistant) (domain.Assistant, error) {
	return f.createFn(ctx, a)
}

func (f *fakeAssistantWriteRepo) Update(ctx context.Context, a domain.Assistant) (domain.Assistant, error) {
	return f.updateFn(ctx, a)
}

func (f *fakeAssistantWriteRepo) List(ctx context.Context, fl repository.AssistantListFilter) ([]domain.Assistant, int, error) {
	return f.listFn(ctx, fl)
}

func (f *fakeAssistantWriteRepo) SetFavorite(ctx context.Context, userID, assistantID uuid.UUID, favorite bool) error {
	return f.setFavoriteFn(ctx, userID, assistantID, favorite)
}

func TestAssistantServiceGetPassthrough(t *testing.T) {
	want := domain.Assistant{ID: uuid.New(), Name: "x"}
	userID := uuid.New()
	var capturedUserID uuid.UUID
	repo := &fakeAssistantWriteRepo{
		getForUserFn: func(_ context.Context, gotUserID, _ uuid.UUID) (domain.Assistant, error) {
			capturedUserID = gotUserID
			return want, nil
		},
	}
	got, err := NewAssistantService(repo).Get(context.Background(), userID, want.ID, true)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if got.ID != want.ID {
		t.Fatalf("got %v want %v", got.ID, want.ID)
	}
	if capturedUserID != userID {
		t.Fatalf("user id not propagated: got %v want %v", capturedUserID, userID)
	}
}

func TestAssistantServiceGetHidesInactiveFromUser(t *testing.T) {
	inactive := domain.Assistant{ID: uuid.New(), IsActive: false}
	repo := &fakeAssistantWriteRepo{
		getForUserFn: func(_ context.Context, _, _ uuid.UUID) (domain.Assistant, error) {
			return inactive, nil
		},
	}
	_, err := NewAssistantService(repo).Get(context.Background(), uuid.New(), inactive.ID, false)
	if !errors.Is(err, domain.ErrAssistantNotFound) {
		t.Fatalf("expected ErrAssistantNotFound, got %v", err)
	}

	got, err := NewAssistantService(repo).Get(context.Background(), uuid.New(), inactive.ID, true)
	if err != nil {
		t.Fatalf("admin must see inactive: %v", err)
	}
	if got.IsActive {
		t.Fatalf("expected inactive assistant")
	}
}

func TestAssistantServiceGetPropagatesNotFound(t *testing.T) {
	repo := &fakeAssistantWriteRepo{
		getForUserFn: func(_ context.Context, _, _ uuid.UUID) (domain.Assistant, error) {
			return domain.Assistant{}, domain.ErrAssistantNotFound
		},
	}
	_, err := NewAssistantService(repo).Get(context.Background(), uuid.New(), uuid.New(), true)
	if !errors.Is(err, domain.ErrAssistantNotFound) {
		t.Fatalf("err: %v", err)
	}
}

func TestAssistantServiceCreateMapsFields(t *testing.T) {
	categoryID := uuid.New()
	example := "ex"
	var captured domain.Assistant
	repo := &fakeAssistantWriteRepo{
		createFn: func(_ context.Context, a domain.Assistant) (domain.Assistant, error) {
			captured = a
			a.ID = uuid.New()

			return a, nil
		},
	}
	in := AssistantCreateInput{
		CategoryID:        categoryID,
		Name:              "Повар",
		Description:       "desc",
		Model:             "gpt",
		Tags:              []string{"еда", "рецепты"},
		SystemPrompt:      "sys",
		ExampleUserPrompt: &example,
		IsActive:          true,
	}
	out, err := NewAssistantService(repo).Create(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if out.ID == (uuid.UUID{}) {
		t.Fatalf("expected generated id")
	}
	if captured.CategoryID != categoryID || captured.Name != "Повар" || captured.SystemPrompt != "sys" || !captured.IsActive {
		t.Fatalf("captured: %+v", captured)
	}
	if captured.ExampleUserPrompt == nil || *captured.ExampleUserPrompt != "ex" {
		t.Fatalf("example prompt not propagated: %+v", captured.ExampleUserPrompt)
	}
	if len(captured.Tags) != 2 || captured.Tags[0] != "еда" || captured.Tags[1] != "рецепты" {
		t.Fatalf("tags not propagated: %+v", captured.Tags)
	}
}

func TestAssistantServiceUpdateMapsFields(t *testing.T) {
	id := uuid.New()
	categoryID := uuid.New()
	var captured domain.Assistant
	repo := &fakeAssistantWriteRepo{
		updateFn: func(_ context.Context, a domain.Assistant) (domain.Assistant, error) {
			captured = a
			return a, nil
		},
	}
	_, err := NewAssistantService(repo).Update(context.Background(), AssistantUpdateInput{
		ID:           id,
		CategoryID:   categoryID,
		Name:         "renamed",
		Description:  "d",
		Model:        "gpt",
		Tags:         []string{"спорт"},
		SystemPrompt: "sys",
		IsActive:     false,
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if captured.ID != id || captured.CategoryID != categoryID || captured.Name != "renamed" || captured.IsActive {
		t.Fatalf("captured: %+v", captured)
	}
	if len(captured.Tags) != 1 || captured.Tags[0] != "спорт" {
		t.Fatalf("tags: %+v", captured.Tags)
	}
}

func TestAssistantServiceListNormalizesPagination(t *testing.T) {
	var captured repository.AssistantListFilter
	repo := &fakeAssistantWriteRepo{
		listFn: func(_ context.Context, f repository.AssistantListFilter) ([]domain.Assistant, int, error) {
			captured = f
			return []domain.Assistant{}, 0, nil
		},
	}
	_, _, err := NewAssistantService(repo).List(context.Background(), AssistantListInput{
		Page:     0,
		PageSize: 0,
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if captured.Limit != defaultAssistantsPageSize {
		t.Fatalf("default page size not applied: got limit=%d want=%d", captured.Limit, defaultAssistantsPageSize)
	}
	if captured.Offset != 0 {
		t.Fatalf("offset for page=1 must be 0, got %d", captured.Offset)
	}
}

func TestAssistantServiceListClampsPageSize(t *testing.T) {
	var captured repository.AssistantListFilter
	repo := &fakeAssistantWriteRepo{
		listFn: func(_ context.Context, f repository.AssistantListFilter) ([]domain.Assistant, int, error) {
			captured = f
			return nil, 0, nil
		},
	}
	_, _, _ = NewAssistantService(repo).List(context.Background(), AssistantListInput{
		Page:     2,
		PageSize: maxPageSize + 50,
	})
	if captured.Limit != maxPageSize {
		t.Fatalf("page size must be clamped to %d, got %d", maxPageSize, captured.Limit)
	}
	if captured.Offset != maxPageSize {
		t.Fatalf("offset for page=2 must be %d, got %d", maxPageSize, captured.Offset)
	}
}

func TestAssistantServiceListPassesFilters(t *testing.T) {
	userID := uuid.New()
	categoryID := uuid.New()
	q := "повар"
	tag := "еда"
	var captured repository.AssistantListFilter
	repo := &fakeAssistantWriteRepo{
		listFn: func(_ context.Context, f repository.AssistantListFilter) ([]domain.Assistant, int, error) {
			captured = f
			return nil, 0, nil
		},
	}
	_, _, _ = NewAssistantService(repo).List(context.Background(), AssistantListInput{
		UserID:          userID,
		CategoryID:      &categoryID,
		Query:           &q,
		Tag:             &tag,
		IncludeInactive: true,
		FavoriteOnly:    true,
		Page:            1,
		PageSize:        10,
	})
	if captured.UserID == nil || *captured.UserID != userID {
		t.Fatalf("user not propagated: %+v", captured.UserID)
	}
	if captured.CategoryID == nil || *captured.CategoryID != categoryID {
		t.Fatalf("category not propagated: %+v", captured.CategoryID)
	}
	if captured.Query == nil || *captured.Query != "повар" {
		t.Fatalf("query not propagated: %+v", captured.Query)
	}
	if captured.Tag == nil || *captured.Tag != "еда" {
		t.Fatalf("tag not propagated: %+v", captured.Tag)
	}
	if !captured.IncludeInactive {
		t.Fatalf("IncludeInactive not propagated")
	}
	if !captured.FavoriteOnly {
		t.Fatalf("FavoriteOnly not propagated")
	}
}

func TestAssistantServiceSetFavoritePassthrough(t *testing.T) {
	userID := uuid.New()
	assistantID := uuid.New()
	var capturedFavorite bool
	repo := &fakeAssistantWriteRepo{
		setFavoriteFn: func(_ context.Context, gotUserID, gotAssistantID uuid.UUID, favorite bool) error {
			if gotUserID != userID {
				t.Fatalf("user id: got %v want %v", gotUserID, userID)
			}
			if gotAssistantID != assistantID {
				t.Fatalf("assistant id: got %v want %v", gotAssistantID, assistantID)
			}
			capturedFavorite = favorite

			return nil
		},
	}

	if err := NewAssistantService(repo).SetFavorite(context.Background(), userID, assistantID, true); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !capturedFavorite {
		t.Fatalf("favorite flag not propagated")
	}
}
