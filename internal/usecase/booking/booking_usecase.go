package booking_uc

import (
	"context"
	"errors"
	"time"

	"event-booker/internal/config"
	"event-booker/internal/domain"
	"event-booker/internal/repository"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

type BookingUsecase struct {
	db        *dbpg.DB
	repo      bookingRepository
	eventRepo eventRepository
	userRepo  userRepository
	notifier  notifier
	cfg       *config.Config
	logger    *zlog.Zerolog
}

func NewBookingUsecase(db *dbpg.DB, repo bookingRepository, eventRepo eventRepository, userRepo userRepository, notifier notifier, cfg *config.Config, logger *zlog.Zerolog) *BookingUsecase {
	return &BookingUsecase{
		db:        db,
		repo:      repo,
		eventRepo: eventRepo,
		userRepo:  userRepo,
		notifier:  notifier,
		cfg:       cfg,
		logger:    logger,
	}
}

func (uc *BookingUsecase) BookPlace(ctx context.Context, eventID, userID string) (*domain.Booking, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		uc.logger.Error().Err(err).Msg("failed to begin transaction")
		return nil, err
	}
	defer tx.Rollback()

	event, err := uc.eventRepo.GetForUpdate(ctx, tx, eventID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEventNotFound
		}
		uc.logger.Error().Err(err).Str("event_id", eventID).Msg("failed to get event")
		return nil, err
	}
	if event.Available <= 0 {
		return nil, ErrNoSeatsAvailable
	}
	ttl := event.BookingTTL
	if ttl == 0 {
		ttl = uc.cfg.Scheduler.BookingTTL
	}
	now := time.Now()
	booking := &domain.Booking{
		ID:        uuid.NewString(),
		EventID:   eventID,
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now, // default
	}
	if event.RequiresPayment {
		booking.Status = domain.BookingPending
		booking.ExpiresAt = now.Add(ttl)
	} else {
		booking.Status = domain.BookingConfirmed
		booking.ConfirmedAt = &now
	}
	if err := uc.repo.Create(ctx, tx, booking); err != nil {
		uc.logger.Error().Err(err).Str("booking_id", booking.ID).Msg("failed to create booking")
		return nil, err
	}
	if err := uc.eventRepo.DecrementAvailableSeats(ctx, tx, eventID); err != nil {
		uc.logger.Error().Err(err).Str("event_id", eventID).Msg("failed to decrement available seats")
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		uc.logger.Error().Err(err).Msg("failed to commit transaction")
		return nil, err
	}
	return booking, nil
}

func (uc *BookingUsecase) ConfirmBooking(ctx context.Context, bookingID string) error {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		uc.logger.Error().Err(err).Msg("failed to begin transaction")
		return err
	}
	defer tx.Rollback()

	booking, err := uc.repo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrBookingNotFound
		}
		uc.logger.Error().Err(err).Str("booking_id", bookingID).Msg("failed to get booking")
		return err
	}
	if booking.Status != domain.BookingPending {
		return ErrBookingNotPending
	}
	if time.Now().After(booking.ExpiresAt) {
		return ErrBookingExpired
	}
	now := time.Now()
	booking.Status = domain.BookingConfirmed
	booking.ConfirmedAt = &now
	if err := uc.repo.Update(ctx, tx, booking); err != nil {
		uc.logger.Error().Err(err).Str("booking_id", bookingID).Msg("failed to update booking")
		return err
	}
	if err := tx.Commit(); err != nil {
		uc.logger.Error().Err(err).Msg("failed to commit transaction")
		return err
	}
	return nil
}

func (uc *BookingUsecase) CancelBooking(ctx context.Context, bookingID string) error {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		uc.logger.Error().Err(err).Msg("failed to begin transaction")
		return err
	}
	defer tx.Rollback()

	booking, err := uc.repo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrBookingNotFound
		}
		uc.logger.Error().Err(err).Str("booking_id", bookingID).Msg("failed to get booking")
		return err
	}
	if booking.Status == domain.BookingCancelled {
		return ErrAlreadyCancelled
	}
	booking.Status = domain.BookingCancelled
	if err := uc.repo.Update(ctx, tx, booking); err != nil {
		uc.logger.Error().Err(err).Str("booking_id", bookingID).Msg("failed to update booking")
		return err
	}
	if err := uc.eventRepo.IncrementAvailableSeats(ctx, tx, booking.EventID); err != nil {
		uc.logger.Error().Err(err).Str("event_id", booking.EventID).Msg("failed to increment available seats")
		return err
	}
	if err := tx.Commit(); err != nil {
		uc.logger.Error().Err(err).Msg("failed to commit transaction")
		return err
	}
	user, err := uc.userRepo.GetByID(ctx, booking.UserID)
	if err == nil {
		if notifyErr := uc.notifier.NotifyCancellation(user, booking); notifyErr != nil {
			uc.logger.Error().Err(notifyErr).Str("user_id", user.ID).Msg("Failed to notify cancellation")
		}
	} else {
		uc.logger.Error().Err(err).Str("user_id", booking.UserID).Msg("failed to get user for notification")
	}
	return nil
}

func (uc *BookingUsecase) GetExpiredBookings(ctx context.Context) ([]*domain.Booking, error) {
	expired, err := uc.repo.GetExpired(ctx, time.Now())
	if err != nil {
		uc.logger.Error().Err(err).Msg("failed to get expired bookings")
		return nil, err
	}
	return expired, nil
}
