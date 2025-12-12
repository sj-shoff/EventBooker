// internal/http-server/handler/event/event.go
package event

import (
	"encoding/json"
	"net/http"

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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	e, err := h.usecase.CreateEvent(r.Context(), req.Name, req.Date, req.TotalSeats, req.BookingTTL, req.RequiresPayment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(e)
}

func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	e, err := h.usecase.GetEvent(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(e)
}
