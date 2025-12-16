package event

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"event-booker/internal/http-server/handler/event/dto"
	eventErr "event-booker/internal/usecase/event"

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
	h.logger.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Msg("Create event request received")
	var req dto.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().
			Err(err).
			Msg("Failed to decode create event request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	h.logger.Debug().
		Str("name", req.Name).
		Str("date", req.Date).
		Int("seats", req.TotalSeats).
		Str("ttl", req.BookingTTL).
		Bool("requires_payment", req.RequiresPayment).
		Msg("Parsed create event request")
	eventDate, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("date_string", req.Date).
			Msg("Failed to parse event date")
		http.Error(w, "Invalid date format. Use RFC3339 format (e.g., 2024-01-01T18:00:00Z)", http.StatusBadRequest)
		return
	}
	bookingTTL, err := time.ParseDuration(req.BookingTTL)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("ttl_string", req.BookingTTL).
			Msg("Failed to parse booking TTL")
		http.Error(w, "Invalid booking_ttl format. Use Go duration format (e.g., 30m, 2h, 24h)", http.StatusBadRequest)
		return
	}
	if bookingTTL <= 0 {
		h.logger.Error().
			Str("ttl", req.BookingTTL).
			Msg("Booking TTL must be positive")
		http.Error(w, "Booking TTL must be positive duration", http.StatusBadRequest)
		return
	}
	if eventDate.Before(time.Now()) {
		h.logger.Error().
			Time("event_date", eventDate).
			Msg("Event date is in the past")
		http.Error(w, "Event date must be in the future", http.StatusBadRequest)
		return
	}
	if req.TotalSeats <= 0 {
		h.logger.Error().
			Int("total_seats", req.TotalSeats).
			Msg("Total seats must be positive")
		http.Error(w, "Total seats must be positive", http.StatusBadRequest)
		return
	}
	h.logger.Info().
		Str("name", req.Name).
		Time("date", eventDate).
		Int("seats", req.TotalSeats).
		Dur("ttl", bookingTTL).
		Bool("requires_payment", req.RequiresPayment).
		Msg("Creating new event")
	event, err := h.usecase.CreateEvent(r.Context(), req.Name, eventDate, req.TotalSeats, bookingTTL, req.RequiresPayment)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("name", req.Name).
			Msg("Failed to create event")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.logger.Info().
		Str("event_id", event.ID).
		Str("name", event.Name).
		Int("available_seats", event.Available).
		Msg("Event created successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(event); err != nil {
		h.logger.Error().
			Err(err).
			Str("event_id", event.ID).
			Msg("Failed to encode event response")
	}
}

func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "id")
	h.logger.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("event_id", eventID).
		Msg("Get event request")
	event, err := h.usecase.GetEvent(r.Context(), eventID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("event_id", eventID).
			Msg("Failed to get event")
		if errors.Is(err, eventErr.ErrEventNotFound) {
			http.Error(w, "Event not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	h.logger.Info().
		Str("event_id", event.ID).
		Str("name", event.Name).
		Int("available_seats", event.Available).
		Msg("Event retrieved successfully")
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(event); err != nil {
		h.logger.Error().
			Err(err).
			Str("event_id", event.ID).
			Msg("Failed to encode event response")
	}
}

func (h *EventHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Msg("List events request received")
	events, err := h.usecase.ListEvents(r.Context())
	if err != nil {
		h.logger.Error().
			Err(err).
			Msg("Failed to list events")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.logger.Info().
		Int("count", len(events)).
		Msg("Events listed successfully")
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(events); err != nil {
		h.logger.Error().
			Err(err).
			Msg("Failed to encode events response")
	}
}

func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "id")
	h.logger.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("event_id", eventID).
		Msg("Delete event request received")
	var req struct {
		Reason string `json:"reason"`
	}
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.logger.Warn().
				Err(err).
				Str("event_id", eventID).
				Msg("Failed to decode cancellation reason, proceeding without reason")
		}
	}
	h.logger.Info().
		Str("event_id", eventID).
		Str("reason", req.Reason).
		Msg("Processing event cancellation")
	if err := h.usecase.CancelEvent(r.Context(), eventID, req.Reason); err != nil {
		h.logger.Error().
			Err(err).
			Str("event_id", eventID).
			Msg("Event cancellation failed")
		if errors.Is(err, eventErr.ErrEventNotFound) {
			http.Error(w, "Event not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	h.logger.Info().
		Str("event_id", eventID).
		Msg("Event cancelled successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Event cancelled successfully",
		"event_id": eventID,
		"reason":   req.Reason,
	})
}
