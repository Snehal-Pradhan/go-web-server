package repository

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/buntdb"

	"github.com/you/go-web-server/models"
)

func NewInMemoryRepository(db *buntdb.DB) Repository {
	return inMemoryRepository{db}
}

type inMemoryRepository struct {
	db *buntdb.DB
}

func (r inMemoryRepository) CreateBook(book models.Book) error {
	return r.db.Update(func(tx *buntdb.Tx) error {
		b, err := json.Marshal(&book)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(r.bookKey(book.ID), string(b), nil)
		return err
	})
}

func (r inMemoryRepository) GetBook(id string) (*models.Book, error) {
	var book *models.Book
	err := r.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(r.bookKey(id))
		if err != nil {
			if err == buntdb.ErrNotFound {
				return errBookNotFound
			}
			return err
		}
		return json.Unmarshal([]byte(val), &book)
	})
	if err != nil {
		return nil, err
	}
	return book, nil
}

func (r inMemoryRepository) bookKey(id string) string {
	return fmt.Sprintf("books::%s", id)
}
