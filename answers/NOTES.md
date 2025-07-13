# Notes on Modifications to the Copilot Implementation

## General

### Time Spent

Approximately 8 hours were spent on the implementation over several sessions of 1-2 hours.
The CODE-REVIEW and DESIGN submissions were composed in a further 1-2 hour session.

### Use of Copilot

I used Github Copilot to generate an initial implementation for the API handlers and
database layer.  The generated code was a good starting point, but required significant
modifications to improve error handling and test coverage.

The graceful shutdown and RateLimiter implementation were implemented without Copilot
assistance (other than code completion suggestions).

### Linting

Code was linted using `golangci-lint` (v2) with default configuration.  All lint issues were
addressed as required.

## API Handlers

### ID Routes

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

### Not Found vs Database Errors

In the initial implementation, the error handling for "not found" and other database errors was
conflated, making it difficult to distinguish between a product not being found and other
database-related issues.

A ErrProductNotFound sentinel error was introduced in the database layer to distinguish
between these two scenarios, enabling API handlers to differentiate between errors and
provide appropriate and more informative responses.

### Test Data Creation

Error checks were added to the test data creation steps to ensure that the test setup is robust.

### CORS OPTIONS Route

An OPTIONS handler method was added to the API handler but this method would never be called
since the OPTIONS request was handled by the CORS middleware, resulting in unreachable and
untestable code that static analysis was unable to identify.

The unnecessary OPTIONS handler method was removed, and the Gorilla Mux router configuration
updated with a nil handler for the OPTIONS routes, documenting the use of CORS middleware
for the handling of these routes. This simplifies the code and enables 100% test coverage of
the api package.

## Rate Limiter

A simple rate limiter was implemented to limit requests from clients identified by remote
IP address to a specified number of requests per interval.

### Rate Limiter Testing

To simplify api tests, a NOOP limiter is provided.

I have used my own github.com/blugnu/time package to provide a mockable time source
used by the rate limiter and tests.  This allows for deterministic testing of rate
limiting behaviour without relying on the passage of real time.

The api handler implements a middleware to apply the rate limiter to incoming requests.

### Use of Channels

The only channels used in the rate limiter implementation are those provided by the
cancellable context and the tickers used by the rate limiter goroutines to manage
the rate limiting logic and client cleanup.  No additional channels were required.
