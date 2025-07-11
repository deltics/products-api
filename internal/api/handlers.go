package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"products-api/internal/db"
	"products-api/internal/models"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

// Handler handles HTTP requests for the products API
type Handler struct {
	db        db.Database
	validator *validator.Validate
}

// NewHandler creates a new API handler
func NewHandler(database db.Database) *Handler {
	return &Handler{
		db:        database,
		validator: validator.New(),
	}
}

// SetupRoutes configures the HTTP routes
func (h *Handler) SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/products", h.GetProducts).Methods("GET")
	api.HandleFunc("/products", h.CreateProduct).Methods("POST")
	api.HandleFunc("/products", nil).Methods("OPTIONS") // handled by CORS middleware
	api.HandleFunc("/products/{id:[0-9]+}", h.GetProduct).Methods("GET")
	api.HandleFunc("/products/{id:[0-9]+}", h.UpdateProduct).Methods("PUT")
	api.HandleFunc("/products/{id:[0-9]+}", h.DeleteProduct).Methods("DELETE")
	api.HandleFunc("/products/{id:[0-9]+}", nil).Methods("OPTIONS") // handled by CORS middleware

	// Health check endpoint
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")

	// Add middleware
	router.Use(h.loggingMiddleware)
	router.Use(h.corsMiddleware)

	return router
}

// GetProducts handles GET /api/v1/products
func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Get products from database
	products, total, err := h.db.GetProducts(page, pageSize)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve products", err.Error())
		return
	}

	// Calculate total pages
	totalPages := (total + pageSize - 1) / pageSize

	response := models.PaginatedResponse{
		Data:       products,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetProduct handles GET /api/v1/products/{id}
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid product ID", "")
		return
	}

	product, err := h.db.GetProductByID(id)
	switch {
	case errors.Is(err, db.ErrNotFound):
		h.writeErrorResponse(w, http.StatusNotFound, "Product not found", "")
		return

	case err != nil:
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve product", err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, product)
}

// CreateProduct handles POST /api/v1/products
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Create product
	product, err := h.db.CreateProduct(req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create product", err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, product)
}

// UpdateProduct handles PUT /api/v1/products/{id}
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid product ID", "")
		return
	}

	var req models.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Update product
	product, err := h.db.UpdateProduct(id, req)
	switch {
	case errors.Is(err, db.ErrNotFound):
		h.writeErrorResponse(w, http.StatusNotFound, "Product not found", "")
		return

	case err != nil:
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update product", err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, product)
}

// DeleteProduct handles DELETE /api/v1/products/{id}
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid product ID", "")
		return
	}

	err = h.db.DeleteProduct(id)
	switch {
	case errors.Is(err, db.ErrNotFound):
		h.writeErrorResponse(w, http.StatusNotFound, "Product not found", "")
		return

	case err != nil:
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete product", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HealthCheck handles GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "healthy",
		"service": "products-api",
	}
	h.writeJSONResponse(w, http.StatusOK, response)
}

// Helper methods

func (h *Handler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeErrorResponse(w http.ResponseWriter, status int, message, details string) {
	response := models.ErrorResponse{
		Error:   message,
		Message: details,
	}
	h.writeJSONResponse(w, status, response)
}

// Middleware

func (h *Handler) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s %s\n", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
