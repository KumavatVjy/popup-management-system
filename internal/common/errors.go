package common

import "errors"

var (
	ErrNotFound            = errors.New("resource not found")
	ErrDuplicateDomain     = errors.New("website domain already exists")
	ErrBadRequest          = errors.New("bad request")
	ErrValidation          = errors.New("validation failed")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrForbidden           = errors.New("forbidden")
	ErrConflict            = errors.New("conflict")
	ErrInternalServerError = errors.New("internal server error")
)

// AppError represents an application error with an explicit HTTP status code.
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "an error occurred"
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
