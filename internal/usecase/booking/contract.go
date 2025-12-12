package booking_uc

import (
	"context"
	"event-booker/internal/domain"
	"time"
)

type bookingRepository interface {
	Create(ctx context.Context, booking *domain.Booking) error
	GetByID(ctx context.Context, id string) (*domain.Booking, error)
	Update(ctx context.Context, booking *domain.Booking) error
	Delete(ctx context.Context, id string) error
	GetExpired(ctx context.Context, now time.Time) ([]*domain.Booking, error)
}

type eventRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Event, error)
	DecrementAvailableSeats(ctx context.Context, id string) error
	IncrementAvailableSeats(ctx context.Context, id string) error
}
