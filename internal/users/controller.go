package users

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"popup-manager-api/config"
	"popup-manager-api/internal/common"
	"popup-manager-api/utils"
)

type UserController struct {
	Service *UserService
}

func NewUserController(service *UserService) *UserController {
	return &UserController{
		Service: service,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (uc *UserController) Login(c *gin.Context) {

	var request LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {

		common.Error(c, http.StatusBadRequest, "Invalid request body")

		return
	}

	user, err := uc.Service.Login(request.Email, request.Password)

	expireHours, err := strconv.Atoi(config.GetEnv("JWT_EXPIRE_HOURS"))

	if err != nil {
		expireHours = 24
	}

	token, err := utils.GenerateJWT(
		user.ID,
		user.Email,
		user.Role,
		config.GetEnv("JWT_SECRET"),
		expireHours,
	)

	if err != nil {
		common.Error(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	if err != nil {

		common.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	common.Success(c, "Login successful", gin.H{
		"token": token,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}