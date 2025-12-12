package domain

import "time"

type Booking struct {
	ID          string
	EventID     string
	UserID      string
	Status      BookingStatus
	CreatedAt   time.Time
	ExpiresAt   time.Time
	ConfirmedAt *time.Time
}

type BookingStatus string

const (
	BookingPending   BookingStatus = "pending"
	BookingConfirmed BookingStatus = "confirmed"
	BookingCancelled BookingStatus = "cancelled"
)
