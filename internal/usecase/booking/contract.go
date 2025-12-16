package booking_uc

import (
	"context"
	"database/sql"
	"time"

	"event-booker/internal/domain"
)

type bookingRepository interface {
	Create(ctx context.Context, tx *sql.Tx, booking *domain.Booking) error
	GetByID(ctx context.Context, id string) (*domain.Booking, error)
	Update(ctx context.Context, tx *sql.Tx, booking *domain.Booking) error
	Delete(ctx context.Context, tx *sql.Tx, id string) error
	GetExpired(ctx context.Context, now time.Time) ([]*domain.Booking, error)
	GetByEventID(ctx context.Context, eventID string) ([]*domain.Booking, error)
	GetAll(ctx context.Context) ([]*domain.Booking, error)
}

type eventRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Event, error)
	GetForUpdate(ctx context.Context, tx *sql.Tx, id string) (*domain.Event, error)
	DecrementAvailableSeats(ctx context.Context, tx *sql.Tx, id string) error
	IncrementAvailableSeats(ctx context.Context, tx *sql.Tx, id string) error
}

type userRepository interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

type notifier interface {
	NotifyCancellation(user *domain.User, booking *domain.Booking) error
}
