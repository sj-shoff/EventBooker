package event_uc

import (
	"context"
	"database/sql"
	"event-booker/internal/domain"
)

type eventRepository interface {
	Create(ctx context.Context, event *domain.Event) error
	GetByID(ctx context.Context, id string) (*domain.Event, error)
	GetForUpdate(ctx context.Context, tx *sql.Tx, id string) (*domain.Event, error)
	GetAll(ctx context.Context) ([]*domain.Event, error)
	GetActiveEvents(ctx context.Context) ([]*domain.Event, error)
	Update(ctx context.Context, tx *sql.Tx, event *domain.Event) error
	Delete(ctx context.Context, id string) error
	DecrementAvailableSeats(ctx context.Context, tx *sql.Tx, id string) error
	IncrementAvailableSeats(ctx context.Context, tx *sql.Tx, id string) error
}

type bookingRepository interface {
	GetByEventID(ctx context.Context, eventID string) ([]*domain.Booking, error)
	Update(ctx context.Context, tx *sql.Tx, booking *domain.Booking) error
}

type userRepository interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

type notifier interface {
	NotifyCancellation(user *domain.User, booking *domain.Booking) error
}
