package event_uc

import "errors"

var (
	ErrEventNotFound         = errors.New("event not found")
	ErrEventAlreadyCancelled = errors.New("event is already cancelled")
	ErrCannotCancelPastEvent = errors.New("cannot cancel past event")
	ErrCancellationTooLate   = errors.New("cannot cancel event less than 24 hours before start")
	ErrEventAlreadyStarted   = errors.New("event has already started")
	ErrInvalidEventStatus    = errors.New("invalid event status")
)
