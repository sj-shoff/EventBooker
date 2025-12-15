// internal/http-server/handler/event/event.go
package event

import (
	"encoding/json"
	"net/http"
	"time"

	"event-booker/internal/http-server/handler/event/dto"

	"github.com/go-chi/chi/v5"
	"github.com/wb-go/wbf/zlog"
)

type EventHandler struct {
	usecase eventUsecase
	logger  *zlog.Zerolog
}

func NewEventHandler(usecase eventUsecase, logger *zlog.Zerolog) *EventHandler {
	return &EventHandler{usecase: usecase, logger: logger}
}

func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("failed to decode create event request")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ttl, err := time.ParseDuration(string(rune(req.BookingTTL)))
	if err != nil {
		h.logger.Error().Err(err).Msg("invalid booking_ttl")
		http.Error(w, "invalid booking_ttl", http.StatusBadRequest)
		return
	}
	e, err := h.usecase.CreateEvent(r.Context(), req.Name, req.Date, req.TotalSeats, ttl, req.RequiresPayment)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create event")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(e); err != nil {
		h.logger.Error().Err(err).Msg("failed to encode event response")
	}
}

func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	e, err := h.usecase.GetEvent(r.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Str("event_id", id).Msg("failed to get event")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err := json.NewEncoder(w).Encode(e); err != nil {
		h.logger.Error().Err(err).Msg("failed to encode event response")
	}
}

func (h *EventHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.usecase.ListEvents(r.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list events")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(events); err != nil {
		h.logger.Error().Err(err).Msg("failed to encode events response")
	}
}
