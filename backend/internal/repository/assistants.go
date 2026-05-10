package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"avito-internship-fs/internal/domain"
)

type AssistantRepository struct {
	db *sql.DB
}

func NewAssistantRepository(db *sql.DB) *AssistantRepository {
	return &AssistantRepository{db: db}
}

func (r *AssistantRepository) Get(ctx context.Context, id uuid.UUID) (domain.Assistant, error) {
	const q = `
		SELECT id, category_id, name, description, model, system_prompt,
		       example_user_prompt, is_active, created_at, updated_at
		FROM assistants
		WHERE id = $1`

	var assistant domain.Assistant
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&assistant.ID, &assistant.CategoryID, &assistant.Name, &assistant.Description, &assistant.Model,
		&assistant.SystemPrompt, &assistant.ExampleUserPrompt, &assistant.IsActive, &assistant.CreatedAt,
		&assistant.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Assistant{}, domain.ErrAssistantNotFound
		}

		return domain.Assistant{}, fmt.Errorf("get assistant: %w", err)
	}

	return assistant, nil
}
