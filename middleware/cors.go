package middleware

import (
	"net/http"
	"strings"

	"popup-manager-api/config"

	"github.com/gin-gonic/gin"
)

// PublicCORSMiddleware returns a Gin middleware that enforces an explicit CORS policy
// for public delivery endpoints. If origins slice is provided, it is used;
// otherwise, it falls back to config.AppConfig.PublicCORSAllowedOrigins.
func PublicCORSMiddleware(customOrigins ...[]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var origins []string
		if len(customOrigins) > 0 && customOrigins[0] != nil {
			origins = customOrigins[0]
		} else if config.AppConfig != nil {
			origins = config.AppConfig.PublicCORSAllowedOrigins
		}

		origin := strings.TrimSpace(c.GetHeader("Origin"))

		// If no Origin header is present (non-browser or server-to-server request),
		// proceed normally without adding CORS headers.
		if origin == "" {
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
			c.Next()
			return
		}

		// Check whether the Origin is explicitly permitted
		isAllowed := false
		for _, o := range origins {
			if strings.TrimSpace(o) == origin {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")

			// Handle preflight OPTIONS request
			if c.Request.Method == http.MethodOptions {
				c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Content-Type, Accept")
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
		} else {
			// Disallowed origin: do not grant cross-origin permission
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
		}

		c.Next()
	}
}
