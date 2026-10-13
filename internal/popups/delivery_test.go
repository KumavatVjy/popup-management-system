package popups

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"popup-manager-api/internal/common"
	"popup-manager-api/internal/websites"
)

type mockWebsiteRepo struct {
	websitesByKey map[string]*websites.Website
}

func (m *mockWebsiteRepo) Create(website *websites.Website) error { return nil }
func (m *mockWebsiteRepo) GetAll(params common.QueryParams) ([]websites.Website, int64, error) {
	return nil, 0, nil
}
func (m *mockWebsiteRepo) GetByID(id uint) (*websites.Website, error) {
	for _, w := range m.websitesByKey {
		if w.ID == id {
			return w, nil
		}
	}
	return nil, common.ErrNotFound
}
func (m *mockWebsiteRepo) GetByDomain(domain string) (*websites.Website, error) {
	return nil, common.ErrNotFound
}
func (m *mockWebsiteRepo) GetByKey(key string) (*websites.Website, error) {
	if w, ok := m.websitesByKey[key]; ok {
		return w, nil
	}
	return nil, common.ErrNotFound
}
func (m *mockWebsiteRepo) Update(website *websites.Website) error { return nil }
func (m *mockWebsiteRepo) Delete(id uint) error                   { return nil }

type mockPopupRepo struct {
	popupsByWebsiteID map[uint][]Popup
}

func (m *mockPopupRepo) Create(popup *Popup) error { return nil }
func (m *mockPopupRepo) GetAll(params common.QueryParams) ([]Popup, int64, error) {
	return nil, 0, nil
}
func (m *mockPopupRepo) GetByID(id uint) (*Popup, error) { return nil, common.ErrNotFound }
func (m *mockPopupRepo) GetByWebsiteID(websiteID uint, params common.QueryParams) ([]Popup, int64, error) {
	return nil, 0, nil
}
func (m *mockPopupRepo) GetByWebsite(websiteID uint) ([]Popup, error) {
	return m.popupsByWebsiteID[websiteID], nil
}
func (m *mockPopupRepo) GetEligiblePublicByWebsite(websiteID uint, now time.Time) ([]Popup, error) {
	return m.popupsByWebsiteID[websiteID], nil
}
func (m *mockPopupRepo) Update(popup *Popup) error { return nil }
func (m *mockPopupRepo) Delete(id uint) error      { return nil }

func setupDeliveryTestRouter(wRepo websites.WebsiteRepository, pRepo PopupRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	service := NewPopupService(pRepo, wRepo)
	ctrl := NewPopupController(service)

	public := router.Group("/api/v1/public")
	public.GET("/popups", ctrl.GetPublicPopups)

	return router
}

func TestInactiveWebsiteDeliveryRules(t *testing.T) {
	activeSite := &websites.Website{
		WebsiteName: "Active Site",
		Domain:      "active.example.com",
		WebsiteKey:  "wg_live_active_key_111",
		Status:      true,
	}
	activeSite.ID = 1

	inactiveSite := &websites.Website{
		WebsiteName: "Inactive Site",
		Domain:      "inactive.example.com",
		WebsiteKey:  "wg_live_inactive_key_222",
		Status:      false,
	}
	inactiveSite.ID = 2

	activeSiteTwo := &websites.Website{
		WebsiteName: "Other Active Site",
		Domain:      "other.example.com",
		WebsiteKey:  "wg_live_other_key_333",
		Status:      true,
	}
	activeSiteTwo.ID = 3

	wRepo := &mockWebsiteRepo{
		websitesByKey: map[string]*websites.Website{
			"wg_live_active_key_111":   activeSite,
			"wg_live_inactive_key_222": inactiveSite,
			"wg_live_other_key_333":    activeSiteTwo,
		},
	}

	pRepo := &mockPopupRepo{
		popupsByWebsiteID: map[uint][]Popup{
			1: {
				{
					Title:     "Active Site Popup",
					Content:   "<p>Active Popup Content</p>",
					Position:  PositionCenter,
					Status:    true,
					WebsiteID: 1,
				},
			},
			2: {
				{
					Title:     "Inactive Site Popup",
					Content:   "<p>Should Not Be Delivered</p>",
					Position:  PositionCenter,
					Status:    true,
					WebsiteID: 2,
				},
			},
			3: {
				{
					Title:     "Other Site Popup",
					Content:   "<p>Other Site Content</p>",
					Position:  PositionTopBanner,
					Status:    true,
					WebsiteID: 3,
				},
			},
		},
	}

	router := setupDeliveryTestRouter(wRepo, pRepo)

	// 1. Active website receives its eligible popups (200 OK)
	t.Run("Active website receives eligible popups", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/public/popups?website_key=wg_live_active_key_111", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp common.APIResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to parse response JSON: %v", err)
		}
		if !resp.Success {
			t.Errorf("Expected success: true, got false")
		}

		popupsSlice, ok := resp.Data.([]interface{})
		if !ok || len(popupsSlice) != 1 {
			t.Fatalf("Expected 1 popup in data array, got: %v", resp.Data)
		}
		popupMap := popupsSlice[0].(map[string]interface{})
		if popupMap["title"] != "Active Site Popup" {
			t.Errorf("Expected title 'Active Site Popup', got %v", popupMap["title"])
		}
	})

	// 2. Inactive website cannot receive popups (404 Not Found)
	t.Run("Inactive website returns 404 and does not deliver popups", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/public/popups?website_key=wg_live_inactive_key_222", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("Expected 404 Not Found for inactive website, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp common.APIResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp.Success {
			t.Errorf("Expected success: false for inactive site")
		}
		if resp.Message != "resource not found" {
			t.Errorf("Expected message 'resource not found', got %q", resp.Message)
		}
	})

	// 3. Unknown key returns 404 (same error behavior as inactive website)
	t.Run("Unknown website key returns 404", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/public/popups?website_key=wg_live_unknown_nonexistent_key", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("Expected 404 Not Found for unknown key, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var resp common.APIResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp.Success {
			t.Errorf("Expected success: false for unknown key")
		}
		if resp.Message != "resource not found" {
			t.Errorf("Expected message 'resource not found', got %q", resp.Message)
		}
	})

	// 4. Cross-site isolation: A key belonging to another active website cannot retrieve first site's popups
	t.Run("Cross-site isolation ensures only site's own popups are returned", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/public/popups?website_key=wg_live_other_key_333", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d", rec.Code)
		}

		var resp common.APIResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		popupsSlice := resp.Data.([]interface{})
		if len(popupsSlice) != 1 {
			t.Fatalf("Expected 1 popup, got %d", len(popupsSlice))
		}
		popupMap := popupsSlice[0].(map[string]interface{})
		if popupMap["title"] != "Other Site Popup" {
			t.Errorf("Expected title 'Other Site Popup', got %v", popupMap["title"])
		}
	})

	// 5. Missing website key returns 400
	t.Run("Missing website_key query param returns 400", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/public/popups", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("Expected 400 Bad Request for missing key, got %d", rec.Code)
		}
	})
}
