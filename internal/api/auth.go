package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/thwqsz/uptime-monitor/internal/service"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authS *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authS}
}

func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	input := json.NewDecoder(r.Body)
	err := input.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("неправильный запрос"))
		return
	}

	err = h.authService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("пользователь уже существует"))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("ошибка сервера"))
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("успешно зарегестрирован"))
	return

}
