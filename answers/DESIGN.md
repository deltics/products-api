
# Design and Architecture Questions

## How would you manage high-concurrency in a Go micro-service (thousands of requests per second)?

This depends on the specific requirements and architecture of the service, but generally,
you would want to focus on:

- Efficient use of goroutines and channels to handle concurrent requests;
- Async/Event-driven architecture to decouple components and improve scalability;
- An appropriate rate limiter to prevent abuse and ensure fair usage;
- Make use of a load balancer to distribute traffic across multiple instances, scaling out
  under load;
- If a database is involved, use a connection pool to manage database connections efficiently;
- Implementing caching strategies to reduce load on the database and improve response times;
  (could be in-memory, distributed e.g., Redis, Memcached, or a combination, depending on the use case);

Not all of the above would be necessary for every service.

## Recommended project structure for large Go services?

Project structure should focus on maintainability, scalability, ease of navigation, and
discoverability over adherence to any prescribed architectural pattern.

In my experience, an approach that works and scales well is an approach that might be
described as "fractal"; each package in a project is itself treated as a micro-service.
This allows for independent development, testing, and deployment of each package,
while still being part of a larger application.

This approach inherently scales with the size of a project, maintaining clear separation
of concerns.

```go
   /api
      /repository
          service.go    // implements the concrete repository for api handlers, taking a dependency on a database
       service.go       // implements api routes and handlers, taking a dependency on a repository
   /events
      /consumer
         /repository
             service.go    // implements the concrete repository for event handlers, taking a dependency on a database
          service.go       // implements the concrete event consumer and handlers, taking a dependency on a repository
      /producer
          service.go    // implements the concrete event producer
       service.go       // implements event routes and handlers, taking a dependency on a repository
   /database
       service.go       // implements the concrete database connection for the project
   go.mod
   main.go
```

The example structure above might be typical of a micro-service implementing very specific
domain functionality for a single domain where decomposing `api` and `events` into separate
domain packages would add little value.

In a larger project, additional packages may be introduced to represent separate domains.  Within
each domain package repository and events packages would be created as required to handle the specific
functionality for that domain:

```go
   /api
      /repository
          service.go    // implements the concrete repository for api handlers, taking a dependency on a database
       service.go       // implements api routes and handlers, taking a dependency on a repository
   /domain
      /products
         /repository
             service.go   // implements the concrete repository for product handlers, taking a dependency on a database
          service.go      // implements product routes and handlers, taking a dependency on a repository
      /orders
         /repository
             service.go   // implements the concrete repository for order handlers, taking a dependency on a database
          service.go      // implements order routes and handlers, taking a dependency on a repository
   /events
      /consumer
         /repository
             service.go   // implements the concrete repository for event handlers, taking a dependency on a database
          service.go      // implements the concrete event consumer and handlers, taking a dependency on a repository
      /producer
          service.go    // implements the concrete event producer
       service.go       // implements event routes and handlers, taking a dependency on a repository
   /database
       service.go       // implements the concrete database connection for the project
   go.mod
   main.go
```

The `api` and `events` packages in such a service would implement adapters.  The `api` and `events/consumer`
handlers would route requests to the appropriate domain service, mapping inputs and results (including errors)
to/from the transport as required.  Domain services are implemented without reference to any transport, such
as HTTP, gRPC, Kafka etc.

### Idiomatic Go project structure?

In general I would avoid using so-called "idiomatic" Go project structures such as `cmd`, `internal`, `pkg` etc.
except where they make sense for the specific project.  These structures are common patterns, but not
prescriptive requirements.  Introducing them unnecessarily can make it harder to navigate
a project.

For example, a `cmd` package is not necessary if the project does not have multiple
entry points or commands.  Instead, the `main.go` can be placed at the root of the project.

## Approach to configuration management in production?

Use a combination of pipeline variables and configuration files for static configuration that varies
between environment, supplemented by a configuration repository such as Vault etc for
dynamic/runtime configuration. A service such as Vault or other secrets management solution would
also be used for sensitive data such as database credentials, API keys, etc.

A feature flag service such as LaunchDarkly or Unleash could be used to manage feature toggles.

Database migration infrastructure might be useful if schema changes are anticipated to be frequent or
complex; for less frequent/complex cases, managing migrations as code may be simpler.

## Observability strategy (logging, metrics, tracing)?

Structured logging should be used to provide suitably enriched logs for debugging and
monitoring.  Logs should be written to a centralized logging service such as ELK,
Splunk, etc to enable searching and analysis.

Simple metrics and tracing can be derived from logs, but a more robust solution such as
Prometheus or OpenTelemetry should be used to collect and expose metrics for monitoring
and alerting.  This should include key performance indicators such as request latency,
error rates, and throughput.

Logging and trace spans should be carefully considered to ensure that the information
collected is useful, without being overwhelming.  Too much data in this area can lead
to performance issues and excessive cost as well as result in a low signal-to-noise
ratio.  In the worst case, this can result in critical issues being missed or ignored
due to the sheer volume of logs.

Automated tests should ensure that critical logs are written for key events
and error conditions, to ensure that any alerting or metrics that are driven by logs
are reliable and accurate.

## Go API framework of choice (e.g., Gin, Chi) and why?

In general, I prefer to use the standard library where-ever possible.  The standard
library provides a solid foundation for building APIs, and is well-documented and
maintained.  It is also lightweight and has no external dependencies, which makes it
easier to maintain and deploy.

Where a more capable framework provide tangible benefits, I would always prefer any
which complement the standard library rather than replace it.

I am most familiar with `go-chi`, which I find to be lightweight and flexible, being
wholly compatible with the standard library whilst providing clearer route definitions
and a developer-friendly API for handling request parameters.

`gorilla/mux` was used in the solution for the exercise in this test, and this was my
first exposure to that.  I was interested to see how it compared to `go-chi`
and found it to be comparable for a simple service such as that in the exercise.

### `blugnu/restapi`

I should also mention that I implemented own own
[`blugnu/restapi`](https://github.com/blugnu/restapi) package.  This is not intended as
a full framework, but rather a set of utilities to simplify the implementation of
REST API endpoints.  The main feature of this package is a middleware (more properly
described as an "end-ware" as it must be the last middleware in the chain).

This end-ware removes the need to interact with the `http.ResponseWriter` in an
endpoint, allowing the endpoint to return a response directly.  This provides several:
benefits:

- simplifies implementation of endpoints and makes them more readable
- eliminates boilerplate code for constructing responses
- ensures consistent, configurable response structure, particularly for errors (customizable)
- supports automatic content negotiation based on request headers to support both JSON
  and XML responses
