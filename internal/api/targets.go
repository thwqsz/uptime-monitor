package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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

//type ITargetService interface {
//}

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
		http.Error(w, "invalid request", http.StatusBadRequest)
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

func (h *TargetHandler) TargetListHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := service.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	ans, err := h.targetService.ListTargets(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ans); err != nil {

		return
	}
}

func (h *TargetHandler) DeleteTargetHandler(w http.ResponseWriter, r *http.Request) {
	targetIDString := chi.URLParam(r, "id")
	targetID, err := strconv.ParseInt(targetIDString, 10, 64)
	if err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}
	userID, ok := service.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	err = h.targetService.DeleteTarget(r.Context(), targetID, userID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidUserID) || errors.Is(err, service.ErrInvalidTargetID) {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrNoTargetFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
