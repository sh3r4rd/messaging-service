package apperrors

type HTTPError struct {
	Err        error
	StatusCode int
	Message    string
}

func NewHTTPError(err error, statusCode int, message string) *HTTPError {
	return &HTTPError{
		Err:        err,
		StatusCode: statusCode,
		Message:    message,
	}
}

func (e *HTTPError) Error() string {
	return e.Message
}
