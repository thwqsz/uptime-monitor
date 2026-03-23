package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/thwqsz/uptime-monitor/internal/service"
)

type CheckHandler struct {
	checkService *service.CheckService
}

func NewCheckHandler(checkServ *service.CheckService) *CheckHandler {
	return &CheckHandler{checkService: checkServ}
}

func (h *CheckHandler) CheckHandler(w http.ResponseWriter, r *http.Request) {
	targetIDString := chi.URLParam(r, "id")
	targetID, err := strconv.ParseInt(targetIDString, 10, 64)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	logTarget, err := h.checkService.CheckTarget(r.Context(), targetID)
	// тут могут быть разные ошибки, надо позже расписать
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(logTarget); err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

}
