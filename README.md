# Products API

A RESTful API for managing products built with Go, featuring CRUD operations and pagination.

## Features

- **CRUD Operations**: Create, Read, Update, and Delete products
- **Pagination**: Efficient pagination for product listings
- **In-Memory Database**: Fast in-memory storage with thread-safe operations
- **Validation**: Request validation using go-playground/validator
- **CORS Support**: Cross-origin resource sharing enabled
- **Health Check**: Health check endpoint for monitoring
- **Middleware**: Logging and CORS middleware

## API Endpoints

### Products

- `GET /api/v1/products` - Get all products (paginated)
  - Query parameters:
    - `page` (default: 1) - Page number
    - `page_size` (default: 10, max: 100) - Number of items per page
- `GET /api/v1/products/{id}` - Get a specific product by ID
- `POST /api/v1/products` - Create a new product
- `PUT /api/v1/products/{id}` - Update a specific product
- `DELETE /api/v1/products/{id}` - Delete a specific product

### Health Check

- `GET /health` - Health check endpoint

## Product Model

```json
{
  "id": 1,
  "name": "Laptop",
  "description": "High-performance laptop for professional use",
  "price": 1299.99,
  "category": "Electronics",
  "in_stock": true,
  "created_at": "2025-07-12T10:00:00Z",
  "updated_at": "2025-07-12T10:00:00Z"
}
```

## Running the Application

### Prerequisites

- Go 1.19 or later

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```

### Running

```bash
go run main.go
```

The server will start on port 8080 by default. You can set a custom port using the `PORT` environment variable:

```bash
PORT=3000 go run main.go
```

### Building

```bash
go build -o products-api
./products-api
```

## Example Usage

### Get all products (paginated)
```bash
curl "http://localhost:8080/api/v1/products?page=1&page_size=5"
```

### Get a specific product
```bash
curl "http://localhost:8080/api/v1/products/1"
```

### Create a new product
```bash
curl -X POST "http://localhost:8080/api/v1/products" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "New Product",
    "description": "A great new product",
    "price": 99.99,
    "category": "Electronics",
    "in_stock": true
  }'
```

### Update a product
```bash
curl -X PUT "http://localhost:8080/api/v1/products/1" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Product Name",
    "price": 109.99
  }'
```

### Delete a product
```bash
curl -X DELETE "http://localhost:8080/api/v1/products/1"
```

### Health check
```bash
curl "http://localhost:8080/health"
```

## Project Structure

```
.
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
└── internal/
    ├── api/
    │   └── handlers.go     # HTTP handlers and routing
    ├── db/
    │   └── memory.go       # In-memory database implementation
    └── models/
        └── product.go      # Data models and structs
```

## Sample Data

The application starts with 5 sample products:

1. Laptop - $1299.99
2. Wireless Mouse - $29.99
3. Coffee Mug - $12.50 (out of stock)
4. Desk Chair - $199.99
5. Smartphone - $899.99

## Dependencies

- [Gorilla Mux](https://github.com/gorilla/mux) - HTTP router and URL matcher
- [go-playground/validator](https://github.com/go-playground/validator) - Struct and field validation

## Development

This project follows Go best practices and idiomatic Go code structure:

- Clean separation of concerns with internal packages
- Thread-safe in-memory database with proper locking
- Comprehensive error handling
- Request validation
- Middleware for cross-cutting concerns
- RESTful API design principles
