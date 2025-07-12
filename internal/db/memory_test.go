package db

import (
	"errors"
	"testing"

	"products-api/internal/models"
)

func TestNewInMemoryDB(t *testing.T) {
	db := NewInMemoryDB()

	if db == nil {
		t.Fatal("NewInMemoryDB() should not return nil")
	}

	if db.nextID != 6 { // Should be 6 after adding 5 sample products
		t.Errorf("Expected nextID to be 6, got %d", db.nextID)
	}

	if len(db.products) != 5 {
		t.Errorf("Expected 5 sample products, got %d", len(db.products))
	}
}

func TestCreateProduct(t *testing.T) {
	db := NewInMemoryDB()
	initialCount := len(db.products)

	req := models.CreateProductRequest{
		Name:        "Test Product",
		Description: "A test product",
		Price:       99.99,
		Category:    "Test",
		InStock:     true,
	}

	product, err := db.CreateProduct(req)
	if err != nil {
		t.Fatalf("CreateProduct() failed: %v", err)
	}

	if product.ID == 0 {
		t.Error("Product ID should not be 0")
	}

	if product.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, product.Name)
	}

	if product.Price != req.Price {
		t.Errorf("Expected price %f, got %f", req.Price, product.Price)
	}

	if product.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}

	if product.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}

	if len(db.products) != initialCount+1 {
		t.Errorf("Expected %d products, got %d", initialCount+1, len(db.products))
	}

	// Verify the product is actually stored
	stored, err := db.GetProductByID(product.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve created product: %v", err)
	}

	if stored.Name != req.Name {
		t.Errorf("Stored product name mismatch: expected %s, got %s", req.Name, stored.Name)
	}
}

func TestGetProductByID(t *testing.T) {
	db := NewInMemoryDB()

	// Test getting existing product
	product, err := db.GetProductByID(1)
	if err != nil {
		t.Fatalf("GetProductByID(1) failed: %v", err)
	}

	if product.ID != 1 {
		t.Errorf("Expected product ID 1, got %d", product.ID)
	}

	if product.Name != "Laptop" {
		t.Errorf("Expected product name 'Laptop', got %s", product.Name)
	}

	// Test getting non-existent product
	_, err = db.GetProductByID(999)
	if err == nil {
		t.Error("Expected error for non-existent product")
	}

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected 'not found' error, got %s", err.Error())
	}

	// Test that returned product is a copy (modification doesn't affect original)
	originalName := product.Name
	product.Name = "Modified Name"

	retrievedAgain, _ := db.GetProductByID(1)
	if retrievedAgain.Name != originalName {
		t.Error("Product should be returned as a copy to prevent external modifications")
	}
}

func TestGetProducts(t *testing.T) {
	db := NewInMemoryDB()

	// Test getting all products (first page)
	products, total, err := db.GetProducts(1, 10)
	if err != nil {
		t.Fatalf("GetProducts() failed: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}

	if len(products) != 5 {
		t.Errorf("Expected 5 products, got %d", len(products))
	}

	// Test pagination
	products, total, err = db.GetProducts(1, 2)
	if err != nil {
		t.Fatalf("GetProducts() with pagination failed: %v", err)
	}

	if len(products) != 2 {
		t.Errorf("Expected 2 products per page, got %d", len(products))
	}

	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}

	// Test second page
	products, _, err = db.GetProducts(2, 2)
	if err != nil {
		t.Fatalf("GetProducts() second page failed: %v", err)
	}

	if len(products) != 2 {
		t.Errorf("Expected 2 products on second page, got %d", len(products))
	}

	// Test page beyond available data
	products, total, err = db.GetProducts(10, 10)
	if err != nil {
		t.Fatalf("GetProducts() beyond available data failed: %v", err)
	}

	if len(products) != 0 {
		t.Errorf("Expected 0 products for page beyond data, got %d", len(products))
	}

	if total != 5 {
		t.Errorf("Expected total 5 even for empty page, got %d", total)
	}

	// Test invalid page/pageSize
	products, _, err = db.GetProducts(0, 0)
	if err != nil {
		t.Fatalf("GetProducts() with invalid params failed: %v", err)
	}

	// Should default to page 1, pageSize 10
	if len(products) != 5 {
		t.Errorf("Expected 5 products with default params, got %d", len(products))
	}

	// Test products are sorted by ID
	for i := 1; i < len(products); i++ {
		if products[i-1].ID >= products[i].ID {
			t.Error("Products should be sorted by ID in ascending order")
			break
		}
	}
}

func TestUpdateProduct(t *testing.T) {
	db := NewInMemoryDB()

	// Test updating existing product
	updateReq := models.UpdateProductRequest{
		Name:  stringPtr("Updated Laptop"),
		Price: float64Ptr(1499.99),
	}

	product, err := db.UpdateProduct(1, updateReq)
	if err != nil {
		t.Fatalf("UpdateProduct() failed: %v", err)
	}

	if product.Name != "Updated Laptop" {
		t.Errorf("Expected updated name 'Updated Laptop', got %s", product.Name)
	}

	if product.Price != 1499.99 {
		t.Errorf("Expected updated price 1499.99, got %f", product.Price)
	}

	// Original description should remain unchanged
	if product.Description != "High-performance laptop for professional use" {
		t.Error("Description should remain unchanged when not specified in update")
	}

	// UpdatedAt should be changed
	originalProduct, _ := db.GetProductByID(1)
	if !originalProduct.UpdatedAt.After(originalProduct.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt after update")
	}

	// Test updating non-existent product
	_, err = db.UpdateProduct(999, updateReq)
	if err == nil {
		t.Error("Expected error when updating non-existent product")
	}

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected 'not found' error, got %s", err.Error())
	}

	// Test partial update
	partialUpdate := models.UpdateProductRequest{
		InStock:     boolPtr(false),
		Description: stringPtr("Updated description for Laptop"),
		Category:    stringPtr("Updated Category"),
	}

	product, err = db.UpdateProduct(1, partialUpdate)
	if err != nil {
		t.Fatalf("Partial update failed: %v", err)
	}

	if product.InStock != false {
		t.Error("InStock should be updated to false")
	}

	if product.Description != "Updated description for Laptop" {
		t.Error("Description should be updated to 'Updated description for Laptop'")
	}

	if product.Category != "Updated Category" {
		t.Error("Category should be updated to 'Updated Category'")
	}

	if product.Name != "Updated Laptop" {
		t.Error("Previously updated fields should remain unchanged")
	}
}

