#!/bin/bash

# Products API Demo Script
# This script demonstrates the various API endpoints

API_BASE="http://localhost:8080"

echo "🚀 Products API Demo"
echo "==================="
echo

# Health Check
echo "1. Health Check:"
echo "GET /health"
curl -s "$API_BASE/health" | jq '.' 2>/dev/null || curl -s "$API_BASE/health"
echo
echo

# Get all products (paginated)
echo "2. Get All Products (Page 1, Size 3):"
echo "GET /api/v1/products?page=1&page_size=3"
curl -s "$API_BASE/api/v1/products?page=1&page_size=3" | jq '.' 2>/dev/null || curl -s "$API_BASE/api/v1/products?page=1&page_size=3"
echo
echo

# Get filtered products
echo "3. Get Products in Category (Office Supplies):"
echo "GET /api/v1/products?category=office%20supplies"
curl -s "$API_BASE/api/v1/products?category=office%20supplies" | jq '.' 2>/dev/null || curl -s "$API_BASE/api/v1/products?category=office%20supplies"
echo
echo

# Get specific product
echo "4. Get Product by ID (ID: 1):"
echo "GET /api/v1/products/1"
curl -s "$API_BASE/api/v1/products/1" | jq '.' 2>/dev/null || curl -s "$API_BASE/api/v1/products/1"
echo
echo

# Create new product
echo "5. Create New Product:"
echo "POST /api/v1/products"
echo "Body: {\"name\": \"Demo Product\", \"description\": \"Created via demo\", \"price\": 99.99, \"category\": \"Demo\", \"in_stock\": true}"
NEW_PRODUCT=$(curl -s -X POST "$API_BASE/api/v1/products" \
    -H "Content-Type: application/json" \
    -d '{"name": "Demo Product", "description": "Created via demo", "price": 99.99, "category": "Demo", "in_stock": true}')
echo "$NEW_PRODUCT" | jq '.' 2>/dev/null || echo "$NEW_PRODUCT"

# Extract ID from the new product (if jq is available)
NEW_ID=$(echo "$NEW_PRODUCT" | jq -r '.id' 2>/dev/null || echo "6")
echo
echo

# Update product
echo "6. Update Product (ID: $NEW_ID):"
echo "PUT /api/v1/products/$NEW_ID"
echo "Body: {\"name\": \"Updated Demo Product\", \"price\": 149.99}"
curl -s -X PUT "$API_BASE/api/v1/products/$NEW_ID" \
    -H "Content-Type: application/json" \
    -d '{"name": "Updated Demo Product", "price": 149.99}' | jq '.' 2>/dev/null ||
    curl -s -X PUT "$API_BASE/api/v1/products/$NEW_ID" \
        -H "Content-Type: application/json" \
        -d '{"name": "Updated Demo Product", "price": 149.99}'
echo
echo

# Get updated product
echo "7. Verify Update - Get Product by ID ($NEW_ID):"
echo "GET /api/v1/products/$NEW_ID"
curl -s "$API_BASE/api/v1/products/$NEW_ID" | jq '.' 2>/dev/null || curl -s "$API_BASE/api/v1/products/$NEW_ID"
echo
echo

# Delete product
echo "8. Delete Product (ID: $NEW_ID):"
echo "DELETE /api/v1/products/$NEW_ID"
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$API_BASE/api/v1/products/$NEW_ID")
echo "HTTP Status: $HTTP_STATUS"
echo
echo

# Verify deletion
echo "9. Verify Deletion - Try to Get Deleted Product:"
echo "GET /api/v1/products/$NEW_ID"
curl -s "$API_BASE/api/v1/products/$NEW_ID" | jq '.' 2>/dev/null || curl -s "$API_BASE/api/v1/products/$NEW_ID"
echo
echo

echo "✅ Demo completed!"
echo "💡 Note: If output is not formatted nicely, install 'jq' for better JSON formatting: brew install jq"
