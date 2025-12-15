package booking

import (
	"context"

	"event-booker/internal/domain"
)

type bookingUsecase interface {
	BookPlace(ctx context.Context, eventID, userID string) (*domain.Booking, error)
	ConfirmBooking(ctx context.Context, bookingID string) error
}
