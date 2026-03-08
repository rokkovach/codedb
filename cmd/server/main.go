package main

import (
	"context"
	"fmt"
	"log"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database, err := db.New(ctx, connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	log.Println("Connected to database")

	apiHandler := api.NewAPI(database)

	go func() {
		if err := apiHandler.SubscriptionSvc().StartListener(ctx); err != nil {
			log.Printf("Failed to start subscription listener: %v", err)
		}
	}()

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      apiHandler.Router(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Starting server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

func init() {
	if _, err := os.Stat("migrations"); os.IsNotExist(err) {
		fmt.Println("Warning: migrations directory not found")
	}
}
