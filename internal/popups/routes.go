package popups

import (
	"popup-manager-api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, controller *PopupController) {
	// Public delivery route (unauthenticated, with CORS middleware)
	public := router.Group("/public")
	public.Use(middleware.PublicCORSMiddleware())
	public.GET("/popups", controller.GetPublicPopups)
	public.OPTIONS("/popups", func(c *gin.Context) {})

	// Admin routes (requires authentication)
	popups := router.Group("/popups")
	popups.Use(middleware.AuthMiddleware())

	popups.POST("", controller.Create)
	popups.GET("", controller.List)
	popups.GET("/:id", controller.GetByID)
	popups.GET("/website/:website_id", controller.GetByWebsiteID)
	popups.PUT("/:id", controller.Update)
	popups.DELETE("/:id", controller.Delete)
}
