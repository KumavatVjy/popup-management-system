package users

import (
	"time"

	"popup-manager-api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, controller *UserController) {

	admin := router.Group("/")
	admin.Use(middleware.AdminCORSMiddleware())
	admin.POST("/login", middleware.NewRateLimiter(10, time.Minute), controller.Login)
	admin.OPTIONS("/login", func(c *gin.Context) {})

	// Protected routes
	// router.GET("/profile", middleware.AuthMiddleware(), controller.Profile)

	auth := admin.Group("/")

	auth.Use(middleware.AuthMiddleware())

	auth.GET("/profile", controller.Profile)
	auth.OPTIONS("/profile", func(c *gin.Context) {})

}
