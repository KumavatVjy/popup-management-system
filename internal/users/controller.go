package users

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

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
		slog.Warn("user login failed", "email", request.Email, "reason", "invalid credentials")
		common.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	// Generate JWT
	token, err := utils.GenerateJWT(
		user.ID,
		user.Email,
		user.Role,
	)

	if err != nil {
		slog.Error("failed to generate jwt token", "error", err)
		common.Error(
			c,
			http.StatusInternalServerError,
			"Failed to generate authentication token",
		)
		return
	}

	// Return successful login response
	slog.Info("user login successful", "user_id", user.ID)
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
