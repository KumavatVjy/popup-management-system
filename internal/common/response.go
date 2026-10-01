package common

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
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
	if vErr, ok := err.(*ValidationError); ok {
		ValidationErrorResponse(c, vErr.Errors)
		return
	}

	if errors.Is(err, ErrDuplicateDomain) {
		Error(c, http.StatusConflict, err.Error())
		return
	}

	if errors.Is(err, ErrNotFound) {
		Error(c, http.StatusNotFound, err.Error())
		return
	}

	// Legacy string comparison for fallback or custom controller errors
	errMsg := err.Error()
	if errMsg == "invalid credentials" || errMsg == "unauthorized" {
		Error(c, http.StatusUnauthorized, errMsg)
		return
	}
	if errMsg == "website not found" || errMsg == "Popup not found" {
		Error(c, http.StatusNotFound, errMsg)
		return
	}
	if errMsg == "website domain already exists" {
		Error(c, http.StatusConflict, errMsg)
		return
	}

	Error(c, http.StatusBadRequest, err.Error())
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
