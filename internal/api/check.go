package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/thwqsz/uptime-monitor/internal/auth"
	"github.com/thwqsz/uptime-monitor/internal/models"
	"github.com/thwqsz/uptime-monitor/internal/service"
)

type CheckHandler struct {
	checkService checkService
}

type checkService interface {
	CheckTargetForUser(ctx context.Context, targetID, userID int64) (*models.CheckLog, error)
}

func NewCheckHandler(checkServ checkService) *CheckHandler {
	return &CheckHandler{checkService: checkServ}
}

func (h *CheckHandler) CheckHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	targetIDString := chi.URLParam(r, "id")
	targetID, err := strconv.ParseInt(targetIDString, 10, 64)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	logTarget, err := h.checkService.CheckTargetForUser(r.Context(), targetID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNoTargetFound):
			http.Error(w, "no target found", http.StatusNotFound)
			return
		case errors.Is(err, service.ErrAccessDenied):
			http.Error(w, "error access denied", http.StatusForbidden)
			return
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(logTarget); err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

}
