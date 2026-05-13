package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"avito-internship-fs/internal/domain"
	"avito-internship-fs/internal/repository"
)

type capturingRunRepo struct {
	*fakeRunRepo
	captured repository.RunListFilter
	items    []domain.AssistantRun
	total    int
}

func (r *capturingRunRepo) List(_ context.Context, f repository.RunListFilter) ([]domain.AssistantRun, int, error) {
	r.captured = f
	return r.items, r.total, nil
}

func TestRunServiceListNormalizesPagination(t *testing.T) {
	repo := &capturingRunRepo{fakeRunRepo: newFakeRunRepo()}
	svc := NewRunService(repo, &fakeProvider{}, 0)
	_, _, err := svc.List(context.Background(), RunListInput{Page: 0, PageSize: 0})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if repo.captured.Limit != defaultRunsPageSize {
		t.Fatalf("default page size: got %d want %d", repo.captured.Limit, defaultRunsPageSize)
	}
	if repo.captured.Offset != 0 {
		t.Fatalf("offset for page=1 must be 0, got %d", repo.captured.Offset)
	}
}

func TestRunServiceListPassesFilters(t *testing.T) {
	repo := &capturingRunRepo{fakeRunRepo: newFakeRunRepo()}
	svc := NewRunService(repo, &fakeProvider{}, 0)

	userID := uuid.New()
	assistantID := uuid.New()
	status := domain.RunSuccess
	_, _, _ = svc.List(context.Background(), RunListInput{
		UserID:      &userID,
		AssistantID: &assistantID,
		Status:      &status,
		Page:        2,
		PageSize:    5,
	})

	if repo.captured.UserID == nil || *repo.captured.UserID != userID {
		t.Fatalf("UserID not propagated: %+v", repo.captured.UserID)
	}
	if repo.captured.AssistantID == nil || *repo.captured.AssistantID != assistantID {
		t.Fatalf("AssistantID not propagated: %+v", repo.captured.AssistantID)
	}
	if repo.captured.Status == nil || *repo.captured.Status != domain.RunSuccess {
		t.Fatalf("Status not propagated: %+v", repo.captured.Status)
	}
	if repo.captured.Limit != 5 || repo.captured.Offset != 5 {
		t.Fatalf("pagination: limit=%d offset=%d", repo.captured.Limit, repo.captured.Offset)
	}
}
