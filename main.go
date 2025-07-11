package main

import (
	"log"
	"net/http"
	"os"

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

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
