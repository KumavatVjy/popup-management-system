package middleware

import (
	"net/http"
	"sync"
	"time"

	"popup-manager-api/internal/common"

	"github.com/gin-gonic/gin"
)

type ipRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

type visitor struct {
	count     int
	lastReset time.Time
}

// NewRateLimiter creates an in-memory IP-based rate limiting middleware.
// It limits each client IP to maxRequests within the specified time window.
func NewRateLimiter(maxRequests int, window time.Duration) gin.HandlerFunc {
	limiter := &ipRateLimiter{
		visitors: make(map[string]*visitor),
		limit:    maxRequests,
		window:   window,
	}

	// Periodically clean up stale visitor records to prevent memory growth
	go limiter.cleanupLoop(5 * time.Minute)

	return func(c *gin.Context) {
		// Do not rate limit preflight OPTIONS requests
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		if clientIP == "" {
			clientIP = "unknown"
		}

		if !limiter.allow(clientIP) {
			common.Error(c, http.StatusTooManyRequests, "Too many requests. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}

func (rl *ipRateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]
	if !exists || now.Sub(v.lastReset) >= rl.window {
		rl.visitors[ip] = &visitor{
			count:     1,
			lastReset: now,
		}
		return true
	}

	if v.count < rl.limit {
		v.count++
		return true
	}

	return false
}

func (rl *ipRateLimiter) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			if now.Sub(v.lastReset) > rl.window*2 {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}
