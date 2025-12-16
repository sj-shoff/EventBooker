package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"event-booker/internal/http-server/handler/user/dto"
	userErr "event-booker/internal/usecase/user"

	"github.com/go-chi/chi/v5"
	"github.com/wb-go/wbf/zlog"
)

type UserHandler struct {
	usecase userUsecase
	logger  *zlog.Zerolog
}

func NewUserHandler(usecase userUsecase, logger *zlog.Zerolog) *UserHandler {
	return &UserHandler{usecase: usecase, logger: logger}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Msg("User registration request received")
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().
			Err(err).
			Msg("Failed to decode register request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Email == "" {
		h.logger.Error().Msg("Email is required")
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}
	h.logger.Info().
		Str("email", req.Email).
		Str("role", string(req.Role)).
		Msg("Registering new user")
	user, err := h.usecase.RegisterUser(r.Context(), req.Email, req.Telegram, req.Role)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("email", req.Email).
			Msg("Failed to register user")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.logger.Info().
		Str("user_id", user.ID).
		Str("email", user.Email).
		Str("role", string(user.Role)).
		Msg("User registered successfully")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		h.logger.Error().
			Err(err).
			Str("user_id", user.ID).
			Msg("Failed to encode user response")
	}
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		h.logger.Error().Msg("User ID is required")
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}
	h.logger.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("user_id", userID).
		Msg("Get user request")
	user, err := h.usecase.GetUser(r.Context(), userID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("user_id", userID).
			Msg("Failed to get user")
		if errors.Is(err, userErr.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	h.logger.Info().
		Str("user_id", user.ID).
		Str("email", user.Email).
		Msg("User retrieved successfully")
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		h.logger.Error().
			Err(err).
			Str("user_id", user.ID).
			Msg("Failed to encode user response")
	}
}
