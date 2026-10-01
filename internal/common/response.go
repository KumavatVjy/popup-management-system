package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
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

func SuccessWithPagination(c *gin.Context, message string, data interface{}, pagination Pagination) {
	c.JSON(http.StatusOK, APIResponse{
		Success:    true,
		Message:    message,
		Data:       data,
		Pagination: &pagination,
	})
}
