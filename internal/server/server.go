// Package server sets up and runs the HTTP server, applying middleware per the Decorator pattern.
package server

import (
	"context"
	"net/http"

	"github.com/gioannid/platform-go-challenge/internal/config"
	"github.com/gioannid/platform-go-challenge/internal/handler"
	"github.com/gorilla/mux"

	_ "github.com/gioannid/platform-go-challenge/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Server holds HTTP server and dependencies
type Server struct {
	httpServer *http.Server
	config     *config.Config
}

// MiddlewareChain represents a chain of middleware functions
type MiddlewareChain []func(http.Handler) http.Handler

// NewChain creates a new middleware chain
func NewChain(middlewares ...func(http.Handler) http.Handler) MiddlewareChain {
	return middlewares
}

// Append adds middleware to the chain
func (mc MiddlewareChain) Append(middlewares ...func(http.Handler) http.Handler) MiddlewareChain {
	return append(mc, middlewares...)
}

// then: wraps the handler with all middleware
func (mc MiddlewareChain) then(h http.Handler) http.Handler {
	for i := len(mc) - 1; i >= 0; i-- {
		h = mc[i](h)
	}
	return h
}

// New creates a new HTTP server
func New(cfg *config.Config, h *handler.Handler, mw MiddlewareChain) *Server {
	r := mux.NewRouter()

	// Health endpoints (no auth required)
	r.HandleFunc("/healthz", h.HealthCheck).Methods(http.MethodGet)
	r.HandleFunc("/readyz", h.ReadinessCheck).Methods(http.MethodGet)

	// Swagger UI endpoint
	r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods(http.MethodGet)

	// API v1 routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Asset management
	api.HandleFunc("/assets", h.CreateAsset).Methods(http.MethodPost)
	api.HandleFunc("/assets/{assetId}/description", h.UpdateAssetDescription).Methods(http.MethodPatch)
	api.HandleFunc("/assets/{assetId}", h.DeleteAsset).Methods(http.MethodDelete)

	// Favourite management
	api.HandleFunc("/users/{userId}/favourites", h.ListFavourites).Methods(http.MethodGet)
	api.HandleFunc("/users/{userId}/favourites", h.AddFavourite).Methods(http.MethodPost)
	api.HandleFunc("/users/{userId}/favourites/{favouriteId}", h.RemoveFavourite).Methods(http.MethodDelete)

	// Apply middleware chain
	handler := mw.then(r)

	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.ServerAddress,
			Handler:      handler,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		config: cfg,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Router returns the http.Handler for the server. This is primarily for testing.
func (s *Server) Router() http.Handler {
	return s.httpServer.Handler
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
