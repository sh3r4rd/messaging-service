package apperrors

type ServiceError struct {
	Err     error
	Message string
}

func NewServiceError(err error, message string) *ServiceError {
	return &ServiceError{Err: err, Message: message}
}

func (e *ServiceError) Error() string {
	return e.Message
}
