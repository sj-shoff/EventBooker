package booking_uc

import (
	"context"
	"errors"
	"time"

	"event-booker/internal/config"
	"event-booker/internal/domain"
	"event-booker/internal/notification"

	"github.com/google/uuid"
)

type BookingUsecase struct {
	repo      bookingRepository
	eventRepo eventRepository
	notifier  notification.Notifier
	cfg       *config.Config
}

func NewBookingUsecase(repo bookingRepository, eventRepo eventRepository, notifier notification.Notifier, cfg *config.Config) *BookingUsecase {
	return &BookingUsecase{
		repo:      repo,
		eventRepo: eventRepo,
		notifier:  notifier,
		cfg:       cfg,
	}
}

func (uc *BookingUsecase) BookPlace(ctx context.Context, eventID, userID string) (*domain.Booking, error) {
	event, err := uc.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if event.Available <= 0 {
		return nil, errors.New("no seats available")
	}

	ttl := event.BookingTTL
	booking := &domain.Booking{
		ID:        uuid.NewString(),
		EventID:   eventID,
		UserID:    userID,
		Status:    domain.BookingPending,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}

	if err := uc.repo.Create(ctx, booking); err != nil {
		return nil, err
	}

	if err := uc.eventRepo.DecrementAvailableSeats(ctx, eventID); err != nil {
		uc.repo.Delete(ctx, booking.ID)
		return nil, err
	}

	return booking, nil
}

func (uc *BookingUsecase) ConfirmBooking(ctx context.Context, bookingID string) error {
	booking, err := uc.repo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}
	if booking.Status != domain.BookingPending {
		return errors.New("booking not pending")
	}
	if time.Now().After(booking.ExpiresAt) {
		return errors.New("booking expired")
	}

	now := time.Now()
	booking.Status = domain.BookingConfirmed
	booking.ConfirmedAt = &now

	return uc.repo.Update(ctx, booking)
}

func (uc *BookingUsecase) CancelBooking(ctx context.Context, bookingID string) error {
	booking, err := uc.repo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}
	if booking.Status == domain.BookingCancelled {
		return errors.New("already cancelled")
	}

	booking.Status = domain.BookingCancelled
	if err := uc.repo.Update(ctx, booking); err != nil {
		return err
	}

	if err := uc.eventRepo.IncrementAvailableSeats(ctx, booking.EventID); err != nil {
		return err
	}

	user, err := uc.userRepo.GetByID(ctx, booking.UserID)
	if err == nil {
		uc.notifier.NotifyCancellation(user.Email, booking)
	}

	return nil
}

func (uc *BookingUsecase) GetExpiredBookings(ctx context.Context) ([]*domain.Booking, error) {
	return uc.repo.GetExpired(ctx, time.Now())
}
