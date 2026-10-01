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

	// Bind and validate request
	if err := c.ShouldBindJSON(&request); err != nil {
		common.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Authenticate user
	user, err := uc.Service.Login(request.Email, request.Password)

	if err != nil {
		common.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	// Get JWT expiry configuration
	expireHours, err := strconv.Atoi(
		config.GetEnv("JWT_EXPIRE_HOURS"),
	)

	if err != nil {
		common.Error(
			c,
			http.StatusInternalServerError,
			"Invalid JWT configuration",
		)
		return
	}

	// Generate JWT
	token, err := utils.GenerateJWT(
		user.ID,
		user.Email,
		user.Role,
		config.GetEnv("JWT_SECRET"),
		expireHours,
	)

	if err != nil {
		common.Error(
			c,
			http.StatusInternalServerError,
			"Failed to generate authentication token",
		)
		return
	}

	// Return successful login response
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

func (uc *UserController) Profile(c *gin.Context) {

	common.Success(c, "Profile fetched successfully", gin.H{
		"user_id": c.GetUint("user_id"),
		"email":   c.GetString("email"),
		"role":    c.GetString("role"),
	})
}
