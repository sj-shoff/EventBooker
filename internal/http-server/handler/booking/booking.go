package booking

import (
	"encoding/json"
	"net/http"

	"event-booker/internal/http-server/handler/booking/dto"

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

func (h *BookingHandler) Book(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.BookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("failed to decode book request")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b, err := h.usecase.BookPlace(r.Context(), id, req.UserID)
	if err != nil {
		h.logger.Error().Err(err).Str("event_id", id).Str("user_id", req.UserID).Msg("failed to book place")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := json.NewEncoder(w).Encode(b); err != nil {
		h.logger.Error().Err(err).Msg("failed to encode booking response")
	}
}

func (h *BookingHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.usecase.ConfirmBooking(r.Context(), id); err != nil {
		h.logger.Error().Err(err).Str("booking_id", id).Msg("failed to confirm booking")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
