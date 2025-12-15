package scheduler

import (
	"context"

	"event-booker/internal/domain"
)

type bookingUsecase interface {
	GetExpiredBookings(ctx context.Context) ([]*domain.Booking, error)
	CancelBooking(ctx context.Context, bookingID string) error
}
