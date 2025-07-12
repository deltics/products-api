package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"products-api/internal/db"
	"products-api/internal/models"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

const (
	cInvalidJSON      = "Invalid JSON"
	cInvalidProductId = "Invalid product ID"
	cProductNotFound  = "Product not found"
	cValidationFailed = "Validation failed"
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
	const productsRoute = "/products"
	const productByIdRoute = "/products/{id:[0-9]+}"

	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc(productsRoute, h.GetProducts).Methods("GET")
	api.HandleFunc(productsRoute, h.CreateProduct).Methods("POST")
	api.HandleFunc(productsRoute, nil).Methods("OPTIONS") // handled by CORS middleware

	api.HandleFunc(productByIdRoute, h.GetProduct).Methods("GET")
	api.HandleFunc(productByIdRoute, h.UpdateProduct).Methods("PUT")
	api.HandleFunc(productByIdRoute, h.DeleteProduct).Methods("DELETE")
	api.HandleFunc(productByIdRoute, nil).Methods("OPTIONS") // handled by CORS middleware

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

	filters, err := h.productFiltersFromQuery(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid query string", err.Error())
		return
	}

	// Get products from database
	products, total, err := h.db.GetProducts(page, pageSize, filters...)
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
		h.writeErrorResponse(w, http.StatusBadRequest, cInvalidProductId, "")
		return
	}

	product, err := h.db.GetProductByID(id)
	switch {
	case errors.Is(err, db.ErrNotFound):
		h.writeErrorResponse(w, http.StatusNotFound, cProductNotFound, "")
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
		h.writeErrorResponse(w, http.StatusBadRequest, cInvalidJSON, err.Error())
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, cValidationFailed, err.Error())
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
		h.writeErrorResponse(w, http.StatusBadRequest, cInvalidProductId, "")
		return
	}

	var req models.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, cInvalidJSON, err.Error())
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, cValidationFailed, err.Error())
		return
	}

	// Update product
	product, err := h.db.UpdateProduct(id, req)
	switch {
	case errors.Is(err, db.ErrNotFound):
		h.writeErrorResponse(w, http.StatusNotFound, cProductNotFound, "")
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
		h.writeErrorResponse(w, http.StatusBadRequest, cInvalidProductId, "")
		return
	}

	err = h.db.DeleteProduct(id)
	switch {
	case errors.Is(err, db.ErrNotFound):
		h.writeErrorResponse(w, http.StatusNotFound, cProductNotFound, "")
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

func (h *Handler) productFiltersFromQuery(r *http.Request) ([]db.ProductFilter, error) {
	var (
		filters []db.ProductFilter
		errs    []error
	)

	// in stock
	if r.URL.Query().Has("in_stock") {
		inStock := r.URL.Query().Get("in_stock")
		switch strings.ToLower(inStock) {
		case "false":
			filters = append(filters, func(product *models.Product) bool {
				return !product.InStock
			})

		case "true":
			filters = append(filters, func(product *models.Product) bool {
				return product.InStock
			})

		default:
			errs = append(errs, fmt.Errorf("invalid in_stock value: %s", inStock))
		}
	}

	// in a specified category
	if category := r.URL.Query().Get("category"); category != "" {
		filters = append(filters, func(product *models.Product) bool {
			return strings.EqualFold(product.Category, category)
		})
	}

	// name contains a substring
	if name := r.URL.Query().Get("name"); name != "" {
		name = strings.ToLower(name)
		filters = append(filters, func(product *models.Product) bool {
			return strings.Contains(strings.ToLower(product.Name), name)
		})
	}

	// >= minimum price
	if priceMinStr := r.URL.Query().Get("price_min"); priceMinStr != "" {
		priceMin, err := strconv.ParseFloat(priceMinStr, 64)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid price_min: %w", err))
		} else {
			filters = append(filters, func(product *models.Product) bool {
				return product.Price >= priceMin
			})
		}
	}

	// <= maximum price
	if priceMaxStr := r.URL.Query().Get("price_max"); priceMaxStr != "" {
		priceMax, err := strconv.ParseFloat(priceMaxStr, 64)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid price_max: %w", err))
		} else {
			filters = append(filters, func(product *models.Product) bool {
				return product.Price <= priceMax
			})
		}
	}

	return filters, errors.Join(errs...)
}
