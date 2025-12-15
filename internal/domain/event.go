package domain

import "time"

type Event struct {
	ID              string
	Name            string
	Date            time.Time
	TotalSeats      int
	Available       int
	BookingTTL      time.Duration
	RequiresPayment bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Status          EventStatus
}

type EventStatus string

const (
	EventActive    EventStatus = "active"
	EventCancelled EventStatus = "cancelled"
	EventCompleted EventStatus = "completed"
)
