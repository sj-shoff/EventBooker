package user_uc

import (
	"context"
	"event-booker/internal/domain"
)

type userRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
}
