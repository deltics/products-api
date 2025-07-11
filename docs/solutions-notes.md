
# API Handlers

## ID Routes

The implementation uses Gorilla Mux with regex patterns to handle API requests. This provides
routes that reject invalid IDs without providing helpful responses or error messages.  i.e.
`/products/abc` will not match a route and so will return a `404 Not Found` without indicating
that the ID is invalid.

It is this behaviour that required the test cases for invalid IDs to be updated to use an
out of range integer value to trigger the `http.StatusBadRequest` response to improve the
test coverage.

This may be a valid design choice, but is not especially helpful for API consumers. An alternative
approach might be to relax the route matching and implement explicit parameter validation
within the handler functions to allow for more informative error responses.

## Not Found vs Database Errors

In the initial implementation, the error handling for "not found" and other database errors was
conflated, making it difficult to distinguish between a product not being found and other
database-related issues.

A ErrProductNotFound sentinel error was introduced in the database layer to distinguish
between these two scenarios, enabling API handlers to differentiate between errors and
provide appropriate and more informative responses.
