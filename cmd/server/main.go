package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rokkovach/codedb/internal/api"
	"github.com/rokkovach/codedb/internal/db"
)

func main() {
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "postgres://postgres:postgres@localhost:5432/codedb?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	jsonLogging := os.Getenv("LOG_FORMAT") == "json"

	logger := api.SetupLogger(jsonLogging, logLevel)
	log := logger.With().Str("component", "main").Logger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database, err := db.New(ctx, connString)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer database.Close()

	log.Info().Msg("Connected to database")

	apiHandler := api.NewAPI(database)

	go func() {
		if err := apiHandler.SubscriptionSvc().StartListener(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to start subscription listener")
		}
	}()

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      api.LoggingMiddleware(logger)(apiHandler.Router()),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Info().Str("port", port).Msg("Starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server stopped")
}

func init() {
	if _, err := os.Stat("migrations"); os.IsNotExist(err) {
		fmt.Println("Warning: migrations directory not found")
	}
}
