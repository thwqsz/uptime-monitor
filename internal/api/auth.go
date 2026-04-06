package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/thwqsz/uptime-monitor/internal/service"
	"go.uber.org/zap"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type authService interface {
	Register(ctx context.Context, email string, password string) error
	Login(ctx context.Context, email string, password string) (string, error)
}

type AuthHandler struct {
	authService authService
	log         *zap.Logger
}

func NewAuthHandler(authS authService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{authService: authS, log: log}
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
		h.log.Error("internal error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("ошибка сервера"))
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("успешно зарегестрирован"))
	return

}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	var err error
	s := json.NewDecoder(r.Body)
	err = s.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("неправильный запрос"))
		return
	}

	token, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("неправильные данные для входа"))
			return
		}
		h.log.Error("internal error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("ошибка сервера"))
		return
	}
	resp := AuthResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&resp)
	if err != nil {
		h.log.Error("error during Encode", zap.Error(err))
		return
	}
	return
}
