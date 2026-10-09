package popups

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"popup-manager-api/config"
	"popup-manager-api/internal/websites"
)

func TestPublicCORS_ScenariosAtoG(t *testing.T) {
	db, router, _, _, _ := setupTestEnvironment(t)
	if db == nil {
		return
	}

	// Ensure config has explicit allowed origins for testing
	config.AppConfig.PublicCORSAllowedOrigins = []string{"http://localhost:3000", "http://localhost:3001"}

	// Create test website and popups for testing
	nowNano := time.Now().UnixNano()
	siteKey := fmt.Sprintf("wg_live_cors_%d", nowNano)

	site := websites.Website{
		WebsiteName: fmt.Sprintf("CORS Test Site %d", nowNano),
		Domain:      fmt.Sprintf("cors-%d.com", nowNano),
		Platform:    websites.PlatformHTML,
		WebsiteKey:  siteKey,
		Status:      true,
		CreatedBy:   1,
	}
	if err := db.Create(&site).Error; err != nil {
		t.Fatalf("Failed to create test website: %v", err)
	}

	defer func() {
		db.Unscoped().Where("website_id = ?", site.ID).Delete(&Popup{})
		db.Unscoped().Where("id = ?", site.ID).Delete(&websites.Website{})
	}()

	now := time.Now()
	past := now.Add(-2 * time.Hour)
	future := now.Add(2 * time.Hour)

	// Popups for eligibility regression (Test G)
	// 1. Eligible: Active, no schedule
	pEligible := Popup{
		WebsiteID: site.ID,
		Title:     "Eligible Popup",
		Content:   "Eligible Content",
		Position:  PositionCenter,
		Status:    true,
		CreatedBy: 1,
	}
	db.Create(&pEligible)

	// 2. Inactive: status = false
	pInactive := Popup{
		WebsiteID: site.ID,
		Title:     "Inactive Popup",
		Content:   "Inactive Content",
		Position:  PositionCenter,
		Status:    false,
		CreatedBy: 1,
	}
	db.Create(&pInactive)
	db.Model(&pInactive).Update("status", false)

	// 3. Future scheduled
	pFuture := Popup{
		WebsiteID: site.ID,
		Title:     "Future Popup",
		Content:   "Future Content",
		Position:  PositionCenter,
		Status:    true,
		StartTime: &future,
		CreatedBy: 1,
	}
	db.Create(&pFuture)

	// 4. Expired past
	pExpired := Popup{
		WebsiteID: site.ID,
		Title:     "Expired Popup",
		Content:   "Expired Content",
		Position:  PositionCenter,
		Status:    true,
		EndTime:   &past,
		CreatedBy: 1,
	}
	db.Create(&pExpired)

	// ==========================================
	// Test A — Allowed origin
	// ==========================================
	reqA, _ := http.NewRequest(http.MethodGet, "/api/v1/public/popups?website_key="+siteKey, nil)
	reqA.Header.Set("Origin", "http://localhost:3000")
	recA := httptest.NewRecorder()
	router.ServeHTTP(recA, reqA)

	if recA.Code != http.StatusOK {
		t.Errorf("Test A failed: Expected HTTP 200, got %d", recA.Code)
	}
	allowOriginA := recA.Header().Get("Access-Control-Allow-Origin")
	if allowOriginA != "http://localhost:3000" {
		t.Errorf("Test A failed: Expected Access-Control-Allow-Origin 'http://localhost:3000', got %q", allowOriginA)
	}
	varyA := recA.Header().Get("Vary")
	if varyA != "Origin" {
		t.Errorf("Test A failed: Expected Vary 'Origin', got %q", varyA)
	}
	if recA.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Errorf("Test A failed: Credentials must NOT be allowed")
	}
	t.Logf("PASS: Test A — Allowed origin http://localhost:3000 received correct CORS headers")

	// ==========================================
	// Test B — Disallowed origin
	// ==========================================
	reqB, _ := http.NewRequest(http.MethodGet, "/api/v1/public/popups?website_key="+siteKey, nil)
	reqB.Header.Set("Origin", "https://unapproved.example")
	recB := httptest.NewRecorder()
	router.ServeHTTP(recB, reqB)

	if recB.Code != http.StatusOK {
		t.Errorf("Test B failed: Expected HTTP 200, got %d", recB.Code)
	}
	allowOriginB := recB.Header().Get("Access-Control-Allow-Origin")
	if allowOriginB != "" {
		t.Errorf("Test B failed: Disallowed origin must NOT receive Access-Control-Allow-Origin, got %q", allowOriginB)
	}
	t.Logf("PASS: Test B — Disallowed origin https://unapproved.example received no CORS headers")

	// ==========================================
	// Test C — No Origin header
	// ==========================================
	reqC, _ := http.NewRequest(http.MethodGet, "/api/v1/public/popups?website_key="+siteKey, nil)
	recC := httptest.NewRecorder()
	router.ServeHTTP(recC, reqC)

	if recC.Code != http.StatusOK {
		t.Errorf("Test C failed: Expected HTTP 200, got %d", recC.Code)
	}
	allowOriginC := recC.Header().Get("Access-Control-Allow-Origin")
	if allowOriginC != "" {
		t.Errorf("Test C failed: Non-browser request without Origin must NOT have Access-Control-Allow-Origin, got %q", allowOriginC)
	}
	t.Logf("PASS: Test C — Request without Origin header processed normally without CORS headers")

	// ==========================================
	// Test D — Allowed preflight
	// ==========================================
	reqD, _ := http.NewRequest(http.MethodOptions, "/api/v1/public/popups?website_key="+siteKey, nil)
	reqD.Header.Set("Origin", "http://localhost:3000")
	reqD.Header.Set("Access-Control-Request-Method", "GET")
	recD := httptest.NewRecorder()
	router.ServeHTTP(recD, reqD)

	if recD.Code != http.StatusNoContent {
		t.Errorf("Test D failed: Expected HTTP 204 No Content for preflight, got %d", recD.Code)
	}
	allowOriginD := recD.Header().Get("Access-Control-Allow-Origin")
	if allowOriginD != "http://localhost:3000" {
		t.Errorf("Test D failed: Expected Access-Control-Allow-Origin 'http://localhost:3000', got %q", allowOriginD)
	}
	methodsD := recD.Header().Get("Access-Control-Allow-Methods")
	if methodsD != "GET, OPTIONS" {
		t.Errorf("Test D failed: Expected Access-Control-Allow-Methods 'GET, OPTIONS', got %q", methodsD)
	}
	headersD := recD.Header().Get("Access-Control-Allow-Headers")
	if headersD != "Content-Type, Accept" {
		t.Errorf("Test D failed: Expected Access-Control-Allow-Headers 'Content-Type, Accept', got %q", headersD)
	}
	if recD.Header().Get("Access-Control-Allow-Credentials") != "" {
		t.Errorf("Test D failed: Credentials must NOT be allowed")
	}
	t.Logf("PASS: Test D — Allowed preflight received 204 No Content and appropriate CORS headers")

	// ==========================================
	// Test E — Disallowed preflight
	// ==========================================
	reqE, _ := http.NewRequest(http.MethodOptions, "/api/v1/public/popups?website_key="+siteKey, nil)
	reqE.Header.Set("Origin", "https://unapproved.example")
	reqE.Header.Set("Access-Control-Request-Method", "GET")
	recE := httptest.NewRecorder()
	router.ServeHTTP(recE, reqE)

	allowOriginE := recE.Header().Get("Access-Control-Allow-Origin")
	if allowOriginE != "" {
		t.Errorf("Test E failed: Disallowed preflight must NOT receive Access-Control-Allow-Origin, got %q", allowOriginE)
	}
	t.Logf("PASS: Test E — Disallowed preflight granted no cross-origin permission")

	// ==========================================
	// Test F — Admin endpoint isolation
	// ==========================================
	reqF, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/popups/website/%d", site.ID), nil)
	reqF.Header.Set("Origin", "http://localhost:3000")
	recF := httptest.NewRecorder()
	router.ServeHTTP(recF, reqF)

	if recF.Code != http.StatusUnauthorized {
		t.Errorf("Test F failed: Expected HTTP 401 for admin endpoint without token, got %d", recF.Code)
	}
	allowOriginF := recF.Header().Get("Access-Control-Allow-Origin")
	if allowOriginF != "" {
		t.Errorf("Test F failed: Admin endpoint must NOT receive Access-Control-Allow-Origin, got %q", allowOriginF)
	}
	t.Logf("PASS: Test F — Admin endpoint remains protected (401) with zero CORS headers")

	// ==========================================
	// Test G — Public eligibility regression
	// ==========================================
	var respA map[string]interface{}
	if err := json.Unmarshal(recA.Body.Bytes(), &respA); err != nil {
		t.Fatalf("Test G failed: Could not parse response JSON: %v", err)
	}

	dataA, ok := respA["data"].([]interface{})
	if !ok {
		t.Fatalf("Test G failed: Expected data array, got %T", respA["data"])
	}

	// Should contain only the 1 eligible popup
	if len(dataA) != 1 {
		t.Errorf("Test G failed: Expected exactly 1 eligible popup, got %d", len(dataA))
	} else {
		item := dataA[0].(map[string]interface{})
		if uint(item["id"].(float64)) != pEligible.ID {
			t.Errorf("Test G failed: Expected popup ID %d, got %v", pEligible.ID, item["id"])
		}

		// Verify DTO structure: only public fields present
		forbiddenKeys := []string{"website_id", "status", "created_by", "created_at", "updated_at", "deleted_at"}
		for _, fk := range forbiddenKeys {
			if _, exists := item[fk]; exists {
				t.Errorf("Test G failed: DTO leaked forbidden field %q", fk)
			}
		}
	}
	t.Logf("PASS: Test G — Public eligibility filtering & sanitized DTO contract completely preserved")
}
