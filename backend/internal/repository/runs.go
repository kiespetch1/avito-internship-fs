package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"avito-internship-fs/internal/domain"
)

type RunRepository struct {
	db *sql.DB
}

func NewRunRepository(db *sql.DB) *RunRepository {
	return &RunRepository{db: db}
}

type RunListFilter struct {
	UserID      *uuid.UUID
	AssistantID *uuid.UUID
	Status      *domain.RunStatus
	Limit       int
	Offset      int
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

const runSelectColumns = `
	r.id, r.assistant_id, a.name, a.category_id, c.name,
	r.user_id, r.model, r.user_prompt, r.output, r.status, r.error,
	r.tokens_in, r.tokens_out, r.latency_ms, r.finish_reason, r.created_at`

func (r *RunRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.AssistantRun, error) {
	q := `
		SELECT ` + runSelectColumns + `
		FROM assistant_runs r
		JOIN assistants a ON a.id = r.assistant_id
		JOIN categories c ON c.id = a.category_id
		WHERE r.id = $1`

	var run domain.AssistantRun
	var status string
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&run.ID, &run.AssistantID, &run.AssistantName, &run.CategoryID, &run.CategoryName,
		&run.UserID, &run.Model, &run.UserPrompt, &run.Output, &status, &run.Error,
		&run.TokensIn, &run.TokensOut, &run.LatencyMs, &run.FinishReason, &run.CreatedAt,
	)
	if err != nil {
		return domain.AssistantRun{}, fmt.Errorf("get run: %w", err)
	}
	run.Status = domain.RunStatus(status)

	return run, nil
}

func (r *RunRepository) List(ctx context.Context, f RunListFilter) ([]domain.AssistantRun, int, error) {
	conds := make([]string, 0, 3)
	args := make([]any, 0, 5)

	if f.UserID != nil {
		args = append(args, *f.UserID)
		conds = append(conds, fmt.Sprintf("r.user_id = $%d", len(args)))
	}
	if f.AssistantID != nil {
		args = append(args, *f.AssistantID)
		conds = append(conds, fmt.Sprintf("r.assistant_id = $%d", len(args)))
	}
	if f.Status != nil {
		args = append(args, string(*f.Status))
		conds = append(conds, fmt.Sprintf("r.status = $%d", len(args)))
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	countQ := "SELECT COUNT(*) FROM assistant_runs r " + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count runs: %w", err)
	}

	args = append(args, f.Limit, f.Offset)
	listQ := fmt.Sprintf(`
		SELECT %s
		FROM assistant_runs r
		JOIN assistants a ON a.id = r.assistant_id
		JOIN categories c ON c.id = a.category_id
		%s
		ORDER BY r.created_at DESC, r.id
		LIMIT $%d OFFSET $%d`, runSelectColumns, where, len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query runs: %w", err)
	}
	defer rows.Close()

	out := make([]domain.AssistantRun, 0)
	for rows.Next() {
		var run domain.AssistantRun
		var status string
		if err := rows.Scan(
			&run.ID, &run.AssistantID, &run.AssistantName, &run.CategoryID, &run.CategoryName,
			&run.UserID, &run.Model, &run.UserPrompt, &run.Output, &status, &run.Error,
			&run.TokensIn, &run.TokensOut, &run.LatencyMs, &run.FinishReason, &run.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan run: %w", err)
		}
		run.Status = domain.RunStatus(status)
		out = append(out, run)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate runs: %w", err)
	}

	return out, total, nil
}
