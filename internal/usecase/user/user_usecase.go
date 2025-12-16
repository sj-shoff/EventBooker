package user_uc

import (
	"context"
	"errors"
	"time"

	"event-booker/internal/domain"
	"event-booker/internal/repository"

	"github.com/google/uuid"
)

type UserUsecase struct {
	repo userRepository
}

func NewUserUsecase(repo userRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (uc *UserUsecase) RegisterUser(ctx context.Context, email, telegram string, role domain.UserRole) (*domain.User, error) {
	if role != domain.RoleUser && role != domain.RoleAdmin {
		return nil, errors.New("invalid role")
	}
	user := &domain.User{
		ID:        uuid.NewString(),
		Email:     email,
		Telegram:  telegram,
		Role:      role,
		CreatedAt: time.Now(),
	}
	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (uc *UserUsecase) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}
