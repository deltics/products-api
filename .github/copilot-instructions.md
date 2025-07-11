<!-- Use this file to provide workspace-specific custom instructions to Copilot. For more details, visit https://code.visualstudio.com/docs/copilot/copilot-customization#_use-a-githubcopilotinstructionsmd-file -->

# Products API Copilot Instructions

This is a Go REST API project for managing products with the following characteristics:

## Project Structure

- `main.go`: Application entry point with server setup
- `internal/models/`: Data models and structs (Product, requests, responses)
- `internal/db/`: Database layer with in-memory implementation
- `internal/api/`: HTTP handlers, routing, and middleware

## Key Patterns to Follow

- Use idiomatic Go code with proper error handling
- Follow the existing package structure and naming conventions
- Use the Database interface for data operations
- Implement proper HTTP status codes and JSON responses
- Include validation for request payloads
- Use thread-safe operations for the in-memory database
- Follow RESTful API design principles

## Dependencies

- `github.com/gorilla/mux` for HTTP routing
- `github.com/go-playground/validator/v10` for validation
- Standard library for JSON, HTTP, and other core functionality

## API Design

- All API endpoints are prefixed with `/api/v1`
- Support pagination for list endpoints
- Use proper HTTP methods (GET, POST, PUT, DELETE)
- Include comprehensive error responses
- Provide health check endpoint at `/health`

When adding new features or modifying existing code, maintain consistency with these patterns and ensure thread safety for the in-memory database operations.
