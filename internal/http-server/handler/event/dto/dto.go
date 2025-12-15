package dto

type CreateEventRequest struct {
	Name            string `json:"name"`
	Date            string `json:"date"`
	TotalSeats      int    `json:"total_seats"`
	BookingTTL      string `json:"booking_ttl"`
	RequiresPayment bool   `json:"requires_payment"`
}
