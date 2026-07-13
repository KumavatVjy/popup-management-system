package users

import (
	"popup-manager-api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, controller *UserController) {

	router.POST("/login", controller.Login)

	// Protected routes
	// router.GET("/profile", middleware.AuthMiddleware(), controller.Profile)


	auth := router.Group("/")

	auth.Use(middleware.AuthMiddleware())

	auth.GET("/profile", controller.Profile)

}