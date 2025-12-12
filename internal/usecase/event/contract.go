package event_uc

import (
	"context"
	"event-booker/internal/domain"
)

type EventRepository interface {
	Create(ctx context.Context, event *domain.Event) error
	GetByID(ctx context.Context, id string) (*domain.Event, error)
}

type BookingRepository interface {
	GetByEventID(ctx context.Context, eventID string) ([]*domain.Booking, error)
}
