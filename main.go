package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"products-api/internal/api"
	"products-api/internal/db"
)

func main() {
	var err error

	// Create a context for the application
	ctx := context.Background()

	// Initialize the in-memory database
	database := db.NewInMemoryDB()

	// Initialize a rate limiter with a cancellable context
	ctx, cancelRateLimiter := context.WithCancel(ctx)
	defer cancelRateLimiter()

	var rateLimit = 100
	if s := os.Getenv("RATE_LIMIT"); s != "" {
		var err error
		rateLimit, err = strconv.Atoi(s)
		if err != nil {
			log.Fatalf("Invalid RATE_LIMIT: %v", err)
		}
	}
	log.Println("RATE_LIMIT:", rateLimit, "requests/sec")

	rateLimiter, err := api.NewRateLimiter(ctx, rateLimit)
	if err != nil {
		log.Fatalf("Failed to create rate limiter: %v", err)
	}

	// Create the API handler with the database
	handler := api.NewHandler(database, rateLimiter)

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

	ctx, cancelShutdown := context.WithTimeout(ctx, 15*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed: %+v", err)
	}
}
