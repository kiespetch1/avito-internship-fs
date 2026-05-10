package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"avito-internship-fs/internal/domain"
)

type RunRepository struct {
	db *sql.DB
}

func NewRunRepository(db *sql.DB) *RunRepository {
	return &RunRepository{db: db}
}

func (r *RunRepository) CreatePending(ctx context.Context, assistantID, userID uuid.UUID, model, userPrompt string) (domain.AssistantRun, error) {
	const q = `
		INSERT INTO assistant_runs (assistant_id, user_id, model, user_prompt, status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING id, created_at`

	run := domain.AssistantRun{
		AssistantID: assistantID,
		UserID:      userID,
		Model:       model,
		UserPrompt:  userPrompt,
		Status:      domain.RunPending,
	}
	err := r.db.QueryRowContext(ctx, q, assistantID, userID, model, userPrompt).
		Scan(&run.ID, &run.CreatedAt)
	if err != nil {
		return domain.AssistantRun{}, fmt.Errorf("insert run: %w", err)
	}

	return run, nil
}

func (r *RunRepository) MarkSuccess(ctx context.Context, id uuid.UUID, output string, tokensIn, tokensOut, latencyMs int, finishReason string) error {
	const q = `
		UPDATE assistant_runs
		SET status = 'success', output = $2, tokens_in = $3, tokens_out = $4,
		    latency_ms = $5, finish_reason = $6, error = NULL
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, q, id, output, tokensIn, tokensOut, latencyMs, finishReason)
	if err != nil {
		return fmt.Errorf("mark run success: %w", err)
	}

	return nil
}

func (r *RunRepository) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string, latencyMs int) error {
	const q = `
		UPDATE assistant_runs
		SET status = 'failed', error = $2, latency_ms = $3
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, q, id, errMsg, latencyMs)
	if err != nil {
		return fmt.Errorf("mark run failed: %w", err)
	}

	return nil
}

func (r *RunRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.AssistantRun, error) {
	const q = `
		SELECT id, assistant_id, user_id, model, user_prompt, output, status, error,
		       tokens_in, tokens_out, latency_ms, finish_reason, created_at
		FROM assistant_runs
		WHERE id = $1`

	var run domain.AssistantRun
	var status string
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&run.ID, &run.AssistantID, &run.UserID, &run.Model, &run.UserPrompt,
		&run.Output, &status, &run.Error,
		&run.TokensIn, &run.TokensOut, &run.LatencyMs, &run.FinishReason, &run.CreatedAt,
	)
	if err != nil {
		return domain.AssistantRun{}, fmt.Errorf("get run: %w", err)
	}
	run.Status = domain.RunStatus(status)

	return run, nil
}
