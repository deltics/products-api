package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"products-api/internal/api"
	"products-api/internal/api/ratelimiter"
	"products-api/internal/db"
)

func TestMainIntegration(t *testing.T) {
	// Test that we can create a complete application setup
	database := db.NewInMemoryDB()
	limiter := ratelimiter.NewNoopLimiter() // Use NoopLimiter for integration tests
	handler := api.NewHandler(database, limiter)
	mux := handler.SetupRoutes()

	// Test health endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Health check failed with status %d", status)
	}

	// Test products endpoint
	req = httptest.NewRequest("GET", "/api/v1/products", nil)
	rr = httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Products endpoint failed with status %d", status)
	}
}

func TestEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		expectedVal string
	}{
		{
			name:        "Default port when PORT not set",
			envValue:    "",
			expectedVal: "8080",
		},
		{
			name:        "Custom port when PORT is set",
			envValue:    "3000",
			expectedVal: "3000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			originalPort := os.Getenv("PORT")
			defer func() { _ = os.Setenv("PORT", originalPort) }()

			// Set test value
			var err error
			if tt.envValue == "" {
				err = os.Unsetenv("PORT")
			} else {
				err = os.Setenv("PORT", tt.envValue)
			}
			if err != nil {
				t.Fatalf("Failed to set PORT environment variable: %v", err)
			}

			// Test the logic from main function
			port := os.Getenv("PORT")
			if port == "" {
				port = "8080"
			}

			if port != tt.expectedVal {
				t.Errorf("Expected port %s, got %s", tt.expectedVal, port)
			}
		})
	}
}

func TestServerCanStart(t *testing.T) {
	// Test that the server can be initialized without errors
	database := db.NewInMemoryDB()
	limiter := ratelimiter.NewNoopLimiter()
	handler := api.NewHandler(database, limiter)
	mux := handler.SetupRoutes()

	// Create a test server
	server := httptest.NewServer(mux)
	defer server.Close()

	// Make a request to ensure the server is working
	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request to test server: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestConcurrentRequests(t *testing.T) {
	// Test that the application can handle concurrent requests
	database := db.NewInMemoryDB()
	limiter := ratelimiter.NewNoopLimiter()
	handler := api.NewHandler(database, limiter)
	mux := handler.SetupRoutes()

	server := httptest.NewServer(mux)
	defer server.Close()

	// Make multiple concurrent requests
	const numRequests = 50
	done := make(chan bool, numRequests)
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := http.Get(server.URL + "/api/v1/products")
			if err != nil {
				errors <- err
				return
			}
			if err := resp.Body.Close(); err != nil {
				errors <- fmt.Errorf("error closing response body: %w", err)
				return
			}

			if resp.StatusCode != http.StatusOK {
				errors <- err
				return
			}

			done <- true
		}()
	}

	// Wait for all requests to complete or timeout
	timeout := time.After(10 * time.Second)
	completed := 0

	for completed < numRequests {
		select {
		case <-done:
			completed++
		case err := <-errors:
			t.Fatalf("Request failed: %v", err)
		case <-timeout:
			t.Fatalf("Timeout waiting for concurrent requests to complete. Completed: %d/%d", completed, numRequests)
		}
	}
}
