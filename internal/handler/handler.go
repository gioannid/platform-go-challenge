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

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"error message"`
}

// BadRequestError represents a 400 error
type BadRequestError struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"unexpected end of JSON input"`
}

// InvalidUUIDError represents invalid UUID format error
type InvalidUUIDError struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"invalid UUID length: 10"`
}

// NotFoundError represents a 404 error
type NotFoundError struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"resource not found"`
}

// ConflictError represents a 409 error
type ConflictError struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"resource already exists"`
}

// InternalServerError represents a 500 error
type InternalServerError struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"internal server error"`
}

// SuccessResponse represents a success response with message
type SuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"operation completed successfully"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Success bool                   `json:"success" example:"true"`
	Data    map[string]interface{} `json:"data" swaggertype:"object,string" example:"{\"status\":\"ok\"}"`
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
