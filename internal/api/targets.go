package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/thwqsz/uptime-monitor/internal/service"
)

type TargetCreateRequest struct {
	URL          string `json:"url"`
	Timeout      int    `json:"timeout"`
	IntervalTime int    `json:"interval_time"`
}

func NewTargetHandler(targetService *service.TargetService) *TargetHandler {
	return &TargetHandler{targetService: targetService}
}

type TargetHandler struct {
	targetService *service.TargetService
}

func (h *TargetHandler) TargetCreateHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := service.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req TargetCreateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "decode problem", http.StatusBadRequest)
		return
	}

	ans, err := h.targetService.CreateTarget(r.Context(), userID, req.URL, req.Timeout, req.IntervalTime)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInterval) || errors.Is(err, service.ErrInvalidURL) || errors.Is(err, service.ErrInvalidTimeout) {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(ans); err != nil {
		return
	}
}
