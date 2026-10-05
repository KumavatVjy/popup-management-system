package common

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// InitLogger sets up a global structured logger using slog.
func InitLogger(env string) {
	var handler slog.Handler

	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// RequestLogger returns a gin middleware that logs HTTP requests using slog.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Milliseconds()
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		var userID any
		if id, exists := c.Get("user_id"); exists {
			userID = id
		}

		args := []any{
			"method", method,
			"path", path,
			"status", status,
			"duration_ms", duration,
			"client_ip", clientIP,
		}

		if userID != nil {
			args = append(args, "user_id", userID)
		}

		if status >= 500 {
			slog.Error("http request", args...)
		} else if status >= 400 {
			slog.Warn("http request", args...)
		} else {
			slog.Info("http request", args...)
		}
	}
}
