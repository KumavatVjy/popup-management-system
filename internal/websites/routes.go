package websites

import (
	"popup-manager-api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, controller *WebsiteController) {

	websites := router.Group("/websites")

	websites.Use(middleware.AdminCORSMiddleware(), middleware.AuthMiddleware())
	websites.OPTIONS("", func(c *gin.Context) {})
	websites.OPTIONS("/*path", func(c *gin.Context) {})

	websites.POST("", controller.Create)

	websites.GET("", controller.List)

	websites.GET("/:id", controller.GetByID)
	websites.GET("/:id/integration", controller.GetIntegrationDetails)

	websites.PUT("/:id", controller.Update)

	websites.DELETE("/:id", controller.Delete)

}
