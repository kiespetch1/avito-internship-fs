package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"avito-internship-fs/internal/domain"
)

type fakeCategoryRepo struct {
	listFn   func(ctx context.Context) ([]domain.Category, error)
	createFn func(ctx context.Context, name string, description *string) (domain.Category, error)
}

func (f *fakeCategoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	return f.listFn(ctx)
}

func (f *fakeCategoryRepo) Create(ctx context.Context, name string, description *string) (domain.Category, error) {
	return f.createFn(ctx, name, description)
}

func TestCategoryServiceListPassthrough(t *testing.T) {
	want := []domain.Category{{ID: uuid.New(), Name: "Food", CreatedAt: time.Now()}}
	repo := &fakeCategoryRepo{
		listFn: func(_ context.Context) ([]domain.Category, error) { return want, nil },
	}
	svc := NewCategoryService(repo)

	got, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Food" {
		t.Fatalf("got %+v", got)
	}
}

func TestCategoryServiceCreatePropagatesNameTaken(t *testing.T) {
	repo := &fakeCategoryRepo{
		createFn: func(_ context.Context, _ string, _ *string) (domain.Category, error) {
			return domain.Category{}, domain.ErrCategoryNameTaken
		},
	}
	svc := NewCategoryService(repo)

	_, err := svc.Create(context.Background(), "Food", nil)
	if !errors.Is(err, domain.ErrCategoryNameTaken) {
		t.Fatalf("got %v want ErrCategoryNameTaken", err)
	}
}
