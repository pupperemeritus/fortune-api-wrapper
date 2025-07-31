package main

import (
	"context"
	"fortune-api/internal/config"
	"fortune-api/internal/handlers"
	"fortune-api/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg := config.Load()

	// Initialize fortune service
	fortuneService := service.NewFortuneService(cfg.FortunePath, logger)

	// Initialize handlers
	handler := handlers.NewHandler(fortuneService, logger)

	// Setup routes
	router := mux.NewRouter()
	router.HandleFunc("/health", handler.HealthCheck).Methods("GET")
	router.HandleFunc("/fortune", handler.GetFortune).Methods("GET")
	router.HandleFunc("/fortune/files", handler.ListFiles).Methods("GET")
	router.HandleFunc("/fortune/search", handler.SearchFortunes).Methods("GET")

	// Add middleware
	router.Use(handlers.LoggingMiddleware(logger))
	router.Use(handlers.CORSMiddleware)

	// Setup server
	srv := &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting Fortune API server", zap.String("address", cfg.ServerAddress))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
