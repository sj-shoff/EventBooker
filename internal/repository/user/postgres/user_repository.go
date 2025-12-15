package user_postgres

import (
	"context"
	"database/sql"

	"event-booker/internal/domain"
	"event-booker/internal/repository"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
)

type UserRepository struct {
	db      *dbpg.DB
	retries retry.Strategy
}

func NewUserRepository(db *dbpg.DB, retries retry.Strategy) *UserRepository {
	return &UserRepository{db: db, retries: retries}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
INSERT INTO users (id, email, telegram, role, created_at)
VALUES ($1, $2, $3, $4, $5)
`
	_, err := r.db.ExecWithRetry(ctx, r.retries, query,
		user.ID, user.Email, user.Telegram, user.Role, user.CreatedAt)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
SELECT id, email, telegram, role, created_at
FROM users WHERE id = $1
`
	row, err := r.db.QueryRowWithRetry(ctx, r.retries, query, id)
	if err != nil {
		return nil, err
	}
	var user domain.User
	err = row.Scan(&user.ID, &user.Email, &user.Telegram, &user.Role, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
