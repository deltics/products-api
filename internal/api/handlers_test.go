package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"products-api/internal/db"
	"products-api/internal/models"

	"github.com/gorilla/mux"
)

// mockDB implements the Database interface for testing
type mockDB struct {
	products   map[int]*models.Product
	nextID     int
	shouldFail bool
}

func newMockDB() *mockDB {
	return &mockDB{
		products: make(map[int]*models.Product),
		nextID:   1,
	}
}

func (m *mockDB) GetProducts(page, pageSize int, filters ...db.ProductFilter) ([]models.Product, int, error) {
	if m.shouldFail {
		return nil, 0, fmt.Errorf("mock database error")
	}

	products := make([]models.Product, 0, len(m.products))
productLoop:
	for _, p := range m.products {
		if len(filters) > 0 {
			for _, filter := range filters {
				if !filter(p) {
					continue productLoop
				}
			}
		}

		products = append(products, *p)
	}

	total := len(products)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		return []models.Product{}, total, nil
	}
	if end > total {
		end = total
	}

	return products[start:end], total, nil
}

func (m *mockDB) GetProductByID(id int) (*models.Product, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock database error")
	}

	product, exists := m.products[id]
	if !exists {
		return nil, db.ErrNotFound
	}

	productCopy := *product
	return &productCopy, nil
}

func (m *mockDB) CreateProduct(req models.CreateProductRequest) (*models.Product, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock database error")
	}

	product := &models.Product{
		ID:          m.nextID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		InStock:     req.InStock,
	}

	m.products[m.nextID] = product
	m.nextID++

	productCopy := *product
	return &productCopy, nil
}

func (m *mockDB) UpdateProduct(id int, req models.UpdateProductRequest) (*models.Product, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock database error")
	}

	product, exists := m.products[id]
	if !exists {
		return nil, db.ErrNotFound
	}

	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Category != nil {
		product.Category = *req.Category
	}
	if req.InStock != nil {
		product.InStock = *req.InStock
	}

	productCopy := *product
	return &productCopy, nil
}

func (m *mockDB) DeleteProduct(id int) error {
	if m.shouldFail {
		return fmt.Errorf("mock database error")
	}

	_, exists := m.products[id]
	if !exists {
		return db.ErrNotFound
	}

	delete(m.products, id)
	return nil
}

func TestHealthCheck(t *testing.T) {
	mockDB := newMockDB()
	handler := NewHandler(mockDB)

	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	handler.HealthCheck(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	expected := `{"service":"products-api","status":"healthy"}`
	body := strings.TrimSpace(rr.Body.String())
	if body != expected {
		t.Errorf("Expected body %s, got %s", expected, body)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestGetProducts(t *testing.T) {
	mockDB := newMockDB()
	handler := NewHandler(mockDB)

	// Add some test products
	testProducts := []models.CreateProductRequest{
		{Name: "Product 1", Price: 10.0, Category: "Test", InStock: true},
		{Name: "Product 2", Price: 20.0, Category: "Test", InStock: false},
		{Name: "Product 3", Price: 30.0, Category: "Test", InStock: true},
	}

	for _, product := range testProducts {
		if _, err := mockDB.CreateProduct(product); err != nil {
			t.Fatalf("Failed to create test product: %v", err)
		}
	}

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedTotal  int
		expectedSize   int
	}{
		{
			name:           "Default pagination",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedTotal:  3,
			expectedSize:   3,
		},
		{
			name:           "Custom page size",
			queryParams:    "?page=1&page_size=2",
			expectedStatus: http.StatusOK,
			expectedTotal:  3,
			expectedSize:   2,
		},
		{
			name:           "Second page",
			queryParams:    "?page=2&page_size=2",
			expectedStatus: http.StatusOK,
			expectedTotal:  3,
			expectedSize:   1,
		},
		{
			name:           "Page beyond data",
			queryParams:    "?page=10&page_size=10",
			expectedStatus: http.StatusOK,
			expectedTotal:  3,
			expectedSize:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/products"+tt.queryParams, nil)
			rr := httptest.NewRecorder()

			handler.GetProducts(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, status)
			}

			var response models.PaginatedResponse
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Total != tt.expectedTotal {
				t.Errorf("Expected total %d, got %d", tt.expectedTotal, response.Total)
			}

			if len(response.Data) != tt.expectedSize {
				t.Errorf("Expected %d products, got %d", tt.expectedSize, len(response.Data))
			}
		})
	}
}

