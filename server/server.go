package server

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/you/go-web-server/server/handlers"
	book2 "github.com/you/go-web-server/server/handlers/book"
	"github.com/you/go-web-server/services/book"
	"github.com/you/go-web-server/services/book/repository"
)

var (
	bookHandler book2.Handler
)

func Server() *http.Server {
	r := mux.NewRouter()

	r.Handle("/book", handlers.Handler(bookHandler.CreateBook)).Methods(http.MethodPost)
	r.Handle("/book/{bookID}", handlers.Handler(bookHandler.GetBook)).Methods(http.MethodGet)

	srv := &http.Server{Handler: r, Addr: ":8080"}
	return srv
}

func init() {
	inMemoryDB, err := connectToInMemoryDB()
	fatalIfErr(err)

	bookRepository := repository.NewInMemoryRepository(inMemoryDB)

	bookService := book.NewService(bookRepository)
	bookHandler = book2.NewBookHandler(bookService)
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}