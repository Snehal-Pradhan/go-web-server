package errors

type Type string


const (
	TypeBadRequest Type = "bad_request_error"
	TypeNotFound Type = "not_found_error"
)


type AppError struct {
	text string
	errType Type
}


func (e AppError) Error() string {
	return e.text 
}

func (e AppError) Type() Type {
	return e.errType
}

func Error(text string) error {
	return &AppError{text: text, errType: TypeBadRequest}
}

func NotFound(entity string) error {
	return &AppError{text:entity+" not found",errType: TypeNotFound}
}

