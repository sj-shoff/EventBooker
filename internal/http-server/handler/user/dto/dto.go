package dto

import "event-booker/internal/domain"

type RegisterRequest struct {
	Email    string          `json:"email"`
	Telegram string          `json:"telegram"`
	Role     domain.UserRole `json:"role"`
}
