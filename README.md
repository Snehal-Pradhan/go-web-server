  <h1 align="center">Go Web Server</h1>
  <p align="center">
    A RESTful Book API server built with Go, backed by an in-memory database. Create and retrieve books through clean JSON endpoints — demonstrating production-ready patterns for building maintainable HTTP services.
  </p>

  <div align="center">
  <picture>
    <img alt="Go logo" src="public/image.png" width="10%">
  </picture>
  <p align="center">
    <a href="#features">Features</a> ·
    <a href="#architecture">Architecture</a> ·
    <a href="#quick-start">Quick Start</a> ·
    <a href="#project-structure">Project Structure</a> ·
    <a href="#patterns">Patterns</a>
  </p>
</div>

&nbsp;

A hands-on exploration of Go web server patterns — layered architecture, custom HTTP handler middleware, typed errors, standardized JSON responses, and dependency injection. Built to learn and demonstrate how to structure a Go HTTP service that scales in complexity without collapsing into spaghetti.

## Features

- **Layered architecture** — models, services, repositories, and handlers separated by concern
- **Custom HTTP handler** — a `handlers.Handler` type that lets you write `func(w, r) error` and catches all errors in one place
- **Standardized JSON responses** — every response uses the same `{success, message, data}` envelope
- **Typed application errors** — custom `AppError` with `TypeBadRequest` / `TypeNotFound` that maps directly to HTTP status codes
- **Repository pattern** — data access behind an interface, swappable between in-memory (buntdb) and any other backend
- **Dependency injection** — repository → service → handler chain wired at startup, no global state

## Architecture

```text
main.go
  │
  └── server.Server()
        │
        ├── gorilla/mux router        ──  route matching
        ├── handlers.Handler          ──  error-catching middleware
        │     └── book.handler        ──  HTTP handlers
        │           ├── decoder       ──  JSON request decoding
        │           └── responses     ──  JSON response formatting
        │
        └── services/book
              ├── service.go          ──  business logic
              └── repository/         ──  data access interface
                    └── inmemory.go   ──  buntdb implementation
```

### Request lifecycle

```text
Client  ──►  gorilla/mux  ──►  handlers.Handler.ServeHTTP()
                                      │
                                      ├── calls book.handler.CreateBook()
                                      │         ├── decoder.DecodeJSON()
                                      │         ├── book.Service.CreateBook()
                                      │         │     └── repository.CreateBook()
                                      │         └── responses.OK().ToJSON()
                                      │
                                      └── on error:
                                            ├── AppError?  ──► 400 / 404
                                            └── unknown?   ──► 500
```

## Quick Start

```bash
# Clone and run
git clone <repo-url>
cd MY_WORK
go run main.go

# Test the API
curl -s http://localhost:8080/book \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"title":"The Go Programming Language"}'

curl -s http://localhost:8080/book/<id-from-response>
```

## Project Structure

```
MY_WORK/
├── main.go                              # Entry point
├── go.mod / go.sum
├── errors/
│   └── error.go                         # Typed AppError (BadRequest, NotFound)
├── models/
│   └── book.go                          # Book data structure
├── services/
│   └── book/
│       ├── service.go                   # Service interface + business logic
│       └── repository/
│           ├── repository.go            # Repository interface
│           └── inmemory.go              # buntdb-backed implementation
└── server/
    ├── server.go                        # Router, DI wiring, init()
    ├── inmemory.go                      # buntdb connection
    ├── decoder/
    │   └── decoder.go                   # JSON decode helper
    ├── responses/
    │   └── response.go                  # Standardized JSON envelope
    └── handlers/
        ├── handler.go                   # Custom Handler type (error wrapper)
        └── book/
            ├── handler.go               # Book HTTP handlers
            └── entity.go                # Request/response DTOs
```

## Patterns

### Custom error-returning handler

Instead of writing `if err != nil` in every handler, define a function type that returns `error` and implement `http.Handler` once:

```go
type Handler func(w http.ResponseWriter, r *http.Request) error

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if err := h(w, r); err != nil {
        respondWithErr(err, w)  // centralized error handling
    }
}
```

### Typed errors → HTTP status codes

```go
// errors/error.go
type AppError struct {
    text    string
    errType Type  // TypeBadRequest or TypeNotFound
}

// In respondWithErr():
errors.As(err, &appError)  // true → 400/404, false → 500
```

### Standardized response envelope

```go
responses.OK("message", data)   // → { "success": true,  "message": "...", "data": {...} }
responses.Fail("message", 400)  // → { "success": false, "message": "..." }
```

### Repository abstraction

```go
type Repository interface {
    GetBook(id string) (*models.Book, error)
    CreateBook(models.Book) error
}
// inmemory.go implements it with buntdb
// mongo.go would implement it with MongoDB — swap without changing services
```
<div align="center">
  <sub>Built with Go</sub>
</div>
