package repository

import (
	"github.com/you/go-web-server/errors"
	"github.com/you/go-web-server/models"
)

var errBookNotFound = errors.NotFound("book")

type Repository interface {
	GetBook(id string) (*models.Book, error)
	CreateBook(book models.Book) error
}
