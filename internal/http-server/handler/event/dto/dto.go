package dto

import "time"

type CreateEventRequest struct {
	Name            string        `json:"name"`
	Date            time.Time     `json:"date"`
	TotalSeats      int           `json:"total_seats"`
	BookingTTL      time.Duration `json:"booking_ttl"`
	RequiresPayment bool          `json:"requires_payment"`
}
