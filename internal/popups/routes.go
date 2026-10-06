package popups

import (
	"popup-manager-api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, controller *PopupController) {
	// Public delivery route (unauthenticated)
	public := router.Group("/public")
	public.GET("/popups", controller.GetPublicPopups)

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
