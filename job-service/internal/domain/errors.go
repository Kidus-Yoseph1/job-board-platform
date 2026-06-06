package domain

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

// constructors for common errors
func ErrNotFound(msg string) *AppError {
	return &AppError{Code: 404, Message: msg}
}

func ErrBadRequest(msg string) *AppError {
	return &AppError{Code: 400, Message: msg}
}

func ErrUnauthorized(msg string) *AppError {
	return &AppError{Code: 401, Message: msg}
}

func ErrInternal(msg string) *AppError {
	return &AppError{Code: 500, Message: msg}
}
