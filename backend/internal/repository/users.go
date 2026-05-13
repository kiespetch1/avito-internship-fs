package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"avito-internship-fs/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, email, passwordHash string) (domain.User, error) {
	const q = `
		INSERT INTO users (email, password, role)
		VALUES ($1, $2, 'user')
		RETURNING id, email, password, role, created_at`

	var u domain.User
	err := r.db.QueryRowContext(ctx, q, email, passwordHash).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return domain.User{}, domain.ErrUserEmailTaken
		}

		return domain.User{}, fmt.Errorf("insert user: %w", err)
	}

	return u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	const q = `
		SELECT id, email, password, role, created_at
		FROM users
		WHERE email = $1`

	var u domain.User
	err := r.db.QueryRowContext(ctx, q, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}

		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}

	return u, nil
}
