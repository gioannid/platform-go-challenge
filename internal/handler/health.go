package handler

import (
	"net/http"
)

// HealthCheck handles GET /healthz
//
//	@Summary		Health check
//	@Description	Returns the health status of the service
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	HealthResponse
//	@Router			/../../healthz [get]
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondSuccess(w, http.StatusOK, map[string]string{
		"status": "ok",
	}, "")
}

// ReadinessCheck handles GET /readyz
//
//	@Summary		Readiness check
//	@Description	Returns whether the service is ready to accept traffic
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	HealthResponse
//	@Router			/../../readyz [get]
func (h *Handler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.service.HealthCheck(r.Context()); err != nil {
		respondError(w, http.StatusServiceUnavailable, err)
		return
	}

	respondSuccess(w, http.StatusOK, map[string]string{
		"status": "ready",
	}, "")
}
