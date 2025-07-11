package db

import (
	"errors"
	"sort"
	"sync"
	"time"

	"products-api/internal/models"
)

// Database interface defines the contract for our database operations
type Database interface {
	GetProducts(page, pageSize int) ([]models.Product, int, error)
	GetProductByID(id int) (*models.Product, error)
	CreateProduct(req models.CreateProductRequest) (*models.Product, error)
	UpdateProduct(id int, req models.UpdateProductRequest) (*models.Product, error)
	DeleteProduct(id int) error
}

// InMemoryDB implements the Database interface using in-memory storage
type InMemoryDB struct {
	products map[int]*models.Product
	nextID   int
	mutex    sync.RWMutex
}

// NewInMemoryDB creates a new in-memory database with some sample data
func NewInMemoryDB() *InMemoryDB {
	db := &InMemoryDB{
		products: make(map[int]*models.Product),
		nextID:   1,
	}

	// Add some sample products
	sampleProducts := []models.CreateProductRequest{
		{
			Name:        "Laptop",
			Description: "High-performance laptop for professional use",
			Price:       1299.99,
			Category:    "Electronics",
			InStock:     true,
		},
		{
			Name:        "Wireless Mouse",
			Description: "Ergonomic wireless mouse with long battery life",
			Price:       29.99,
			Category:    "Electronics",
			InStock:     true,
		},
		{
			Name:        "Coffee Mug",
			Description: "Ceramic coffee mug with company logo",
			Price:       12.50,
			Category:    "Office Supplies",
			InStock:     false,
		},
		{
			Name:        "Desk Chair",
			Description: "Comfortable ergonomic office chair",
			Price:       199.99,
			Category:    "Furniture",
			InStock:     true,
		},
		{
			Name:        "Smartphone",
			Description: "Latest smartphone with advanced camera",
			Price:       899.99,
			Category:    "Electronics",
			InStock:     true,
		},
	}

	for _, req := range sampleProducts {
		db.CreateProduct(req)
	}

	return db
}

// GetProducts returns a paginated list of products
func (db *InMemoryDB) GetProducts(page, pageSize int) ([]models.Product, int, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// Convert map to slice and sort by ID
	products := make([]models.Product, 0, len(db.products))
	for _, product := range db.products {
		products = append(products, *product)
	}

	sort.Slice(products, func(i, j int) bool {
		return products[i].ID < products[j].ID
	})

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

// GetProductByID returns a product by its ID
func (db *InMemoryDB) GetProductByID(id int) (*models.Product, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	product, exists := db.products[id]
	if !exists {
		return nil, errors.New("product not found")
	}

	// Return a copy to prevent external modifications
	productCopy := *product
	return &productCopy, nil
}

// CreateProduct creates a new product
func (db *InMemoryDB) CreateProduct(req models.CreateProductRequest) (*models.Product, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	now := time.Now()
	product := &models.Product{
		ID:          db.nextID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		InStock:     req.InStock,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	db.products[db.nextID] = product
	db.nextID++

	// Return a copy
	productCopy := *product
	return &productCopy, nil
}

// UpdateProduct updates an existing product
func (db *InMemoryDB) UpdateProduct(id int, req models.UpdateProductRequest) (*models.Product, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	product, exists := db.products[id]
	if !exists {
		return nil, errors.New("product not found")
	}

	// Update fields if provided
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

	product.UpdatedAt = time.Now()

	// Return a copy
	productCopy := *product
	return &productCopy, nil
}

// DeleteProduct deletes a product by its ID
func (db *InMemoryDB) DeleteProduct(id int) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	_, exists := db.products[id]
	if !exists {
		return errors.New("product not found")
	}

	delete(db.products, id)
	return nil
}
