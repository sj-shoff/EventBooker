package event_uc

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"event-booker/internal/domain"
	"event-booker/internal/repository"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

type EventUsecase struct {
	db          *dbpg.DB
	repo        eventRepository
	bookingRepo bookingRepository
	userRepo    userRepository
	notifier    notifier
	logger      *zlog.Zerolog
}

func NewEventUsecase(db *dbpg.DB, repo eventRepository, bookingRepo bookingRepository, userRepo userRepository, notifier notifier, logger *zlog.Zerolog) *EventUsecase {
	return &EventUsecase{
		db:          db,
		repo:        repo,
		bookingRepo: bookingRepo,
		userRepo:    userRepo,
		notifier:    notifier,
		logger:      logger,
	}
}

func (uc *EventUsecase) CancelEvent(ctx context.Context, eventID string, reason string) error {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		uc.logger.Error().Err(err).Str("event_id", eventID).Msg("Failed to begin transaction")
		return err
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				uc.logger.Error().Err(rbErr).Str("event_id", eventID).Msg("Failed to rollback transaction")
			}
		}
	}()
	event, err := uc.repo.GetForUpdate(ctx, tx, eventID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrEventNotFound
		}
		uc.logger.Error().Err(err).Str("event_id", eventID).Msg("Failed to get event for update")
		return err
	}
	if err := uc.validateEventCancellation(event); err != nil {
		return err
	}
	bookings, err := uc.bookingRepo.GetByEventID(ctx, eventID)
	if err != nil {
		uc.logger.Error().Err(err).Str("event_id", eventID).Msg("Failed to get bookings for event")
		return err
	}
	var notifications []notificationData
	for _, booking := range bookings {
		if booking.Status != domain.BookingCancelled {
			notifications = append(notifications, notificationData{
				bookingID: booking.ID,
				userID:    booking.UserID,
			})
		}
	}
	cancelledCount := 0
	for _, booking := range bookings {
		if booking.Status != domain.BookingCancelled {
			if err := uc.cancelBookingInTx(ctx, tx, booking); err != nil {
				uc.logger.Error().Err(err).
					Str("booking_id", booking.ID).
					Str("event_id", eventID).
					Msg("Failed to cancel booking in transaction")
				return err
			}
			cancelledCount++
		}
	}
	event.Status = domain.EventCancelled
	event.UpdatedAt = time.Now()
	if err := uc.repo.Update(ctx, tx, event); err != nil {
		uc.logger.Error().Err(err).Str("event_id", eventID).Msg("Failed to update event status")
		return err
	}
	if err := tx.Commit(); err != nil {
		uc.logger.Error().Err(err).Str("event_id", eventID).Msg("Failed to commit transaction")
		return err
	}
	uc.sendNotificationsAsync(notifications, event)
	uc.logger.Info().
		Str("event_id", eventID).
		Str("event_name", event.Name).
		Int("total_bookings", len(bookings)).
		Int("cancelled_bookings", cancelledCount).
		Str("reason", reason).
		Msg("Event cancelled successfully")
	return nil
}

func (uc *EventUsecase) validateEventCancellation(event *domain.Event) error {
	if event.Status == domain.EventCancelled {
		return ErrEventAlreadyCancelled
	}
	if event.Date.Before(time.Now()) {
		return ErrCannotCancelPastEvent
	}
	minCancellationTime := 24 * time.Hour
	if time.Until(event.Date) < minCancellationTime {
		return ErrCancellationTooLate
	}
	return nil
}

func (uc *EventUsecase) cancelBookingInTx(ctx context.Context, tx *sql.Tx, booking *domain.Booking) error {
	oldStatus := booking.Status
	booking.Status = domain.BookingCancelled
	if err := uc.bookingRepo.Update(ctx, tx, booking); err != nil {
		return err
	}
	if err := uc.repo.IncrementAvailableSeats(ctx, tx, booking.EventID); err != nil {
		return err
	}
	uc.logger.Debug().
		Str("booking_id", booking.ID).
		Str("old_status", string(oldStatus)).
		Str("new_status", string(booking.Status)).
		Str("user_id", booking.UserID).
		Msg("Booking cancelled in transaction")
	return nil
}

type notificationData struct {
	bookingID string
	userID    string
}

func (uc *EventUsecase) sendNotificationsAsync(notifications []notificationData, event *domain.Event) {
	if len(notifications) == 0 {
		return
	}
	go func() {
		notifyCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		uc.logger.Info().
			Int("notification_count", len(notifications)).
			Str("event_id", event.ID).
			Msg("Starting async notification sending")
		sentCount := 0
		failedCount := 0
		for _, data := range notifications {
			select {
			case <-notifyCtx.Done():
				uc.logger.Warn().
					Str("event_id", event.ID).
					Msg("Notification sending cancelled due to timeout")
				return
			default:
				if err := uc.sendSingleNotification(notifyCtx, data.userID, data.bookingID, event); err != nil {
					failedCount++
					uc.logger.Error().
						Err(err).
						Str("user_id", data.userID).
						Str("booking_id", data.bookingID).
						Msg("Failed to send notification")
				} else {
					sentCount++
				}
			}
		}
		uc.logger.Info().
			Str("event_id", event.ID).
			Int("sent", sentCount).
			Int("failed", failedCount).
			Int("total", len(notifications)).
			Msg("Notification sending completed")
	}()
}

func (uc *EventUsecase) sendSingleNotification(ctx context.Context, userID, bookingID string, event *domain.Event) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	notificationBooking := &domain.Booking{
		ID:      bookingID,
		EventID: event.ID,
		UserID:  userID,
		Status:  domain.BookingCancelled,
	}
	if err := uc.notifier.NotifyCancellation(user, notificationBooking); err != nil {
		return err
	}
	return nil
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
		Status:          domain.EventActive,
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
