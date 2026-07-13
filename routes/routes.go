package routes

import (
	"popup-manager-api/internal/users"
    "popup-manager-api/internal/websites"
	"github.com/gin-gonic/gin"
)

type Controllers struct {
	User *users.UserController
	Website *websites.WebsiteController
}

func SetupRoutes(router *gin.Engine, controllers *Controllers) {

	api := router.Group("/api")

	v1 := api.Group("/v1")

	users.RegisterRoutes(v1, controllers.User)

	websites.RegisterRoutes(v1, controllers.Website)

	router.GET("/", func(c *gin.Context) {

		c.JSON(200, gin.H{
			"success": true,
			"message": "Popup Management API",
			"version": "v1",
		})

	})
}