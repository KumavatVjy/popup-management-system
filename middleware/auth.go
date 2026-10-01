package middleware

import (
	"net/http"
	"strings"

	"popup-manager-api/config"
	"popup-manager-api/internal/common"
	"popup-manager-api/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			common.Error(c, http.StatusUnauthorized, "Authorization header is required")
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			common.Error(c, http.StatusUnauthorized, "Invalid authorization format")
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := utils.ValidateJWT(
			token,
			config.GetEnv("JWT_SECRET"),
		)

		if err != nil {
			common.Error(c, http.StatusUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}
