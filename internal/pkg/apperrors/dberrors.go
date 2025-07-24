package apperrors

import "errors"

var DBErrorNotFound = NewDBError(
	errors.New("resource not found"),
	"The requested database resource was not found",
)

type DBError struct {
	Err     error
	Message string
}

func NewDBError(err error, message string) *DBError {
	return &DBError{Err: err, Message: message}
}

func (e *DBError) Error() string {
	return e.Message
}
