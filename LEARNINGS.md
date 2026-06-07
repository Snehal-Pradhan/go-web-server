# Q&A

### 1. Are we using MongoDB or just the in-memory database?

Only **buntdb** (in-memory key-value store). The `Repository` interface was designed to support both in-memory and MongoDB, but `server/server.go` only instantiates the in-memory implementation:

```go
bookRepository := repository.NewInMemoryRepository(inMemoryDB)
```

The original code tried to connect to MongoDB on startup, but I removed that because it would block the server if Mongo was not running.

---

### 2. What is the request flow from `main.go` to the handler, and what exactly is a "handler"?

Request flow:

```
main.go → server.Server() → gorilla/mux router → handlers.Handler → book.handler
```

There are **two** handler types:

| Layer | File | Role |
|-------|------|------|
| `handlers.Handler` | `server/handlers/handler.go` | Error-catching wrapper. Implements `http.Handler`. Calls your function and handles any returned error centrally. |
| `book.handler` | `server/handlers/book/handler.go` | Business logic. Contains `CreateBook`/`GetBook`. Returns errors instead of writing error responses inline. |

The key insight: `handlers.Handler` is **middleware** that lets you write your functions as `func(w, r) error` and handles all error responses in one place.

---

### 3. What is gorilla/mux? Is it an ORM?

**No**, it is an **HTTP request router** (multiplexer), not an ORM.

| gorilla/mux | ORM (e.g. GORM) |
|-------------|------------------|
| Routes URLs to handler functions | Maps database tables to Go structs |
| `r.Handle("/book/{id}", h)` | `db.Where("id = ?", id).First(&book)` |

It provides path variables (`{bookID}`), method-based routing (`.Methods("POST")`), and other URL-matching features that Go's built-in `http.ServeMux` lacks.

---

### 4. What is a multiplexer?

A **multiplexer** (or "mux" for short) is a component that takes multiple inputs and routes them to the correct handler based on some criteria. In HTTP servers, the mux receives all incoming requests and decides which handler function should process each one based on the URL path and HTTP method.

Go's `net/http` package has a default `ServeMux`. gorilla/mux is a more feature-rich replacement with support for path variables, method matching, host matching, and subrouters.

Think of it like a **switchboard operator** — a request arrives, the mux looks at the URL, and connects it to the right handler.

---

### 5. The handler's job is to separate error handling from business logic. What is a repository?

**Correct.** The flow of responsibility:

```
Handler  ──►  Service  ──►  Repository
  │                │              │
  │ HTTP concerns  │ Business     │ Data access
  │ (decode JSON,  │ logic        │ (save/find/
  │  write resp.)  │ (validate,   │  delete in
  │  error handler │  orchestrate)│  database)
```

A **Repository** is the layer that talks to the database. It contains all the code for:

- Saving a record (`CreateBook`)
- Fetching a record (`GetBook`)
- (Would also contain) Updating, deleting, listing

The repository hides the database implementation behind an **interface**. The service layer calls `repository.GetBook(id)` without caring whether the data comes from buntdb, MongoDB, a file, or an API. This makes it easy to swap databases or mock them in tests.

---

### 6. What is `responses/response.go` and what does it do?

It defines a **standardized JSON response format** for the entire API. Every response — success or failure — goes through the same structure:

```go
// Every response looks like this in JSON:
{
  "success": true,         // or false
  "message": "...",        // human-readable message
  "data": { ... }         // optional payload
}
```

Two helper functions:

- `responses.OK(msg, data)` — builds a 200 response
- `responses.Fail(msg, statusCode)` — builds an error response with any status code

The `ToJSON(w)` method writes it to the HTTP response with `Content-Type: application/json`. This ensures consistency — no handler writes JSON in a different format.

---

### 7. What is `decoder/decoder.go` and what does it do?

A tiny utility that wraps `json.NewDecoder(r).Decode(v)` and converts JSON parse errors into our custom `errors.Error()` (typed AppError):

```go
func DecodeJSON(r io.Reader, v interface{}) error {
    if err := json.NewDecoder(r).Decode(v); err != nil {
        return errors.Error(err.Error())  // typed as bad request
    }
    return nil
}
```

Without it, handlers would have to write:

```go
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    return errors.Error(err.Error())  // every single time
}
```

With it, handlers just call:

```go
if err := decoder.DecodeJSON(r.Body, &req); err != nil {
    return err  // error is already wrapped with the right type
}
```

It keeps the error type mapping in one place rather than repeating it in every handler.

---

### 8. If we're not using MongoDB, why do we use `bson.ObjectId` as the ID type?

We don't anymore — we replaced it with a plain `string` type. The original project used `bson.ObjectId` because it was designed for both in-memory and MongoDB. Since we only use buntdb, a simple hex string ID (generated via `crypto/rand`) is cleaner and removes the MongoDB dependency entirely.

```go
// models/book.go
type Book struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    CreatedAt time.Time `json:"createdAt"`
}

// service.go — ID generation
func newID() string {
    b := make([]byte, 12)
    rand.Read(b)
    return fmt.Sprintf("%x", b)
}
```
