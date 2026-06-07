<div>
  <img alt="Project Status" src="https://img.shields.io/badge/Project%20Status-Completed%20/%20Stable-success">
</div>
<h1 align="center">Go Web Server</h1>
<div align="center">
  <p align="center">
    A RESTful Book API server built with Go, backed by an in-memory database. Create and retrieve books through clean JSON endpoints — demonstrating production-ready patterns for building maintainable HTTP services.
  </p>
  <picture>
    <img alt="Go logo" src="public/image.png" width="10%">
  </picture>
  <p align="center">
    <a href="#features">Features</a> ·
    <a href="#architecture">Architecture</a> ·
    <a href="#tech-stack">Tech Stack</a> ·
    <a href="#quick-start">Quick Start</a> ·
    <a href="#project-structure">Structure</a> ·
    <a href="#patterns">Patterns</a>
  </p>
</div>
<br>
> **Project Status:** 🟢 **Complete & Stable.** All core features and test suites are fully implemented and verified. This repository is in maintenance mode.

## Features

| Area | |
|------|--|
| **Layered architecture** | Models, services, repositories, and handlers separated by concern |
| **Custom HTTP handler** | `func(w, r) error` — errors caught in one place, not scattered across handlers |
| **Standardized responses** | Every reply uses the `{success, message, data}` envelope |
| **Typed errors** | `AppError` maps directly to HTTP 400 / 404 — never hard-code status codes |
| **Repository pattern** | Data access behind an interface — swap buntdb for MongoDB without touching business logic |
| **Dependency injection** | Repository → Service → Handler, wired at startup with zero globals |

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

## Tech Stack

| Technology | Purpose |
|------------|---------|
| [Go](https://go.dev/) | Language — stdlib `net/http` for the server |
| [gorilla/mux](https://github.com/gorilla/mux) | HTTP request router (path variables, method matching) |
| [buntdb](https://github.com/tidwall/buntdb) | In-memory key-value store (book persistence) |

## Quick Start

```bash
# Clone and run
git clone <repo-url>
cd go-web-server
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
go-web-server/
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

## Detailed Breakdown

### Handler layer — error centralization

Every HTTP handler returns `error`. The custom `handlers.Handler` type (which implements `http.Handler`) catches that error in `ServeHTTP` and routes it through `respondWithErr`. This means zero error-handling boilerplate in your handler functions — you just return an error and it gets formatted as the right JSON response.

### Service layer — business logic

The service is an interface (`book.Service`) with a single implementation. It generates IDs, timestamps, and delegates persistence to the repository. The handler never touches the database — it only calls the service.

### Repository layer — data access

`repository.Repository` is an interface with two methods: `GetBook` and `CreateBook`. The in-memory implementation uses buntdb (a fast embedded key-value store). Books are serialized to JSON and stored under keys like `books::<id>`. Because the service depends on the interface, swapping to PostgreSQL, MongoDB, or a file backend requires zero changes to business logic.

### Error types — semantic HTTP mapping

`AppError` carries a `Type` field (`TypeBadRequest` or `TypeNotFound`). The central `respondWithErr` function uses `errors.As` to unwrap the error, checks its type, and writes the correct HTTP status code. Unknown/unexpected errors always return 500 with a generic message — no information leakage.

### Responses — consistent envelope

Every endpoint writes through `responses.OK()` or `responses.Fail()`. The JSON shape is always `{ "success": bool, "message": string, "data": ... }`. The `ToJSON()` method sets `Content-Type: application/json` and writes the status code in one place — no handler can accidentally omit the header or use a different format.

 

<div align="center">
  <sub>Built with Go</sub>
</div>
