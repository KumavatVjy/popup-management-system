package routes

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"popup-manager-api/internal/popups"
	"popup-manager-api/internal/users"
	"popup-manager-api/internal/websites"
)

type Controllers struct {
	User    *users.UserController
	Website *websites.WebsiteController
	Popup   *popups.PopupController
}

func SetupRoutes(router *gin.Engine, controllers *Controllers) {
	// Serve the client SDK directly so external websites only need a single script tag
	router.GET("/sdk/popup-manager.min.js", func(c *gin.Context) {
		c.Header("Content-Type", "application/javascript")
		c.Header("Access-Control-Allow-Origin", "*")
		const bundlePath = "popup-manager-sdk/dist/popup-manager.min.js"
		candidatePaths := make([]string, 0, 4)
		if configuredPath := strings.TrimSpace(os.Getenv("SDK_DIST_PATH")); configuredPath != "" {
			candidatePaths = append(candidatePaths, configuredPath)
		}
		candidatePaths = append(candidatePaths,
			filepath.Join("..", bundlePath),
			filepath.Join(".", bundlePath),
		)
		if executablePath, err := os.Executable(); err == nil {
			candidatePaths = append(candidatePaths,
				filepath.Join(filepath.Dir(executablePath), "..", bundlePath),
			)
		}
		for _, p := range candidatePaths {
			if _, err := os.Stat(p); err == nil {
				c.File(p)
				return
			}
		}
		c.String(http.StatusNotFound, "SDK distribution file not found")
	})

	api := router.Group("/api")

	v1 := api.Group("/v1")

	users.RegisterRoutes(v1, controllers.User)

	websites.RegisterRoutes(v1, controllers.Website)

	popups.RegisterRoutes(v1, controllers.Popup)

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
			"message": "Popup Management API",
			"version": "v1",
		})
	})
}
