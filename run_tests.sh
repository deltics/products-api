#!/bin/bash

echo "🧪 Products API Test Suite"
echo "=========================="
echo

echo "📋 Running all unit tests..."
go test ./... -v

echo
echo "📊 Running tests with coverage..."
go test ./... -coverprofile=coverage.out

echo
echo "📈 Coverage Summary:"
go tool cover -html=coverage.out
