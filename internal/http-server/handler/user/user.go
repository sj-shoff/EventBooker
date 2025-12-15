package user

import (
	"encoding/json"
	"net/http"

	"event-booker/internal/http-server/handler/user/dto"

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
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error().Err(err).Msg("failed to decode register request")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	u, err := h.usecase.RegisterUser(r.Context(), req.Email, req.Telegram, req.Role)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to register user")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(u); err != nil {
		h.logger.Error().Err(err).Msg("failed to encode user response")
	}
}