func TestDeleteProduct(t *testing.T) {
	db := NewInMemoryDB()
	initialCount := len(db.products)

	// Test deleting existing product
	err := db.DeleteProduct(1)
	if err != nil {
		t.Fatalf("DeleteProduct() failed: %v", err)
	}

	if len(db.products) != initialCount-1 {
		t.Errorf("Expected %d products after deletion, got %d", initialCount-1, len(db.products))
	}

	// Verify product is actually deleted
	_, err = db.GetProductByID(1)
	if err == nil {
		t.Error("Expected error when getting deleted product")
	}

	// Test deleting non-existent product
	err = db.DeleteProduct(999)
	if err == nil {
		t.Error("Expected error when deleting non-existent product")
	}

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expected 'not found' error, got %s", err.Error())
	}

	// Test deleting same product twice
	err = db.DeleteProduct(1)
	if err == nil {
		t.Error("Expected error when deleting already deleted product")
	}
}

func TestConcurrentAccess(t *testing.T) {
	db := NewInMemoryDB()
	done := make(chan bool, 4)

	// Test concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			_, _, err := db.GetProducts(1, 10)
			if err != nil {
				t.Errorf("Concurrent read failed: %v", err)
			}
		}
		done <- true
	}()

	// Test concurrent writes
	go func() {
		for i := 0; i < 50; i++ {
			req := models.CreateProductRequest{
				Name:        "Concurrent Product",
				Description: "Created concurrently",
				Price:       float64(i + 1),
				Category:    "Test",
				InStock:     true,
			}
			_, err := db.CreateProduct(req)
			if err != nil {
				t.Errorf("Concurrent create failed: %v", err)
			}
		}
		done <- true
	}()

	// Test concurrent updates
	go func() {
		for i := 0; i < 50; i++ {
			updateReq := models.UpdateProductRequest{
				Price: float64Ptr(float64(i + 100)),
			}
			_, err := db.UpdateProduct(2, updateReq)
			if err != nil && !errors.Is(err, ErrNotFound) {
				t.Errorf("Concurrent update failed: %v", err)
			}
		}
		done <- true
	}()

	// Test concurrent deletes and creates
	go func() {
		for i := 0; i < 25; i++ {
			// Try to delete (might fail if already deleted)
			if err := db.DeleteProduct(3); err != nil && !errors.Is(err, ErrNotFound) {
				t.Errorf("Concurrent delete failed: %v", err)
			}

			// Create new product
			req := models.CreateProductRequest{
				Name:        "Temp Product",
				Description: "Temporary",
				Price:       1.0,
				Category:    "Temp",
				InStock:     true,
			}
			if _, err := db.CreateProduct(req); err != nil {
				t.Errorf("Concurrent create failed: %v", err)
			}
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 4; i++ {
		<-done
	}

	// Verify database is still in a consistent state
	products, total, err := db.GetProducts(1, 100)
	if err != nil {
		t.Fatalf("Database inconsistent after concurrent access: %v", err)
	}

	if len(products) != total {
		t.Errorf("Product count mismatch: got %d products but total is %d", len(products), total)
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

func boolPtr(b bool) *bool {
	return &b
}
