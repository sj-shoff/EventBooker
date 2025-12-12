package user_uc

import (
	"context"

	"event-booker/internal/domain"

	"github.com/google/uuid"
)

type UserUsecase struct {
	repo userRepository
}

func NewUserUsecase(repo userRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (uc *UserUsecase) RegisterUser(ctx context.Context, email, telegram string, role domain.UserRole) (*domain.User, error) {
	user := &domain.User{
		ID:       uuid.NewString(),
		Email:    email,
		Telegram: telegram,
		Role:     role,
	}
	return user, uc.repo.Create(ctx, user)
}

func (uc *UserUsecase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return uc.repo.GetByID(ctx, id)
}
