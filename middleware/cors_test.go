package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupCORSTestRouter(allowedOrigins []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// Public subgroup with CORS
	public := router.Group("/api/v1/public")
	public.Use(PublicCORSMiddleware(allowedOrigins))
	public.GET("/popups", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": []string{}})
	})
	public.OPTIONS("/popups", func(c *gin.Context) {})

	// Admin subgroup without CORS
	admin := router.Group("/api/v1/popups")
	admin.Use(AuthMiddleware())
	admin.GET("/website/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	return router
}

func TestPublicCORS_AllowedOrigin(t *testing.T) {
	router := setupCORSTestRouter([]string{"http://localhost:3000", "http://localhost:3001"})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/public/popups?website_key=test_key", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "http://localhost:3000" {
		t.Errorf("Expected Access-Control-Allow-Origin: http://localhost:3000, got %q", allowOrigin)
	}

	vary := rec.Header().Get("Vary")
	if vary != "Origin" {
		t.Errorf("Expected Vary: Origin, got %q", vary)
	}

	if creds := rec.Header().Get("Access-Control-Allow-Credentials"); creds != "" {
		t.Errorf("Access-Control-Allow-Credentials must not be present, got %q", creds)
	}
}

func TestPublicCORS_DisallowedOrigin(t *testing.T) {
	router := setupCORSTestRouter([]string{"http://localhost:3000"})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/public/popups?website_key=test_key", nil)
	req.Header.Set("Origin", "https://unapproved.example")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "" {
		t.Errorf("Disallowed origin must NOT receive Access-Control-Allow-Origin, got %q", allowOrigin)
	}
}

func TestPublicCORS_NoOriginHeader(t *testing.T) {
	router := setupCORSTestRouter([]string{"http://localhost:3000"})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/public/popups?website_key=test_key", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rec.Code)
	}

	allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "" {
		t.Errorf("Non-browser request without Origin must NOT have Access-Control-Allow-Origin, got %q", allowOrigin)
	}
}

func TestPublicCORS_AllowedPreflight(t *testing.T) {
	router := setupCORSTestRouter([]string{"http://localhost:3000"})

	req, _ := http.NewRequest(http.MethodOptions, "/api/v1/public/popups?website_key=test_key", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("Expected preflight status 204 No Content, got %d", rec.Code)
	}

	allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "http://localhost:3000" {
		t.Errorf("Expected Access-Control-Allow-Origin: http://localhost:3000, got %q", allowOrigin)
	}

	allowMethods := rec.Header().Get("Access-Control-Allow-Methods")
	if allowMethods != "GET, OPTIONS" {
		t.Errorf("Expected Access-Control-Allow-Methods: 'GET, OPTIONS', got %q", allowMethods)
	}

	allowHeaders := rec.Header().Get("Access-Control-Allow-Headers")
	if allowHeaders != "Content-Type, Accept" {
		t.Errorf("Expected Access-Control-Allow-Headers: 'Content-Type, Accept', got %q", allowHeaders)
	}

	if creds := rec.Header().Get("Access-Control-Allow-Credentials"); creds != "" {
		t.Errorf("Access-Control-Allow-Credentials must not be present, got %q", creds)
	}
}

func TestPublicCORS_DisallowedPreflight(t *testing.T) {
	router := setupCORSTestRouter([]string{"http://localhost:3000"})

	req, _ := http.NewRequest(http.MethodOptions, "/api/v1/public/popups?website_key=test_key", nil)
	req.Header.Set("Origin", "https://unapproved.example")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "" {
		t.Errorf("Disallowed preflight must NOT receive Access-Control-Allow-Origin, got %q", allowOrigin)
	}
}

func TestPublicCORS_AdminEndpointIsolation(t *testing.T) {
	router := setupCORSTestRouter([]string{"http://localhost:3000"})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/popups/website/2", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Admin endpoint without token must return 401, got %d", rec.Code)
	}

	allowOrigin := rec.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "" {
		t.Errorf("Admin endpoint must NOT have Access-Control-Allow-Origin, got %q", allowOrigin)
	}
}
