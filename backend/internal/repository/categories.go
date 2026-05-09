package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"avito-internship-fs/internal/domain"
)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	const q = `SELECT id, name, description, created_at FROM categories ORDER BY name`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Category, 0)
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}

	return out, nil
}

func (r *CategoryRepository) Create(ctx context.Context, name string, description *string) (domain.Category, error) {
	const q = `
		INSERT INTO categories (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at`

	var c domain.Category
	err := r.db.QueryRowContext(ctx, q, name, description).
		Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return domain.Category{}, domain.ErrCategoryNameTaken
		}

		return domain.Category{}, fmt.Errorf("insert category: %w", err)
	}

	return c, nil
}

func (r *CategoryRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	const q = `SELECT 1 FROM categories WHERE id = $1`

	var x int
	err := r.db.QueryRowContext(ctx, q, id).Scan(&x)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("check category: %w", err)
	}

	return true, nil
}
