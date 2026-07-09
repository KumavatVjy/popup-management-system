package routes

import (
	"popup-manager-api/internal/users"

	"github.com/gin-gonic/gin"
)

type Controllers struct {
	User *users.UserController
}

func SetupRoutes(router *gin.Engine, controllers *Controllers) {

	api := router.Group("/api")

	v1 := api.Group("/v1")

	users.RegisterRoutes(v1, controllers.User)

	router.GET("/", func(c *gin.Context) {

		c.JSON(200, gin.H{
			"success": true,
			"message": "Popup Management API",
			"version": "v1",
		})

	})
}