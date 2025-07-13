# Code Review

```go
   var users = make(map[string]string)

   func createUser(name string) {
      users[name] = time.Now().String()
   }

   func handleRequest(w http.ResponseWriter, r *http.Request) {
      name := r.URL.Query().Get("name")
      go createUser(name)
      w.WriteHeader(http.StatusOK)      
   }
```

- the `users` map is accessed concurrently by multiple goroutines without synchronization;
  synchronization should be used or a concurrent map

- context should be used for request cancellation and timeout; context can also be a useful
  vector for carrying request-scoped values and dependencies such as a logger, clock etc

- there is no validation of the `name` parameter, which could lead to issues such as
  empty names or names that are not valid identifiers

- use of `time.Now()` makes the code non-deterministic, which can lead to issues in testing
  or debugging; the use case here is clearly synthetic, but a package should be used that
  allows for mockable time sources

- if `createUser` were a more complex operation, invoking it in a goroutine
  without recovery or error handling could lead to panics or errors that are not
  identified or handled; panics/errors should be logged and handled appropriately

- `createUser` provides no mechanism for error handling or reporting; the caller has no
  way of knowing if a user was created or not

- a `200 OK` response is returned immediately (without the resource), but the caller has
  no way of knowing whether the user was created successfully; with the example implementation,
  any of `201 Created`, `202 Accepted` or `204 No Content` might be a more appropriate
  response.  In a more realistic scenario, `202 Accepted` is probably the most appropriate
  response, with a response that provides information that the client can use to verify
  the outcome
