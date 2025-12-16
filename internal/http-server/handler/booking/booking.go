package booking

import (
	"encoding/json"
	"net/http"

	"event-booker/internal/http-server/handler/booking/dto"
	bookingErr "event-booker/internal/usecase/booking"

	"github.com/go-chi/chi/v5"
	"github.com/wb-go/wbf/zlog"
)

type BookingHandler struct {
	usecase bookingUsecase
	logger  *zlog.Zerolog
}

func NewBookingHandler(usecase bookingUsecase, logger *zlog.Zerolog) *BookingHandler {
	return &BookingHandler{usecase: usecase, logger: logger}
}

func (h *BookingHandler) ListBookings(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Msg("List bookings request received")
	bookings, err := h.usecase.ListBookings(r.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list bookings")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(bookings); err != nil {
		h.logger.Error().Err(err).Msg("Failed to encode bookings")
	}
}

func (h *BookingHandler) Book(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "id")
	h.logger.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("event_id", eventID).
		Msg("Booking request received")
	var req dto.BookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().
			Err(err).
			Str("event_id", eventID).
			Msg("Failed to decode booking request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.UserID == "" {
		h.logger.Error().
			Str("event_id", eventID).
			Msg("User ID is required")
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}
	h.logger.Info().
		Str("event_id", eventID).
		Str("user_id", req.UserID).
		Msg("Processing booking")
	booking, err := h.usecase.BookPlace(r.Context(), eventID, req.UserID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("event_id", eventID).
			Str("user_id", req.UserID).
			Msg("Booking failed")
		switch err {
		case bookingErr.ErrEventNotFound:
			http.Error(w, "Event not found", http.StatusNotFound)
		case bookingErr.ErrNoSeatsAvailable:
			http.Error(w, "No seats available", http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	h.logger.Info().
		Str("booking_id", booking.ID).
		Str("event_id", eventID).
		Str("status", string(booking.Status)).
		Msg("Booking successful")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(booking); err != nil {
		h.logger.Error().
			Err(err).
			Str("booking_id", booking.ID).
			Msg("Failed to encode booking response")
	}
}

func (h *BookingHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	bookingID := chi.URLParam(r, "id")
	h.logger.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("booking_id", bookingID).
		Msg("Confirmation request received")
	if err := h.usecase.ConfirmBooking(r.Context(), bookingID); err != nil {
		h.logger.Error().
			Err(err).
			Str("booking_id", bookingID).
			Msg("Confirmation failed")
		switch err {
		case bookingErr.ErrBookingNotFound:
			http.Error(w, "Booking not found", http.StatusNotFound)
		case bookingErr.ErrBookingNotPending:
			http.Error(w, "Booking is not pending confirmation", http.StatusBadRequest)
		case bookingErr.ErrBookingExpired:
			http.Error(w, "Booking has expired", http.StatusGone)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	h.logger.Info().
		Str("booking_id", bookingID).
		Msg("Booking confirmed successfully")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Booking confirmed successfully",
		"booking_id": bookingID,
	})
}

func (h *BookingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	bookingID := chi.URLParam(r, "id")
	h.logger.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("booking_id", bookingID).
		Msg("Cancel booking request received")
	if err := h.usecase.CancelBooking(r.Context(), bookingID); err != nil {
		h.logger.Error().
			Err(err).
			Str("booking_id", bookingID).
			Msg("Cancellation failed")
		switch err {
		case bookingErr.ErrBookingNotFound:
			http.Error(w, "Booking not found", http.StatusNotFound)
		case bookingErr.ErrAlreadyCancelled:
			http.Error(w, "Booking is already cancelled", http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	h.logger.Info().
		Str("booking_id", bookingID).
		Msg("Booking cancelled successfully")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Booking cancelled successfully",
		"booking_id": bookingID,
	})
}
