package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"avito-internship-fs/internal/domain"
)

type AssistantRepository struct {
	db *sql.DB
}

func NewAssistantRepository(db *sql.DB) *AssistantRepository {
	return &AssistantRepository{db: db}
}

type AssistantListFilter struct {
	UserID          *uuid.UUID
	CategoryID      *uuid.UUID
	Query           *string
	IncludeInactive bool
	FavoriteOnly    bool
	Limit           int
	Offset          int
}

const assistantBaseSelectColumns = `
	a.id, a.category_id, c.name, a.name, a.description, a.model, a.system_prompt,
	a.example_user_prompt, a.is_active, a.created_at, a.updated_at`

func assistantSelectColumns(favoriteExpr string) string {
	return assistantBaseSelectColumns + ", " + favoriteExpr
}

func scanAssistant(s interface{ Scan(...any) error }, a *domain.Assistant) error {
	return s.Scan(
		&a.ID, &a.CategoryID, &a.CategoryName, &a.Name, &a.Description, &a.Model, &a.SystemPrompt,
		&a.ExampleUserPrompt, &a.IsActive, &a.CreatedAt, &a.UpdatedAt, &a.IsFavorite,
	)
}

func (r *AssistantRepository) Get(ctx context.Context, id uuid.UUID) (domain.Assistant, error) {
	q := `
		SELECT ` + assistantSelectColumns("FALSE") + `
		FROM assistants a
		JOIN categories c ON c.id = a.category_id
		WHERE a.id = $1`

	var a domain.Assistant
	if err := scanAssistant(r.db.QueryRowContext(ctx, q, id), &a); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Assistant{}, domain.ErrAssistantNotFound
		}

		return domain.Assistant{}, fmt.Errorf("get assistant: %w", err)
	}

	return a, nil
}

func (r *AssistantRepository) GetForUser(ctx context.Context, userID, id uuid.UUID) (domain.Assistant, error) {
	q := `
		SELECT ` + assistantSelectColumns(`EXISTS (
			SELECT 1
			FROM favorite_assistants fa
			WHERE fa.assistant_id = a.id AND fa.user_id = $2
		)`) + `
		FROM assistants a
		JOIN categories c ON c.id = a.category_id
		WHERE a.id = $1`

	var a domain.Assistant
	if err := scanAssistant(r.db.QueryRowContext(ctx, q, id, userID), &a); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Assistant{}, domain.ErrAssistantNotFound
		}

		return domain.Assistant{}, fmt.Errorf("get assistant for user: %w", err)
	}

	return a, nil
}

