package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiter_AllowsWithinLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(NewRateLimiter(3, 1*time.Minute))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 on request %d, got %d", i+1, rec.Code)
		}
	}
}

func TestRateLimiter_BlocksExceededLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(NewRateLimiter(2, 1*time.Minute))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Request 1: OK
	req1, _ := http.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec1.Code)
	}

	// Request 2: OK
	req2, _ := http.NewRequest(http.MethodGet, "/test", nil)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec2.Code)
	}

	// Request 3: 429 Too Many Requests
	req3, _ := http.NewRequest(http.MethodGet, "/test", nil)
	rec3 := httptest.NewRecorder()
	router.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", rec3.Code)
	}
}

func TestRateLimiter_AllowsOptionsPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(NewRateLimiter(1, 1*time.Minute))
	router.OPTIONS("/test", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	// Multiple OPTIONS requests should not be blocked or count against limit
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status 204 on OPTIONS request %d, got %d", i+1, rec.Code)
		}
	}
}