func TestGetProductsError(t *testing.T) {
	mockDB := newMockDB()
	mockDB.shouldFail = true
	handler := NewHandler(mockDB)

	req := httptest.NewRequest("GET", "/api/v1/products", nil)
	rr := httptest.NewRecorder()

	handler.GetProducts(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, status)
	}

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if errorResponse.Error != "Failed to retrieve products" {
		t.Errorf("Expected error message 'Failed to retrieve products', got %s", errorResponse.Error)
	}
}

func TestGetProduct(t *testing.T) {
	mockDB := newMockDB()
	handler := NewHandler(mockDB)

	// Add a test product
	req := models.CreateProductRequest{
		Name:        "Test Product",
		Description: "A test product",
		Price:       99.99,
		Category:    "Test",
		InStock:     true,
	}
	product, _ := mockDB.CreateProduct(req)

	tests := []struct {
		name           string
		productID      string
		dbShouldFail   bool
		expectedStatus int
		expectedName   string
	}{
		{
			name:           "Valid product ID",
			productID:      "1",
			expectedStatus: http.StatusOK,
			expectedName:   "Test Product",
		},
		{
			name:           "Invalid product ID",
			productID:      "9999999999999999999", // exceeds int range
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent product",
			productID:      "999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Database error",
			productID:      "1",
			dbShouldFail:   true,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB.shouldFail = tt.dbShouldFail

			req := httptest.NewRequest("GET", "/api/v1/products/"+tt.productID, nil)
			rr := httptest.NewRecorder()

			// Set up router to parse URL parameters
			router := mux.NewRouter()
			router.HandleFunc("/api/v1/products/{id:[0-9]+}", handler.GetProduct).Methods("GET")
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, status)
			}

			if tt.expectedStatus == http.StatusOK {
				var response models.Product
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.Name != tt.expectedName {
					t.Errorf("Expected product name %s, got %s", tt.expectedName, response.Name)
				}

				if response.ID != product.ID {
					t.Errorf("Expected product ID %d, got %d", product.ID, response.ID)
				}
			}
		})
	}
}

func TestCreateProduct(t *testing.T) {
	mockDB := newMockDB()
	handler := NewHandler(mockDB)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedName   string
	}{
		{
			name: "Valid product",
			requestBody: models.CreateProductRequest{
				Name:        "New Product",
				Description: "A new product",
				Price:       49.99,
				Category:    "Electronics",
				InStock:     true,
			},
			expectedStatus: http.StatusCreated,
			expectedName:   "New Product",
		},
		{
			name: "Missing required fields",
			requestBody: models.CreateProductRequest{
				Description: "Missing name and price",
				Category:    "Electronics",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid price",
			requestBody: models.CreateProductRequest{
				Name:     "Invalid Price Product",
				Price:    -10.0,
				Category: "Electronics",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest("POST", "/api/v1/products", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.CreateProduct(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, status)
			}

			if tt.expectedStatus == http.StatusCreated {
				var response models.Product
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.Name != tt.expectedName {
					t.Errorf("Expected product name %s, got %s", tt.expectedName, response.Name)
				}

				if response.ID == 0 {
					t.Error("Product ID should not be 0")
				}
			}
		})
	}
}

func TestCreateProductDatabaseError(t *testing.T) {
	mockDB := newMockDB()
	mockDB.shouldFail = true
	handler := NewHandler(mockDB)

	requestBody := models.CreateProductRequest{
		Name:        "Test Product",
		Description: "A test product",
		Price:       99.99,
		Category:    "Test",
		InStock:     true,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/products", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.CreateProduct(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, status)
	}
}

