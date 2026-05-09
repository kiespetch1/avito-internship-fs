package service

import (
	"context"

	"avito-internship-fs/internal/domain"
)

type CategoryRepo interface {
	List(ctx context.Context) ([]domain.Category, error)
	Create(ctx context.Context, name string, description *string) (domain.Category, error)
}

type CategoryService struct {
	repo CategoryRepo
}

func NewCategoryService(repo CategoryRepo) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) List(ctx context.Context) ([]domain.Category, error) {
	return s.repo.List(ctx)
}

func (s *CategoryService) Create(ctx context.Context, name string, description *string) (domain.Category, error) {
	return s.repo.Create(ctx, name, description)
}
