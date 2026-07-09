package users

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, controller *UserController) {

	router.POST("/login", controller.Login)

}