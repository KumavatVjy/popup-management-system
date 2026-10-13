package websites

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"popup-manager-api/config"
	"popup-manager-api/internal/common"
	"popup-manager-api/utils"
)

type mockWebsiteRepoForIntegration struct {
	websites map[uint]*Website
}

func (m *mockWebsiteRepoForIntegration) Create(website *Website) error {
	m.websites[website.ID] = website
	return nil
}

func (m *mockWebsiteRepoForIntegration) GetAll(params common.QueryParams) ([]Website, int64, error) {
	var list []Website
	for _, w := range m.websites {
		list = append(list, *w)
	}
	return list, int64(len(list)), nil
}

func (m *mockWebsiteRepoForIntegration) GetByID(id uint) (*Website, error) {
	if w, ok := m.websites[id]; ok {
		return w, nil
	}
	return nil, common.ErrNotFound
}

func (m *mockWebsiteRepoForIntegration) GetByDomain(domain string) (*Website, error) {
	for _, w := range m.websites {
		if w.Domain == domain {
			return w, nil
		}
	}
	return nil, common.ErrNotFound
}

func (m *mockWebsiteRepoForIntegration) GetByKey(key string) (*Website, error) {
	for _, w := range m.websites {
		if w.WebsiteKey == key {
			return w, nil
		}
	}
	return nil, common.ErrNotFound
}

func (m *mockWebsiteRepoForIntegration) Update(website *Website) error {
	m.websites[website.ID] = website
	return nil
}

func (m *mockWebsiteRepoForIntegration) Delete(id uint) error {
	if _, ok := m.websites[id]; !ok {
		return common.ErrNotFound
	}
	delete(m.websites, id)
	return nil
}

func setupWebsiteTestRouter(repo WebsiteRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	service := NewWebsiteService(repo)
	controller := NewWebsiteController(service)

	api := router.Group("/api/v1")
	RegisterRoutes(api, controller)

	return router
}

func TestWebsiteIntegrationDetailsEndpoint(t *testing.T) {
	// Setup test config for JWT validation
	config.AppConfig = &config.Config{
		JWTSecret:      "test-secret-key-32-chars-long-123",
		JWTExpireHours: 24,
	}

	validToken, err := utils.GenerateJWT(1, "admin@example.com", "Super Admin")
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	sampleWebsite := &Website{
		WebsiteName: "Acme Store",
		Domain:      "acme.store",
		Platform:    PlatformReact,
		WebsiteKey:  "wg_live_acme_secret_integration_key_999",
		Status:      true,
		CreatedBy:   1,
	}
	sampleWebsite.ID = 42

	repo := &mockWebsiteRepoForIntegration{
		websites: map[uint]*Website{
			42: sampleWebsite,
		},
	}

	router := setupWebsiteTestRouter(repo)

	// 1. Unauthenticated request returns 401 Unauthorized
	t.Run("Unauthenticated request returns 401", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/websites/42/integration", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("Expected 401 Unauthorized, got %d. Body: %s", rec.Code, rec.Body.String())
		}
	})

	// 2. Properly authenticated request returns integration details (200 OK)
	t.Run("Authenticated request returns integration details", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/websites/42/integration", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp struct {
			Success bool                       `json:"success"`
			Message string                     `json:"message"`
			Data    WebsiteIntegrationResponse `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		if !resp.Success {
			t.Errorf("Expected success: true, got false")
		}
		if resp.Data.ID != 42 {
			t.Errorf("Expected website ID 42, got %d", resp.Data.ID)
		}
		if resp.Data.Domain != "acme.store" {
			t.Errorf("Expected domain 'acme.store', got %q", resp.Data.Domain)
		}
		if resp.Data.WebsiteKey != "wg_live_acme_secret_integration_key_999" {
			t.Errorf("Expected matching website_key, got %q", resp.Data.WebsiteKey)
		}

		// Ensure no sensitive or internal user/database fields are present in integration response
		var rawMap map[string]interface{}
		_ = json.Unmarshal(rec.Body.Bytes(), &rawMap)
		dataMap := rawMap["data"].(map[string]interface{})
		for _, forbiddenKey := range []string{"created_by", "created_at", "updated_at", "deleted_at", "status", "platform"} {
			if _, exists := dataMap[forbiddenKey]; exists {
				t.Errorf("Forbidden metadata field %q leaked in integration details!", forbiddenKey)
			}
		}
	})

	// 3. Repeated requests return the same key without re-generation
	t.Run("Repeated requests return identical integration key", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/websites/42/integration", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		var resp struct {
			Data WebsiteIntegrationResponse `json:"data"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)

		if resp.Data.WebsiteKey != "wg_live_acme_secret_integration_key_999" {
			t.Errorf("Website key changed unexpectedly: %q", resp.Data.WebsiteKey)
		}
	})

	// 4. Unknown website ID returns 404
	t.Run("Unknown website ID returns 404", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/websites/9999/integration", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("Expected 404 Not Found, got %d", rec.Code)
		}
	})

	// 5. Invalid website ID syntax returns 400
	t.Run("Invalid ID syntax returns 400", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/websites/invalid-id/integration", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Expected 400 Bad Request, got %d", rec.Code)
		}
	})

	// 6. Verification that website key is absent from standard website detail & list DTOs
	t.Run("Website key is absent from standard GET /websites/:id and GET /websites", func(t *testing.T) {
		// Detail endpoint
		reqDetail, _ := http.NewRequest(http.MethodGet, "/api/v1/websites/42", nil)
		reqDetail.Header.Set("Authorization", "Bearer "+validToken)
		recDetail := httptest.NewRecorder()
		router.ServeHTTP(recDetail, reqDetail)

		if recDetail.Code != http.StatusOK {
			t.Fatalf("Expected 200 for detail endpoint, got %d", recDetail.Code)
		}

		var rawDetail map[string]interface{}
		_ = json.Unmarshal(recDetail.Body.Bytes(), &rawDetail)
		detailData := rawDetail["data"].(map[string]interface{})
		if _, exists := detailData["website_key"]; exists {
			t.Errorf("LEAK: website_key must NOT be present in standard WebsiteResponse detail DTO!")
		}

		// List endpoint
		reqList, _ := http.NewRequest(http.MethodGet, "/api/v1/websites", nil)
		reqList.Header.Set("Authorization", "Bearer "+validToken)
		recList := httptest.NewRecorder()
		router.ServeHTTP(recList, reqList)

		if recList.Code != http.StatusOK {
			t.Fatalf("Expected 200 for list endpoint, got %d", recList.Code)
		}

		var rawList map[string]interface{}
		_ = json.Unmarshal(recList.Body.Bytes(), &rawList)
		listData := rawList["data"].([]interface{})
		firstItem := listData[0].(map[string]interface{})
		if _, exists := firstItem["website_key"]; exists {
			t.Errorf("LEAK: website_key must NOT be present in standard WebsiteResponse list DTO!")
		}
	})
}