func (r *AssistantRepository) Create(ctx context.Context, in domain.Assistant) (domain.Assistant, error) {
	const insertQ = `
		INSERT INTO assistants (category_id, name, description, model, system_prompt, example_user_prompt, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, insertQ,
		in.CategoryID, in.Name, in.Description, in.Model, in.SystemPrompt, in.ExampleUserPrompt, in.IsActive,
	).Scan(&id)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23503" {
			return domain.Assistant{}, domain.ErrCategoryNotFound
		}

		return domain.Assistant{}, fmt.Errorf("insert assistant: %w", err)
	}

	return r.Get(ctx, id)
}

func (r *AssistantRepository) Update(ctx context.Context, in domain.Assistant) (domain.Assistant, error) {
	const updateQ = `
		UPDATE assistants
		SET category_id = $2, name = $3, description = $4, model = $5,
		    system_prompt = $6, example_user_prompt = $7, is_active = $8, updated_at = now()
		WHERE id = $1
		RETURNING id`

	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, updateQ,
		in.ID, in.CategoryID, in.Name, in.Description, in.Model, in.SystemPrompt, in.ExampleUserPrompt, in.IsActive,
	).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Assistant{}, domain.ErrAssistantNotFound
		}
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23503" {
			return domain.Assistant{}, domain.ErrCategoryNotFound
		}

		return domain.Assistant{}, fmt.Errorf("update assistant: %w", err)
	}

	return r.Get(ctx, id)
}

func (r *AssistantRepository) List(ctx context.Context, f AssistantListFilter) ([]domain.Assistant, int, error) {
	conds := make([]string, 0, 4)
	args := make([]any, 0, 6)

	if !f.IncludeInactive {
		conds = append(conds, "a.is_active = TRUE")
	}
	if f.CategoryID != nil {
		args = append(args, *f.CategoryID)
		conds = append(conds, fmt.Sprintf("a.category_id = $%d", len(args)))
	}
	if f.Query != nil && strings.TrimSpace(*f.Query) != "" {
		args = append(args, "%"+strings.TrimSpace(*f.Query)+"%")
		conds = append(conds, fmt.Sprintf("(a.name ILIKE $%d OR a.description ILIKE $%d)", len(args), len(args)))
	}
	if f.FavoriteOnly {
		if f.UserID == nil {
			conds = append(conds, "FALSE")
		} else {
			args = append(args, *f.UserID)
			conds = append(conds, fmt.Sprintf(`EXISTS (
				SELECT 1
				FROM favorite_assistants fa_filter
				WHERE fa_filter.assistant_id = a.id AND fa_filter.user_id = $%d
			)`, len(args)))
		}
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	countArgs := append([]any(nil), args...)
	countQ := "SELECT COUNT(*) FROM assistants a " + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQ, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count assistants: %w", err)
	}

	listArgs := append([]any(nil), args...)
	favoriteExpr := "FALSE"
	if f.UserID != nil {
		listArgs = append(listArgs, *f.UserID)
		favoriteExpr = fmt.Sprintf(`EXISTS (
			SELECT 1
			FROM favorite_assistants fa
			WHERE fa.assistant_id = a.id AND fa.user_id = $%d
		)`, len(listArgs))
	}

	listArgs = append(listArgs, f.Limit, f.Offset)
	listQ := fmt.Sprintf(`
		SELECT %s
		FROM assistants a
		JOIN categories c ON c.id = a.category_id
		%s
		ORDER BY a.created_at DESC, a.id
		LIMIT $%d OFFSET $%d`, assistantSelectColumns(favoriteExpr), where, len(listArgs)-1, len(listArgs))

	rows, err := r.db.QueryContext(ctx, listQ, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query assistants: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Assistant, 0)
	for rows.Next() {
		var a domain.Assistant
		if err := scanAssistant(rows, &a); err != nil {
			return nil, 0, fmt.Errorf("scan assistant: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate assistants: %w", err)
	}

	return out, total, nil
}

func (r *AssistantRepository) SetFavorite(ctx context.Context, userID, assistantID uuid.UUID, favorite bool) error {
	if favorite {
		const q = `
			INSERT INTO favorite_assistants (user_id, assistant_id)
			VALUES ($1, $2)
			ON CONFLICT (user_id, assistant_id) DO NOTHING`
		if _, err := r.db.ExecContext(ctx, q, userID, assistantID); err != nil {
			if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23503" {
				return domain.ErrAssistantNotFound
			}

			return fmt.Errorf("add favorite assistant: %w", err)
		}

		return nil
	}

	const deleteQ = `
		DELETE FROM favorite_assistants
		WHERE user_id = $1 AND assistant_id = $2`
	res, err := r.db.ExecContext(ctx, deleteQ, userID, assistantID)
	if err != nil {
		return fmt.Errorf("remove favorite assistant: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("remove favorite rows affected: %w", err)
	}
	if affected > 0 {
		return nil
	}

	var exists bool
	if err := r.db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM assistants WHERE id = $1)", assistantID).Scan(&exists); err != nil {
		return fmt.Errorf("check assistant exists for favorite removal: %w", err)
	}
	if !exists {
		return domain.ErrAssistantNotFound
	}

	return nil
}
