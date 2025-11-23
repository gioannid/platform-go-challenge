// Package main implements the entry point for the GlobalWebIndex Engineering Challenge application.
// It sets up configuration, initializes dependencies, and starts the HTTP server by default listening to port 8080
// with graceful shutdown.
//
//	@title						GWI Platform Favourites API
//	@version					1.0
//	@description				REST API for managing user favourites with clean architecture
//	@description				Allows users to favourite assets (charts, insights, audiences)
//	@termsOfService				http://swagger.io/terms/
//
//	@contact.name				API Support
//	@contact.url				https://github.com/gioannid/platform-go-challenge
//	@contact.email				support@example.com
//
//	@license.name				MIT
//	@license.url				https://opensource.org/licenses/MIT
//
//	@host						localhost:8080
//	@BasePath					/api/v1
//
//	@schemes					http https
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gioannid/platform-go-challenge/internal/config"
	"github.com/gioannid/platform-go-challenge/internal/handler"
	"github.com/gioannid/platform-go-challenge/internal/middleware"
	"github.com/gioannid/platform-go-challenge/internal/repository/memory"
	"github.com/gioannid/platform-go-challenge/internal/server"
	"github.com/gioannid/platform-go-challenge/internal/service"
)

func main() {
	// Load configuration from environment
	cfg := config.Get()

	// Initialize repository (in-memory by default, swap with postgres if needed)
	repo := memory.NewRepository()

	// Initialize service layer
	svc := service.NewFavouriteService(repo)

	// Initialize HTTP handlers
	h := handler.NewHandler(svc)

	// Setup middleware chain
	mw := server.NewChain(
		middleware.Logger(), // TODO: add other middleware as needed
		//        middleware.Recovery(),
		//        middleware.CORS(cfg.AllowedOrigins),
		//        middleware.RateLimit(cfg.RateLimitRequests, cfg.RateLimitWindow),
	)

	// Add optional authentication if enabled
	if cfg.AuthEnabled {
		//		mw = mw.Append(middleware.JWTAuth(cfg.JWTSecret))	// TODO: implement JWTAuth middleware
	}

	// Create and configure HTTP server
	srv := server.New(cfg, h, mw)

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s", cfg.ServerAddress)
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
