package common

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type APIResponse struct {
	Success    bool              `json:"success"`
	Message    string            `json:"message"`
	Data       interface{}       `json:"data,omitempty"`
	Pagination *Pagination       `json:"pagination,omitempty"`
	Errors     map[string]string `json:"errors,omitempty"`
}

type ValidationError struct {
	Errors map[string]string
}

func (e *ValidationError) Error() string {
	return "Validation failed"
}

func NewValidationError() *ValidationError {
	return &ValidationError{
		Errors: make(map[string]string),
	}
}

func (e *ValidationError) Add(field, message string) {
	e.Errors[field] = message
}

func (e *ValidationError) HasErrors() bool {
	return len(e.Errors) > 0
}

func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{
		Success: false,
		Message: message,
	})
}

func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// 1. Validation error with structured field details (HTTP 422)
	var vErr *ValidationError
	if errors.As(err, &vErr) {
		ValidationErrorResponse(c, vErr.Errors)
		return
	}

	// 2. Explicit typed application error
	var appErr *AppError
	if errors.As(err, &appErr) {
		if appErr.Code >= http.StatusInternalServerError {
			slog.Error("internal server error", "error", appErr.Err, "message", appErr.Message)
			Error(c, appErr.Code, "Internal server error")
			return
		}
		Error(c, appErr.Code, appErr.Error())
		return
	}

	// 3. Unauthenticated (HTTP 401)
	if errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrInvalidCredentials) {
		Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	// 4. Forbidden (HTTP 403)
	if errors.Is(err, ErrForbidden) {
		Error(c, http.StatusForbidden, err.Error())
		return
	}

	// 5. Conflict / Duplicate (HTTP 409)
	if errors.Is(err, ErrDuplicateDomain) || errors.Is(err, ErrConflict) {
		Error(c, http.StatusConflict, err.Error())
		return
	}

	// 6. Not Found (HTTP 404)
	if errors.Is(err, ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
		if errors.Is(err, gorm.ErrRecordNotFound) && !errors.Is(err, ErrNotFound) {
			Error(c, http.StatusNotFound, "resource not found")
			return
		}
		Error(c, http.StatusNotFound, err.Error())
		return
	}

	// 7. Bad request / validation failure (HTTP 400)
	if errors.Is(err, ErrBadRequest) || errors.Is(err, ErrValidation) {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 8. Legacy string comparison for fallback controller / service error strings
	errMsg := err.Error()
	switch errMsg {
	case "invalid credentials", "unauthorized", "invalid email or password", "user account is disabled":
		Error(c, http.StatusUnauthorized, errMsg)
		return
	case "forbidden":
		Error(c, http.StatusForbidden, errMsg)
		return
	case "website not found", "Popup not found", "resource not found":
		Error(c, http.StatusNotFound, errMsg)
		return
	case "website domain already exists":
		Error(c, http.StatusConflict, errMsg)
		return
	case "website_key is required", "website_key query parameter is required", "bad request":
		Error(c, http.StatusBadRequest, errMsg)
		return
	}

	// 9. Unrecognized / unexpected internal error (HTTP 500)
	// Server logs diagnostic information; never expose SQL errors, stack traces, paths, or keys.
	slog.Error("unexpected internal error", "error", err)
	Error(c, http.StatusInternalServerError, "Internal server error")
}

func ValidationErrorResponse(c *gin.Context, errors map[string]string) {
	c.JSON(http.StatusUnprocessableEntity, APIResponse{
		Success: false,
		Message: "Validation failed",
		Errors:  errors,
	})
}

func SuccessWithPagination(c *gin.Context, message string, data interface{}, pagination Pagination) {
	c.JSON(http.StatusOK, APIResponse{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: &pagination,
	})
}
