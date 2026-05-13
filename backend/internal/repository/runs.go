package repository

import (
	"context"
	"database/sql"
	"errors"
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

func (r *RunRepository) CreatePendingForActiveAssistant(ctx context.Context, assistantID, userID uuid.UUID, userPrompt string) (domain.Assistant, domain.AssistantRun, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Assistant{}, domain.AssistantRun{}, fmt.Errorf("begin create pending run: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const assistantQ = `
		SELECT id, model, system_prompt, is_active
		FROM assistants
		WHERE id = $1
		FOR UPDATE`

	assistant := domain.Assistant{}
	err = tx.QueryRowContext(ctx, assistantQ, assistantID).
		Scan(&assistant.ID, &assistant.Model, &assistant.SystemPrompt, &assistant.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Assistant{}, domain.AssistantRun{}, domain.ErrAssistantNotFound
		}

		return domain.Assistant{}, domain.AssistantRun{}, fmt.Errorf("load assistant for run: %w", err)
	}
	if !assistant.IsActive {
		return domain.Assistant{}, domain.AssistantRun{}, domain.ErrAssistantInactive
	}

	const insertQ = `
		INSERT INTO assistant_runs (assistant_id, user_id, model, user_prompt, status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING id, created_at`

	run := domain.AssistantRun{
		AssistantID: assistantID,
		UserID:      userID,
		Model:       assistant.Model,
		UserPrompt:  userPrompt,
		Status:      domain.RunPending,
	}
	err = tx.QueryRowContext(ctx, insertQ, assistantID, userID, assistant.Model, userPrompt).
		Scan(&run.ID, &run.CreatedAt)
	if err != nil {
		return domain.Assistant{}, domain.AssistantRun{}, fmt.Errorf("insert run: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.Assistant{}, domain.AssistantRun{}, fmt.Errorf("commit create pending run: %w", err)
	}

	return assistant, run, nil
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
	r.tokens_in, r.tokens_out, r.latency_ms, r.finish_reason, r.created_at,
	rf.rating`

func scanRun(s interface{ Scan(...any) error }, run *domain.AssistantRun) error {
	var status string
	var feedbackRating sql.NullInt16
	err := s.Scan(
		&run.ID, &run.AssistantID, &run.AssistantName, &run.CategoryID, &run.CategoryName,
		&run.UserID, &run.Model, &run.UserPrompt, &run.Output, &status, &run.Error,
		&run.TokensIn, &run.TokensOut, &run.LatencyMs, &run.FinishReason, &run.CreatedAt,
		&feedbackRating,
	)
	if err != nil {
		return err
	}

	run.Status = domain.RunStatus(status)
	if feedbackRating.Valid {
		run.FeedbackRating = new(domain.RunFeedbackRating(feedbackRating.Int16))
	} else {
		run.FeedbackRating = nil
	}

	return nil
}

func (r *RunRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.AssistantRun, error) {
	q := `
		SELECT ` + runSelectColumns + `
		FROM assistant_runs r
		JOIN assistants a ON a.id = r.assistant_id
		JOIN categories c ON c.id = a.category_id
		LEFT JOIN run_feedback rf ON rf.run_id = r.id
		WHERE r.id = $1`

	var run domain.AssistantRun
	if err := scanRun(r.db.QueryRowContext(ctx, q, id), &run); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.AssistantRun{}, domain.ErrRunNotFound
		}

		return domain.AssistantRun{}, fmt.Errorf("get run: %w", err)
	}

	return run, nil
}

func (r *RunRepository) UpsertFeedback(ctx context.Context, runID uuid.UUID, rating domain.RunFeedbackRating) (domain.AssistantRun, error) {
	const q = `
		INSERT INTO run_feedback (run_id, rating)
		VALUES ($1, $2)
		ON CONFLICT (run_id) DO UPDATE SET rating = EXCLUDED.rating
		RETURNING run_id`

	var id uuid.UUID
	if err := r.db.QueryRowContext(ctx, q, runID, int(rating)).Scan(&id); err != nil {
		return domain.AssistantRun{}, fmt.Errorf("upsert run feedback: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *RunRepository) List(ctx context.Context, f RunListFilter) ([]domain.AssistantRun, int, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("begin list runs: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

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
	if err := tx.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count runs: %w", err)
	}

	args = append(args, f.Limit, f.Offset)
	listQ := fmt.Sprintf(`
		SELECT %s
		FROM assistant_runs r
		JOIN assistants a ON a.id = r.assistant_id
		JOIN categories c ON c.id = a.category_id
		LEFT JOIN run_feedback rf ON rf.run_id = r.id
		%s
		ORDER BY r.created_at DESC, r.id
		LIMIT $%d OFFSET $%d`, runSelectColumns, where, len(args)-1, len(args))

	rows, err := tx.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query runs: %w", err)
	}
	defer rows.Close()

	out := make([]domain.AssistantRun, 0)
	for rows.Next() {
		var run domain.AssistantRun
		if err := scanRun(rows, &run); err != nil {
			return nil, 0, fmt.Errorf("scan run: %w", err)
		}
		out = append(out, run)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate runs: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, 0, fmt.Errorf("commit list runs: %w", err)
	}

	return out, total, nil
}
