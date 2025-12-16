package router

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"event-booker/internal/http-server/handler/booking"
	"event-booker/internal/http-server/handler/event"
	"event-booker/internal/http-server/handler/user"
	"event-booker/internal/http-server/middleware"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	EventHandler   *event.EventHandler
	BookingHandler *booking.BookingHandler
	UserHandler    *user.UserHandler
}

func SetupRouter(h *Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware)
	r.Route("/api", func(r chi.Router) {
		r.Route("/events", func(r chi.Router) {
			r.Get("/", h.EventHandler.ListEvents)
			r.Post("/", h.EventHandler.CreateEvent)
			r.Get("/{id}", h.EventHandler.GetEvent)
			r.Delete("/{id}", h.EventHandler.DeleteEvent)
			r.Post("/{id}/book", h.BookingHandler.Book)
			r.Post("/{id}/confirm", func(w http.ResponseWriter, r *http.Request) {
				var req struct {
					BookingID string `json:"booking_id"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, "Invalid request body", http.StatusBadRequest)
					return
				}
				if req.BookingID == "" {
					http.Error(w, "booking_id required", http.StatusBadRequest)
					return
				}
				h.BookingHandler.Confirm(w, r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, map[string]string{"id": req.BookingID})))
			})
		})
		r.Route("/bookings", func(r chi.Router) {
			r.Get("/", h.BookingHandler.ListBookings)
			r.Post("/{id}/confirm", h.BookingHandler.Confirm)
			r.Delete("/{id}", h.BookingHandler.Cancel)
		})
		r.Route("/users", func(r chi.Router) {
			r.Post("/", h.UserHandler.Register)
			r.Get("/{id}", h.UserHandler.GetUser)
		})
	})
	workDir, _ := os.Getwd()
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(workDir, "static")))))
	r.Get("/", serveIndex)
	r.Get("/admin", serveAdmin)
	return r
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

func serveAdmin(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/admin.html")
}
