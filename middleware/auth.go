package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"popup-manager-api/internal/common"
	"popup-manager-api/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			slog.Warn("authentication failed", "reason", "missing authorization header")
			common.Error(c, http.StatusUnauthorized, "Authorization header is required")
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			slog.Warn("authentication failed", "reason", "invalid authorization header")
			common.Error(c, http.StatusUnauthorized, "Invalid authorization format")
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := utils.ValidateJWT(
			token,
		)

		if err != nil {
			slog.Warn("authentication failed", "reason", "invalid token")
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
