package booking_uc

import "errors"

var (
	ErrEventNotFound     = errors.New("event not found")
	ErrNoSeatsAvailable  = errors.New("no seats available")
	ErrBookingNotFound   = errors.New("booking not found")
	ErrBookingNotPending = errors.New("booking not pending")
	ErrBookingExpired    = errors.New("booking expired")
	ErrAlreadyCancelled  = errors.New("booking already cancelled")
)
