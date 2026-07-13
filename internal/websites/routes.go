package websites

import (
	"popup-manager-api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, controller *WebsiteController) {

	websites := router.Group("/websites")

	websites.Use(middleware.AuthMiddleware())

	websites.POST("", controller.Create)

	websites.GET("", controller.List)

}