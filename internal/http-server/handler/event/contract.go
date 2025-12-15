package event

import (
	"context"

	"event-booker/internal/domain"
	"time"
)

type eventUsecase interface {
	CreateEvent(ctx context.Context, name string, date time.Time, totalSeats int, ttl time.Duration, requiresPayment bool) (*domain.Event, error)
	GetEvent(ctx context.Context, id string) (*domain.Event, error)
	ListEvents(ctx context.Context) ([]*domain.Event, error)
}
