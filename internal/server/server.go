// Package server sets up and runs the HTTP server, applying middleware per the Decorator pattern.
package server

import (
	"context"
	"net/http"

	"github.com/gioannid/platform-go-challenge/internal/config"
	"github.com/gioannid/platform-go-challenge/internal/handler"
	"github.com/gioannid/platform-go-challenge/internal/middleware"
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

func New(cfg *config.Config, h *handler.Handler, generalMW MiddlewareChain) *Server {
	r := mux.NewRouter()

	// Apply general middleware to the main router.
	// These middlewares will be applied to all routes (health, swagger, and API v1).
	for _, m := range generalMW {
		r.Use(m)
	}

	// Health endpoints (no auth required)
	r.HandleFunc("/healthz", h.HealthCheck).Methods(http.MethodGet)
	r.HandleFunc("/readyz", h.ReadinessCheck).Methods(http.MethodGet)

	// Swagger UI endpoint (no auth required)
	r.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods(http.MethodGet)

	// API v1 routes subrouter
	api := r.PathPrefix("/api/v1").Subrouter()

	// Conditionally apply JWT authentication middleware *only* to the API v1 subrouter
	if cfg.AuthEnabled {
		api.Use(middleware.JWTAuth(cfg.JWTSecret))
	}

	// Asset management (these handlers will be protected if auth is enabled)
	api.HandleFunc("/assets", h.CreateAsset).Methods(http.MethodPost)
	api.HandleFunc("/assets", h.ListAssets).Methods(http.MethodGet)
	api.HandleFunc("/assets/{assetId}/description", h.UpdateAssetDescription).Methods(http.MethodPatch)
	api.HandleFunc("/assets/{assetId}", h.DeleteAsset).Methods(http.MethodDelete)

	// Favourite management (these handlers will be protected if auth is enabled)
	// TODO: Update routes to remove {userId} from path parameters.
	// The userId will now be extracted from the JWT token in the middleware and
	// passed via the request context to the handlers.
	api.HandleFunc("/users/{userId}/favourites", h.ListFavourites).Methods(http.MethodGet)
	api.HandleFunc("/users/{userId}/favourites", h.AddFavourite).Methods(http.MethodPost)
	api.HandleFunc("/users/{userId}/favourites/{favouriteId}", h.RemoveFavourite).Methods(http.MethodDelete)

	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.ServerAddress,
			Handler:      r, // The main router 'r' is now the handler, with middleware applied via .Use()
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
