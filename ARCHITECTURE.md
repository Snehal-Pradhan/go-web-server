# Architectural Flow

## Layer Diagram

```
┌──────────────────────────────────────────────────────┐
│                    main.go                           │
│   server.Server() → ListenAndServe()                 │
└────────────┬─────────────────────────────────────────┘
             │
             ▼
┌──────────────────────────────────────────────────────┐
│               server/server.go                       │
│                                                      │
│  init():                                             │
│    connectToInMemoryDB() ─► buntdb (in-memory)       │
│         │                                            │
│         ▼                                            │
│    repository.NewInMemoryRepository(db)              │
│         │                                            │
│         ▼                                            │
│    book.NewService(repository)                       │
│         │                                            │
│         ▼                                            │
│    book2.NewBookHandler(service)                     │
│         │                                            │
│  Server():                                           │
│    r := mux.NewRouter()                              │
│    r.Handle("/book",      POST)  ──► bookHandler     │
│    r.Handle("/book/{id}", GET)   ──► bookHandler     │
│    return &http.Server{Handler: r}                   │
└────────────┬─────────────────────────────────────────┘
             │  HTTP Request arrives
             ▼
┌──────────────────────────────────────────────────────┐
│          server/handlers/handler.go                  │
│                                                      │
│  handlers.Handler (type):                            │
│    func(w, r) error                                  │
│                                                      │
│  ServeHTTP(w, r):         ◄── implements http.Handler│
│    err := h(w, r)         ◄── calls your handler fn  │
│    if err != nil:                                    │
│      respondWithErr(err, w)                          │
│        ├── errors.As(err, &AppError) ──► 400/404     │
│        └── else                  ──► 500             │
└────────────┬─────────────────────────────────────────┘
             │  routes to specific handler
             ▼
┌──────────────────────────────────────────────────────┐
│     server/handlers/book/handler.go                  │
│                                                      │
│  CreateBook(w, r):                                   │
│    decoder.DecodeJSON(r.Body, &req)                  │
│    bookService.CreateBook(req.Title)                 │
│    responses.OK("msg", data).ToJSON(w)               │
│                                                      │
│  GetBook(w, r):                                      │
│    mux.Vars(r)["bookID"]                             │
│    bson.IsObjectIdHex(id) ──► validate               │
│    bookService.GetBook(id)                           │
│    responses.OK("msg", data).ToJSON(w)               │
└────────────┬─────────────────────────────────────────┘
             │  delegates business logic
             ▼
┌──────────────────────────────────────────────────────┐
│           services/book/service.go                   │
│                                                      │
│  Service (interface):                                │
│    GetBook(id)    (*Book, error)                     │
│    CreateBook(title)  (*Book, error)                 │
│                                                      │
│  service (impl):                                     │
│    CreateBook:                                       │
│      book = Book{ID, Title, time.Now()}              │
│      repository.CreateBook(book)                     │
│                                                      │
│    GetBook:                                          │
│      repository.GetBook(id)                          │
└────────────┬─────────────────────────────────────────┘
             │  persistence
             ▼
┌──────────────────────────────────────────────────────┐
│   services/book/repository/repository.go             │
│                                                      │
│  Repository (interface):                             │
│    GetBook(id)    (*Book, error)                     │
│    CreateBook(book)  error                           │
│                                                      │
│  inMemoryRepository (impl via buntdb):               │
│    CreateBook: tx.Set("books::<id>", JSON, nil)      │
│    GetBook:    tx.Get("books::<id>") → JSON.Unmarshal│
└──────────────────────────────────────────────────────┘
```

## Request Lifecycle (POST /book)

```
 Client                     Server
   │                          │
   │  POST /book              │
   │  {"title":"Go"}          │
   │─────────────────────────►│
   │                          │
   │                ┌─────────▼──────────┐
   │                │  gorilla/mux       │
   │                │  matches /book     │
   │                └─────────┬──────────┘
   │                          │
   │                ┌─────────▼──────────┐
   │                │  handlers.Handler  │
   │                │  .ServeHTTP(w, r)  │
   │                └─────────┬──────────┘
   │                          │
   │                ┌─────────▼──────────┐
   │                │  book.handler      │
   │                │  .CreateBook(w, r) │
   │                │                    │
   │                │  1. DecodeJSON     │
   │                │     r.Body → title │
   │                │                    │
   │                │  2. bookService    │
   │                │     .CreateBook()  │
   │                │       │            │
   │                │       ├─► Generate │
   │                │       │   ID + now │
   │                │       │            │
   │                │       ├─► repo     │
   │                │       │   .Create  │
   │                │       │   (buntdb) │
   │                │       │            │
   │                │  3. responses.OK() │
   │                │     .ToJSON(w)     │
   │                └─────────┬──────────┘
   │                          │
   │  {"success":true,        │
   │   "data":{"book":{...}}  │
   │◄─────────────────────────┤
   │                          │
```

## Error Flow

```
 Handler returns error
       │
       ▼
 ┌─────────────────┐     yes     ┌──────────────────────┐
 │ errors.As(err,  │───────────►│ responses.Fail(       │
 │ &AppError)?     │            │   toSentenceCase(),   │
 └─────────────────┘            │   statusCode          │
       │                        │ ).ToJSON(w)           │
       │ no                     └──────────────────────┘
       ▼
 ┌─────────────────┐
 │ log.Println()   │
 │ responses.Fail( │
 │   "Internal...",│
 │   500           │
 │ ).ToJSON(w)     │
 └─────────────────┘
```

## Dependency Injection Chain

```
main.go
  │
  ▼
server.Server() ──► init()
                        │
 buntdb.Open() ◄────────┤
   │                    │
   ▼                    │
 repository.NewInMemory │
   │                    │
   ▼                    │
 book.NewService        │
   │                    │
   ▼                    │
 book2.NewBookHandler   │
   │                    │
   ▼                    │
 handlers.Handler ──────┤  (wraps handler methods)
                        │
 mux.NewRouter() ◄──────┤
                        │
 &http.Server ◄─────────┘
```

## Package Dependency Graph

```
main
  └── server
        ├── handlers ──── errors
        │     └── book ── errors, decoder, responses, services/book
        ├── decoder ───── errors
        ├── responses
        ├── inmemory
        └── services/book
              ├── models
              └── services/book/repository ──── errors, models
```
