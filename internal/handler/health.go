package handler

import (
	"net/http"
)

// HealthCheck handles GET /healthz (liveness probe)
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondSuccess(w, http.StatusOK, map[string]string{
		"status": "ok",
	}, "")
}

// ReadinessCheck handles GET /readyz (readiness probe)
func (h *Handler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.service.HealthCheck(r.Context()); err != nil {
		respondError(w, http.StatusServiceUnavailable, err)
		return
	}

	respondSuccess(w, http.StatusOK, map[string]string{
		"status": "ready",
	}, "")
}
