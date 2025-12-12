package event_uc

import (
	"context"
	"time"

	"event-booker/internal/domain"

	"github.com/google/uuid"
)

type EventUsecase struct {
	repo        EventRepository
	bookingRepo BookingRepository
}

func NewEventUsecase(repo EventRepository, bookingRepo BookingRepository) *EventUsecase {
	return &EventUsecase{repo: repo, bookingRepo: bookingRepo}
}

func (uc *EventUsecase) CreateEvent(ctx context.Context, name string, date time.Time, totalSeats int, ttl time.Duration, requiresPayment bool) (*domain.Event, error) {
	event := &domain.Event{
		ID:              uuid.NewString(),
		Name:            name,
		Date:            date,
		TotalSeats:      totalSeats,
		Available:       totalSeats,
		BookingTTL:      ttl,
		RequiresPayment: requiresPayment,
	}
	return event, uc.repo.Create(ctx, event)
}

func (uc *EventUsecase) GetEvent(ctx context.Context, id string) (*domain.Event, error) {
	event, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	bookings, err := uc.bookingRepo.GetByEventID(ctx, id)
	if err != nil {
		return nil, err
	}
	event.Available = event.TotalSeats - len(bookings) // approximate, assuming no concurrent
	return event, nil
}