func TestUpdateProduct(t *testing.T) {
	mockDB := newMockDB()
	handler := NewHandler(mockDB)

	// Create a product to update
	createReq := models.CreateProductRequest{
		Name:        "Original Product",
		Description: "Original description",
		Price:       100.0,
		Category:    "Original",
		InStock:     true,
	}
	if _, err := mockDB.CreateProduct(createReq); err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	tests := []struct {
		name           string
		productID      string
		requestBody    interface{}
		dbShouldFail   bool
		expectedStatus int
		expectedName   string
	}{
		{
			name:      "Valid update",
			productID: "1",
			requestBody: models.UpdateProductRequest{
				Name:  byref("Updated Product"),
				Price: byref(150.0),
			},
			expectedStatus: http.StatusOK,
			expectedName:   "Updated Product",
		},
		{
			name:      "Partial update",
			productID: "1",
			requestBody: models.UpdateProductRequest{
				InStock: byref(false),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Non-existent product",
			productID: "999",
			requestBody: models.UpdateProductRequest{
				Name: byref("Should not work"),
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:      "Invalid product ID",
			productID: "9999999999999999999", // exceeds int range
			requestBody: models.UpdateProductRequest{
				Name: byref("Should not work"),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			productID:      "1",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Invalid price",
			productID: "1",
			requestBody: models.UpdateProductRequest{
				Price: byref(-50.0),
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Database error",
			productID:      "1",
			requestBody:    models.UpdateProductRequest{Name: byref("Should not work")},
			dbShouldFail:   true,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			mockDB.shouldFail = tt.dbShouldFail

			req := httptest.NewRequest("PUT", "/api/v1/products/"+tt.productID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			// Set up router to parse URL parameters
			router := mux.NewRouter()
			router.HandleFunc("/api/v1/products/{id:[0-9]+}", handler.UpdateProduct).Methods("PUT")
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, status)
			}

			if tt.expectedStatus == http.StatusOK && tt.expectedName != "" {
				var response models.Product
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.Name != tt.expectedName {
					t.Errorf("Expected product name %s, got %s", tt.expectedName, response.Name)
				}
			}
		})
	}
}

func TestDeleteProduct(t *testing.T) {
	mockDB := newMockDB()
	handler := NewHandler(mockDB)

	// Create a product to delete
	createReq := models.CreateProductRequest{
		Name:        "Product to Delete",
		Description: "Will be deleted",
		Price:       50.0,
		Category:    "Test",
		InStock:     true,
	}
	if _, err := mockDB.CreateProduct(createReq); err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	tests := []struct {
		name           string
		productID      string
		dbShouldFail   bool
		expectedStatus int
	}{
		{
			name:           "Valid deletion",
			productID:      "1",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Non-existent product",
			productID:      "999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid product ID",
			productID:      "9999999999999999999", // exceeds int range
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Database error",
			productID:      "1",
			dbShouldFail:   true,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB.shouldFail = tt.dbShouldFail

			req := httptest.NewRequest("DELETE", "/api/v1/products/"+tt.productID, nil)
			rr := httptest.NewRecorder()

			// Set up router to parse URL parameters
			router := mux.NewRouter()
			router.HandleFunc("/api/v1/products/{id:[0-9]+}", handler.DeleteProduct).Methods("DELETE")
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, status)
			}

			if tt.expectedStatus == http.StatusNoContent {
				if rr.Body.Len() != 0 {
					t.Error("Expected empty response body for successful deletion")
				}
			}
		})
	}
}

func TestSetupRoutes(t *testing.T) {
	realDB := db.NewInMemoryDB()
	handler := NewHandler(realDB)
	router := handler.SetupRoutes()

	// Test that routes are properly configured
	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/products"},
		{"POST", "/api/v1/products"},
		{"GET", "/api/v1/products/1"},
		{"PUT", "/api/v1/products/1"},
		{"DELETE", "/api/v1/products/1"},
		{"GET", "/health"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.method, tt.path), func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			// We don't expect 404 for these routes
			if rr.Code == http.StatusNotFound {
				t.Errorf("Route %s %s not found", tt.method, tt.path)
			}
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	mockDB := newMockDB()
	handler := NewHandler(mockDB)
	router := handler.SetupRoutes()

	// Test OPTIONS request on a valid route
	req := httptest.NewRequest("OPTIONS", "/api/v1/products", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d for OPTIONS request, got %d", http.StatusOK, rr.Code)
	}

	// Check CORS headers
	headers := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Authorization",
	}

	for header, expectedValue := range headers {
		actualValue := rr.Header().Get(header)
		if actualValue != expectedValue {
			t.Errorf("Expected %s header to be %s, got %s", header, expectedValue, actualValue)
		}
	}

	// Test that CORS headers are also present on regular requests
	req = httptest.NewRequest("GET", "/api/v1/products", nil)
	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	corsOrigin := rr.Header().Get("Access-Control-Allow-Origin")
	if corsOrigin != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin header on GET request, got %s", corsOrigin)
	}
}

// Helper function for creating pointers to literals
func byref[T any](v T) *T {
	return &v
}
