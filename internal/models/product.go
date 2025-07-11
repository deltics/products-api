package models

import (
	"time"
)

// Product represents a product in our system
type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	Price       float64   `json:"price" validate:"required,min=0"`
	Category    string    `json:"category"`
	InStock     bool      `json:"in_stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProductRequest represents the request body for creating a product
type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,min=0"`
	Category    string  `json:"category"`
	InStock     bool    `json:"in_stock"`
}

// UpdateProductRequest represents the request body for updating a product
type UpdateProductRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty" validate:"omitempty,min=0"`
	Category    *string  `json:"category,omitempty"`
	InStock     *bool    `json:"in_stock,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       []Product `json:"data"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	Total      int       `json:"total"`
	TotalPages int       `json:"total_pages"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
