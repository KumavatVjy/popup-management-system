package routes

import (
	"github.com/gin-gonic/gin"
	"popup-manager-api/internal/popups"
	"popup-manager-api/internal/users"
	"popup-manager-api/internal/websites"
)

type Controllers struct {
	User    *users.UserController
	Website *websites.WebsiteController
	Popup   *popups.PopupController
}

func SetupRoutes(router *gin.Engine, controllers *Controllers) {

	api := router.Group("/api")

	v1 := api.Group("/v1")

	users.RegisterRoutes(v1, controllers.User)

	websites.RegisterRoutes(v1, controllers.Website)

	popups.RegisterRoutes(v1, controllers.Popup)

	router.GET("/", func(c *gin.Context) {

		c.JSON(200, gin.H{
			"success": true,
			"message": "Popup Management API",
			"version": "v1",
		})

	})
}
