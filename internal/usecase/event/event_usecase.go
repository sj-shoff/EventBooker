package event_uc

import (
	"context"
	"errors"
	"time"

	"event-booker/internal/domain"
	"event-booker/internal/repository"

	"github.com/google/uuid"
)

type EventUsecase struct {
	repo        eventRepository
	bookingRepo bookingRepository
}

func NewEventUsecase(repo eventRepository, bookingRepo bookingRepository) *EventUsecase {
	return &EventUsecase{repo: repo, bookingRepo: bookingRepo}
}

func (uc *EventUsecase) CreateEvent(ctx context.Context, name string, date time.Time, totalSeats int, ttl time.Duration, requiresPayment bool) (*domain.Event, error) {
	now := time.Now()
	event := &domain.Event{
		ID:              uuid.NewString(),
		Name:            name,
		Date:            date,
		TotalSeats:      totalSeats,
		Available:       totalSeats,
		BookingTTL:      ttl,
		RequiresPayment: requiresPayment,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := uc.repo.Create(ctx, event); err != nil {
		return nil, err
	}
	return event, nil
}

func (uc *EventUsecase) GetEvent(ctx context.Context, id string) (*domain.Event, error) {
	event, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEventNotFound
		}
		return nil, err
	}
	return event, nil
}

func (uc *EventUsecase) ListEvents(ctx context.Context) ([]*domain.Event, error) {
	return uc.repo.GetAll(ctx)
}
