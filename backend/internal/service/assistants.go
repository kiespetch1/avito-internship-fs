package service

import (
	"context"

	"github.com/google/uuid"

	"avito-internship-fs/internal/domain"
	"avito-internship-fs/internal/repository"
)

type AssistantWriteRepo interface {
	Get(ctx context.Context, id uuid.UUID) (domain.Assistant, error)
	GetForUser(ctx context.Context, userID, id uuid.UUID) (domain.Assistant, error)
	Create(ctx context.Context, in domain.Assistant) (domain.Assistant, error)
	Update(ctx context.Context, in domain.Assistant) (domain.Assistant, error)
	List(ctx context.Context, f repository.AssistantListFilter) ([]domain.Assistant, int, error)
	SetFavorite(ctx context.Context, userID, assistantID uuid.UUID, favorite bool) error
}

type AssistantService struct {
	repo AssistantWriteRepo
}

func NewAssistantService(repo AssistantWriteRepo) *AssistantService {
	return &AssistantService{repo: repo}
}

type AssistantCreateInput struct {
	CategoryID        uuid.UUID
	Name              string
	Description       string
	Model             string
	Tags              []string
	SystemPrompt      string
	ExampleUserPrompt *string
	IsActive          bool
}

type AssistantUpdateInput struct {
	ID                uuid.UUID
	CategoryID        uuid.UUID
	Name              string
	Description       string
	Model             string
	Tags              []string
	SystemPrompt      string
	ExampleUserPrompt *string
	IsActive          bool
}

type AssistantListInput struct {
	UserID          uuid.UUID
	CategoryID      *uuid.UUID
	Query           *string
	Tag             *string
	IncludeInactive bool
	FavoriteOnly    bool
	Page            int
	PageSize        int
}

func (s *AssistantService) Get(ctx context.Context, userID, id uuid.UUID) (domain.Assistant, error) {
	return s.repo.GetForUser(ctx, userID, id)
}

func (s *AssistantService) Create(ctx context.Context, in AssistantCreateInput) (domain.Assistant, error) {
	return s.repo.Create(ctx, domain.Assistant{
		CategoryID:        in.CategoryID,
		Name:              in.Name,
		Description:       in.Description,
		Model:             in.Model,
		Tags:              in.Tags,
		SystemPrompt:      in.SystemPrompt,
		ExampleUserPrompt: in.ExampleUserPrompt,
		IsActive:          in.IsActive,
	})
}

func (s *AssistantService) Update(ctx context.Context, in AssistantUpdateInput) (domain.Assistant, error) {
	return s.repo.Update(ctx, domain.Assistant{
		ID:                in.ID,
		CategoryID:        in.CategoryID,
		Name:              in.Name,
		Description:       in.Description,
		Model:             in.Model,
		Tags:              in.Tags,
		SystemPrompt:      in.SystemPrompt,
		ExampleUserPrompt: in.ExampleUserPrompt,
		IsActive:          in.IsActive,
	})
}

func (s *AssistantService) List(ctx context.Context, in AssistantListInput) ([]domain.Assistant, int, error) {
	page, pageSize := normalizePage(in.Page, in.PageSize, defaultAssistantsPageSize, maxPageSize)

	return s.repo.List(ctx, repository.AssistantListFilter{
		UserID:          &in.UserID,
		CategoryID:      in.CategoryID,
		Query:           in.Query,
		Tag:             in.Tag,
		IncludeInactive: in.IncludeInactive,
		FavoriteOnly:    in.FavoriteOnly,
		Limit:           pageSize,
		Offset:          (page - 1) * pageSize,
	})
}

func (s *AssistantService) SetFavorite(ctx context.Context, userID, assistantID uuid.UUID, favorite bool) error {
	return s.repo.SetFavorite(ctx, userID, assistantID, favorite)
}
