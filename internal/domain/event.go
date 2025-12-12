package domain

import "time"

type Event struct {
	ID              string
	Name            string
	Date            time.Time
	TotalSeats      int
	Available       int
	BookingTTL      time.Duration
	CreatedAt       time.Time
	UpdatedAt       time.Time
	RequiresPayment bool
}

type EventStatus string

const (
	EventActive EventStatus = "active"
	EventClosed EventStatus = "closed"
)
