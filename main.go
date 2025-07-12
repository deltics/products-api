package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"products-api/internal/api"
	"products-api/internal/db"
)

func main() {
	// Initialize the in-memory database
	database := db.NewInMemoryDB()

	// Create the API handler with the database
	handler := api.NewHandler(database)

	// Set up routes
	mux := handler.SetupRoutes()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Start the HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	go func() {
		log.Printf("Server starting on port %s", port)
		err := srv.ListenAndServe()
		switch {
		case errors.Is(err, http.ErrServerClosed):
			log.Println("Server exited gracefully")
		case err != nil:
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	<-stop // wait for stop signal

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed: %+v", err)
	}
}
