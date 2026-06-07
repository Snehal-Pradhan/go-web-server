package book

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/you/go-web-server/models"
	"github.com/you/go-web-server/services/book/repository"
)

type Service interface {
	GetBook(id string) (*models.Book, error)
	CreateBook(title string) (*models.Book, error)
}

func NewService(repository repository.Repository) Service {
	return service{repository}
}

type service struct {
	repository repository.Repository
}

func (s service) CreateBook(title string) (*models.Book, error) {
	book := models.Book{ID: newID(), Title: title, CreatedAt: time.Now().UTC()}
	if err := s.repository.CreateBook(book); err != nil {
		return nil, err
	}
	return &book, nil
}

func (s service) GetBook(id string) (*models.Book, error) {
	return s.repository.GetBook(id)
}

func newID() string {
	b := make([]byte, 12)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
