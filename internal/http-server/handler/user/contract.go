package user

import (
	"context"

	"event-booker/internal/domain"
)

type userUsecase interface {
	RegisterUser(ctx context.Context, email, telegram string, role domain.UserRole) (*domain.User, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
}
