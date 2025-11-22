// Package handler defines the Presentation Layers (HTTP handlers for the API)
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gioannid/platform-go-challenge/internal/domain"
	"github.com/gioannid/platform-go-challenge/internal/service"
)

// Handler holds all HTTP handlers
type Handler struct {
	service *service.FavouriteService
}

// NewHandler creates a new handler instance
func NewHandler(service *service.FavouriteService) *Handler {
	return &Handler{
		service: service,
	}
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

// respondError sends an error response
func respondError(w http.ResponseWriter, statusCode int, err error) {
	respondJSON(w, statusCode, Response{
		Success: false,
		Error:   err.Error(),
	})
}

// respondSuccess sends a success response
func respondSuccess(w http.ResponseWriter, statusCode int, data interface{}, message string) {
	respondJSON(w, statusCode, Response{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// mapDomainError maps domain errors to HTTP status codes
func mapDomainError(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrInvalidAssetType),
		errors.Is(err, domain.ErrMissingAssetData),
		errors.Is(err, domain.ErrInvalidChartData),
		errors.Is(err, domain.ErrInvalidInsightData),
		errors.Is(err, domain.ErrInvalidAudienceData):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
