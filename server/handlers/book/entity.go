package book

import "github.com/you/go-web-server/models"

type getBookResponse struct {
	Book *models.Book `json:"book"`
}

type createBookResponse struct {
	Book *models.Book `json:"book"`
}