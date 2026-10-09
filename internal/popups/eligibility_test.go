package popups

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-yaml"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"popup-manager-api/config"
	"popup-manager-api/internal/common"
	"popup-manager-api/internal/websites"
	"popup-manager-api/utils"
)

func setupTestEnvironment(t *testing.T) (*gorm.DB, *gin.Engine, websites.WebsiteRepository, PopupRepository, string) {
	gin.SetMode(gin.TestMode)

	// Attempt to load .env from current or parent directories
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load(".env")

	if err := config.Load(); err != nil {
		t.Skipf("Skipping test: config could not be loaded: %v", err)
		return nil, nil, nil, nil, ""
	}

	common.InitLogger(config.AppConfig.AppEnv)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.AppConfig.DBUser,
		config.AppConfig.DBPassword,
		config.AppConfig.DBHost,
		config.AppConfig.DBPort,
		config.AppConfig.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping test: database connection failed: %v", err)
		return nil, nil, nil, nil, ""
	}

	websiteRepo := websites.NewWebsiteRepository(db)
	websiteService := websites.NewWebsiteService(websiteRepo)
	websiteCtrl := websites.NewWebsiteController(websiteService)

	popupRepo := NewPopupRepository(db)
	popupService := NewPopupService(popupRepo, websiteRepo)
	popupCtrl := NewPopupController(popupService)

	router := gin.New()
	router.Use(gin.Recovery())

	api := router.Group("/api/v1")
	websites.RegisterRoutes(api, websiteCtrl)
	RegisterRoutes(api, popupCtrl)

	adminToken, err := utils.GenerateJWT(1, "admin@example.com", "Super Admin")
	if err != nil {
		t.Fatalf("Failed to generate admin JWT: %v", err)
	}

	return db, router, websiteRepo, popupRepo, adminToken
}

func TestServerSidePopupEligibility(t *testing.T) {
	db, router, _, popupRepo, adminToken := setupTestEnvironment(t)
	if db == nil {
		return
	}

	nowNano := time.Now().UnixNano()
	w1Key := fmt.Sprintf("wg_live_w1_%d", nowNano)
	w2Key := fmt.Sprintf("wg_live_w2_%d", nowNano)
	w3Key := fmt.Sprintf("wg_live_w3_%d", nowNano)

	w1 := websites.Website{
		WebsiteName: fmt.Sprintf("Eligibility Test Site 1 %d", nowNano),
		Domain:      fmt.Sprintf("elig1-%d.com", nowNano),
		Platform:    websites.PlatformHTML,
		WebsiteKey:  w1Key,
		Status:      true,
		CreatedBy:   1,
	}
	if err := db.Create(&w1).Error; err != nil {
		t.Fatalf("Failed to create website 1: %v", err)
	}

	w2 := websites.Website{
		WebsiteName: fmt.Sprintf("Eligibility Test Site 2 %d", nowNano),
		Domain:      fmt.Sprintf("elig2-%d.com", nowNano),
		Platform:    websites.PlatformHTML,
		WebsiteKey:  w2Key,
		Status:      true,
		CreatedBy:   1,
	}
	if err := db.Create(&w2).Error; err != nil {
		t.Fatalf("Failed to create website 2: %v", err)
	}

	w3 := websites.Website{
		WebsiteName: fmt.Sprintf("Eligibility Test Site 3 %d", nowNano),
		Domain:      fmt.Sprintf("elig3-%d.com", nowNano),
		Platform:    websites.PlatformHTML,
		WebsiteKey:  w3Key,
		Status:      true,
		CreatedBy:   1,
	}
	if err := db.Create(&w3).Error; err != nil {
		t.Fatalf("Failed to create website 3: %v", err)
	}

	// Defer cleanup of all test records
	defer func() {
		db.Unscoped().Where("website_id IN (?, ?, ?)", w1.ID, w2.ID, w3.ID).Delete(&Popup{})
		db.Unscoped().Where("id IN (?, ?, ?)", w1.ID, w2.ID, w3.ID).Delete(&websites.Website{})
	}()

	now := time.Now()
	past := now.Add(-2 * time.Hour)
	future := now.Add(2 * time.Hour)

	// Controlled test records for W1:
	// Test A: Active popup with no schedule (status=true, start=nil, end=nil) -> Eligible
	popA := Popup{
		WebsiteID: w1.ID,
		Title:     "Popup A: Active No Schedule",
		Content:   "<p>No schedule</p>",
		Position:  PositionCenter,
		Status:    true,
		StartTime: nil,
		EndTime:   nil,
		CreatedBy: 1,
	}
	if err := db.Create(&popA).Error; err != nil {
		t.Fatalf("Failed to create popA: %v", err)
	}

	// Test B: Start time in past, no end time -> Eligible
	popB := Popup{
		WebsiteID: w1.ID,
		Title:     "Popup B: Past Start Time",
		Content:   "<p>Past start</p>",
		Position:  PositionTopBanner,
		Status:    true,
		StartTime: &past,
		EndTime:   nil,
		CreatedBy: 1,
	}
	if err := db.Create(&popB).Error; err != nil {
		t.Fatalf("Failed to create popB: %v", err)
	}

	// Test C: Start time in future, no end time -> Excluded
	popC := Popup{
		WebsiteID: w1.ID,
		Title:     "Popup C: Future Start Time",
		Content:   "<p>Future start</p>",
		Position:  PositionBottomLeft,
		Status:    true,
		StartTime: &future,
		EndTime:   nil,
		CreatedBy: 1,
	}
	if err := db.Create(&popC).Error; err != nil {
		t.Fatalf("Failed to create popC: %v", err)
	}

	// Test D: End time in past, no start time -> Excluded
	popD := Popup{
		WebsiteID: w1.ID,
		Title:     "Popup D: Past End Time",
		Content:   "<p>Past end</p>",
		Position:  PositionBottomRight,
		Status:    true,
		StartTime: nil,
		EndTime:   &past,
		CreatedBy: 1,
	}
	if err := db.Create(&popD).Error; err != nil {
		t.Fatalf("Failed to create popD: %v", err)
	}

	// Test E: Both dates within valid window -> Eligible
	popE := Popup{
		WebsiteID: w1.ID,
		Title:     "Popup E: Inside Window",
		Content:   "<p>Inside window</p>",
		Position:  PositionCenter,
		Status:    true,
		StartTime: &past,
		EndTime:   &future,
		CreatedBy: 1,
	}
	if err := db.Create(&popE).Error; err != nil {
		t.Fatalf("Failed to create popE: %v", err)
	}

	// Test F: Inactive popup -> Excluded
	popF := Popup{
		WebsiteID: w1.ID,
		Title:     "Popup F: Inactive",
		Content:   "<p>Inactive</p>",
		Position:  PositionCenter,
		Status:    false,
		StartTime: nil,
		EndTime:   nil,
		CreatedBy: 1,
	}
	if err := db.Create(&popF).Error; err != nil {
		t.Fatalf("Failed to create popF: %v", err)
	}
	if err := db.Model(&popF).Update("status", false).Error; err != nil {
		t.Fatalf("Failed to set popF status to false: %v", err)
	}

	// Test G: Soft-deleted popup -> Excluded
	popG := Popup{
		WebsiteID: w1.ID,
		Title:     "Popup G: Soft Deleted",
		Content:   "<p>Soft deleted</p>",
		Position:  PositionCenter,
		Status:    true,
		StartTime: nil,
		EndTime:   nil,
		CreatedBy: 1,
	}
	if err := db.Create(&popG).Error; err != nil {
		t.Fatalf("Failed to create popG: %v", err)
	}
	if err := db.Delete(&popG).Error; err != nil {
		t.Fatalf("Failed to soft-delete popG: %v", err)
	}

	// Test H: Popup belonging to another website (W2) -> Excluded from W1
	popH := Popup{
		WebsiteID: w2.ID,
		Title:     "Popup H: Another Website",
		Content:   "<p>Other website</p>",
		Position:  PositionCenter,
		Status:    true,
		StartTime: nil,
		EndTime:   nil,
		CreatedBy: 1,
	}
	if err := db.Create(&popH).Error; err != nil {
		t.Fatalf("Failed to create popH: %v", err)
	}

	// Helper to make public GET request
	callPublicAPI := func(key string) (*httptest.ResponseRecorder, map[string]interface{}) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/public/popups?website_key="+key, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		var resp map[string]interface{}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		return rec, resp
	}

	recW1, respW1 := callPublicAPI(w1Key)
	if recW1.Code != http.StatusOK {
		t.Fatalf("Expected HTTP 200 for W1, got %d: %s", recW1.Code, recW1.Body.String())
	}

	dataW1, ok := respW1["data"].([]interface{})
	if !ok {
		t.Fatalf("Expected data array in response, got %T", respW1["data"])
	}

	idInList := func(targetID uint) bool {
		for _, item := range dataW1 {
			m, ok := item.(map[string]interface{})
			if ok && uint(m["id"].(float64)) == targetID {
				return true
			}
		}
		return false
	}

	// Test A: Active with no schedule -> returned
	if !idInList(popA.ID) {
		t.Errorf("Test A failed: Popup A (active, no schedule) was not returned")
	} else {
		t.Logf("PASS: Test A — Active popup with no schedule (ID: %d) successfully returned", popA.ID)
	}

	// Test B: Start time in past, no end time -> returned
	if !idInList(popB.ID) {
		t.Errorf("Test B failed: Popup B (start in past, no end) was not returned")
	} else {
		t.Logf("PASS: Test B — Start time in past (ID: %d) successfully returned", popB.ID)
	}

	// Test C: Start time in future -> excluded
	if idInList(popC.ID) {
		t.Errorf("Test C failed: Popup C (start in future) was unexpectedly returned")
	} else {
		t.Logf("PASS: Test C — Start time in future (ID: %d) successfully excluded", popC.ID)
	}

	// Test D: End time in past -> excluded
	if idInList(popD.ID) {
		t.Errorf("Test D failed: Popup D (end in past) was unexpectedly returned")
	} else {
		t.Logf("PASS: Test D — End time in past (ID: %d) successfully excluded", popD.ID)
	}

	// Test E: Both dates within valid window -> returned
	if !idInList(popE.ID) {
		t.Errorf("Test E failed: Popup E (inside window) was not returned")
	} else {
		t.Logf("PASS: Test E — Both dates within window (ID: %d) successfully returned", popE.ID)
	}

	// Test F: Inactive popup -> excluded
	if idInList(popF.ID) {
		t.Errorf("Test F failed: Popup F (inactive status=false) was unexpectedly returned")
	} else {
		t.Logf("PASS: Test F — Inactive popup (ID: %d) successfully excluded", popF.ID)
	}

	// Test G: Soft-deleted popup -> excluded
	if idInList(popG.ID) {
		t.Errorf("Test G failed: Popup G (soft-deleted) was unexpectedly returned")
	} else {
		t.Logf("PASS: Test G — Soft-deleted popup (ID: %d) successfully excluded", popG.ID)
	}

	// Test H: Another website's popup -> excluded from W1, returned for W2
	if idInList(popH.ID) {
		t.Errorf("Test H failed: Popup H (belongs to W2) was returned for W1")
	}
	recW2, respW2 := callPublicAPI(w2Key)
	if recW2.Code != http.StatusOK {
		t.Errorf("Expected HTTP 200 for W2, got %d", recW2.Code)
	}
	dataW2, _ := respW2["data"].([]interface{})
	foundInW2 := false
	for _, item := range dataW2 {
		m, ok := item.(map[string]interface{})
		if ok && uint(m["id"].(float64)) == popH.ID {
			foundInW2 = true
		}
	}
	if !foundInW2 {
		t.Errorf("Test H failed: Popup H was not returned for its own website W2")
	} else {
		t.Logf("PASS: Test H — Other website popup (ID: %d) excluded from W1 and present in W2", popH.ID)
	}

	// Test I: Empty result -> HTTP 200 with "data": []
	recW3, respW3 := callPublicAPI(w3Key)
	if recW3.Code != http.StatusOK {
		t.Errorf("Test I failed: Expected HTTP 200 for W3, got %d", recW3.Code)
	}
	if respW3["success"] != true {
		t.Errorf("Test I failed: Expected success=true, got %v", respW3["success"])
	}
	if respW3["message"] != "Public popups fetched successfully" {
		t.Errorf("Test I failed: Expected message 'Public popups fetched successfully', got %v", respW3["message"])
	}
	rawBodyW3 := recW3.Body.String()
	if !strings.Contains(rawBodyW3, `"data":[]`) {
		t.Errorf("Test I failed: Expected '\"data\":[]' in raw JSON body, got: %s", rawBodyW3)
	} else {
		t.Logf("PASS: Test I — Empty result returned HTTP 200 with JSON \"data\":[] (body: %s)", rawBodyW3)
	}

	// Test J: Unknown website key -> HTTP 404
	recUnknown, respUnknown := callPublicAPI("wg_live_nonexistent_key_123456789012345678901234567890")
	if recUnknown.Code != http.StatusNotFound {
		t.Errorf("Test J failed: Expected HTTP 404 for unknown website key, got %d", recUnknown.Code)
	}
	if respUnknown["success"] != false {
		t.Errorf("Test J failed: Expected success=false for unknown key, got %v", respUnknown["success"])
	} else {
		t.Logf("PASS: Test J — Unknown website key returned HTTP 404 Not Found (body: %s)", recUnknown.Body.String())
	}

	// Test J2: Missing website_key parameter -> HTTP 400
	reqMissing, _ := http.NewRequest(http.MethodGet, "/api/v1/public/popups", nil)
	recMissing := httptest.NewRecorder()
	router.ServeHTTP(recMissing, reqMissing)
	if recMissing.Code != http.StatusBadRequest {
		t.Errorf("Test J2 failed: Expected HTTP 400 for missing website_key, got %d", recMissing.Code)
	}
	var respMissing map[string]interface{}
	_ = json.Unmarshal(recMissing.Body.Bytes(), &respMissing)
	if respMissing["success"] != false || respMissing["message"] != "website_key query parameter is required" {
		t.Errorf("Test J2 failed: Expected message 'website_key query parameter is required', got %v", respMissing["message"])
	} else {
		t.Logf("PASS: Test J2 — Missing website_key parameter returned HTTP 400 Bad Request")
	}

	// Test J3: Empty or whitespace-only website_key parameter -> HTTP 400
	reqEmpty, _ := http.NewRequest(http.MethodGet, "/api/v1/public/popups?website_key=%20%20%20", nil)
	recEmpty := httptest.NewRecorder()
	router.ServeHTTP(recEmpty, reqEmpty)
	if recEmpty.Code != http.StatusBadRequest {
		t.Errorf("Test J3 failed: Expected HTTP 400 for whitespace website_key, got %d", recEmpty.Code)
	}
	var respEmpty map[string]interface{}
	_ = json.Unmarshal(recEmpty.Body.Bytes(), &respEmpty)
	if respEmpty["success"] != false || respEmpty["message"] != "website_key query parameter is required" {
		t.Errorf("Test J3 failed: Expected message 'website_key query parameter is required', got %v", respEmpty["message"])
	} else {
		t.Logf("PASS: Test J3 — Whitespace-only website_key parameter returned HTTP 400 Bad Request")
	}

	// Test K: Authentication boundaries & Admin endpoints unchanged
	// K1: Public endpoint accessible without admin JWT
	recPubNoAuth, _ := callPublicAPI(w1Key)
	if recPubNoAuth.Code != http.StatusOK {
		t.Errorf("Test K1 failed: Public endpoint rejected unauthenticated request: %d", recPubNoAuth.Code)
	} else {
		t.Logf("PASS: Test K1 — Public delivery endpoint accessible without authentication (HTTP 200)")
	}

	// K2: Admin endpoint GET /api/v1/popups/website/{website_id} requires JWT (returns 401 without auth)
	reqAdminNoAuth, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/popups/website/%d", w1.ID), nil)
	recAdminNoAuth := httptest.NewRecorder()
	router.ServeHTTP(recAdminNoAuth, reqAdminNoAuth)
	if recAdminNoAuth.Code != http.StatusUnauthorized {
		t.Errorf("Test K2 failed: Admin endpoint allowed unauthenticated access: %d", recAdminNoAuth.Code)
	} else {
		t.Logf("PASS: Test K2 — Admin endpoint GET /api/v1/popups/website/:id rejected unauthenticated request (HTTP 401)")
	}

	// K3: Admin endpoint GET /api/v1/popups/website/{website_id} works with valid JWT
	reqAdminAuth, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/popups/website/%d", w1.ID), nil)
	reqAdminAuth.Header.Set("Authorization", "Bearer "+adminToken)
	recAdminAuth := httptest.NewRecorder()
	router.ServeHTTP(recAdminAuth, reqAdminAuth)
	if recAdminAuth.Code != http.StatusOK {
		t.Errorf("Test K3 failed: Admin endpoint rejected authenticated request: %d", recAdminAuth.Code)
	}

	// Admin listing returns ALL 6 non-deleted popups for W1 (A, B, C, D, E, F), showing admin filtering is unchanged!
	var adminResp map[string]interface{}
	_ = json.Unmarshal(recAdminAuth.Body.Bytes(), &adminResp)
	adminData, _ := adminResp["data"].([]interface{})
	if len(adminData) != 6 {
		t.Errorf("Test K3 failed: Admin listing expected 6 popups for W1, got %d", len(adminData))
	} else {
		t.Logf("PASS: Test K3 — Admin endpoint returned all 6 popups (active, inactive, future, past), admin behavior unchanged")
	}

	// K4: Admin endpoint rejected with malformed Authorization header -> HTTP 401
	reqAdminMalformed, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/popups/website/%d", w1.ID), nil)
	reqAdminMalformed.Header.Set("Authorization", "InvalidHeaderFormat")
	recAdminMalformed := httptest.NewRecorder()
	router.ServeHTTP(recAdminMalformed, reqAdminMalformed)
	if recAdminMalformed.Code != http.StatusUnauthorized {
		t.Errorf("Test K4 failed: Expected HTTP 401 for malformed Authorization header, got %d", recAdminMalformed.Code)
	} else {
		t.Logf("PASS: Test K4 — Malformed authorization header rejected with HTTP 401")
	}

	// K5: Admin endpoint rejected with invalid/expired JWT token -> HTTP 401
	reqAdminInvalidToken, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/popups/website/%d", w1.ID), nil)
	reqAdminInvalidToken.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature")
	recAdminInvalidToken := httptest.NewRecorder()
	router.ServeHTTP(recAdminInvalidToken, reqAdminInvalidToken)
	if recAdminInvalidToken.Code != http.StatusUnauthorized {
		t.Errorf("Test K5 failed: Expected HTTP 401 for invalid JWT token, got %d", recAdminInvalidToken.Code)
	} else {
		t.Logf("PASS: Test K5 — Invalid JWT token rejected with HTTP 401")
	}

	// Ordering test: created_at DESC, id DESC
	var prevCreatedAt time.Time
	var prevID uint
	for i, item := range dataW1 {
		m := item.(map[string]interface{})
		id := uint(m["id"].(float64))

		var p Popup
		db.First(&p, id)

		if i > 0 {
			if p.CreatedAt.After(prevCreatedAt) {
				t.Errorf("Ordering failed: Popup %d created_at %v is after previous %v", p.ID, p.CreatedAt, prevCreatedAt)
			} else if p.CreatedAt.Equal(prevCreatedAt) && p.ID > prevID {
				t.Errorf("Ordering tie-breaker failed: Popup %d id > previous id %d with same created_at", p.ID, prevID)
			}
		}
		prevCreatedAt = p.CreatedAt
		prevID = p.ID
	}
	t.Logf("PASS: Ordering test — Eligible popups sorted by created_at DESC, id DESC")

	// Boundary test: Exact start and exact end match
	boundaryTime := time.Date(2026, 10, 9, 12, 0, 0, 0, time.Local)
	popExactStart := Popup{
		WebsiteID: w1.ID,
		Title:     "Exact Start Boundary",
		Content:   "Boundary",
		Position:  PositionCenter,
		Status:    true,
		StartTime: &boundaryTime,
		EndTime:   nil,
		CreatedBy: 1,
	}
	db.Create(&popExactStart)

	popExactEnd := Popup{
		WebsiteID: w1.ID,
		Title:     "Exact End Boundary",
		Content:   "Boundary",
		Position:  PositionCenter,
		Status:    true,
		StartTime: nil,
		EndTime:   &boundaryTime,
		CreatedBy: 1,
	}
	db.Create(&popExactEnd)

	eligibleAtBoundary, err := popupRepo.GetEligiblePublicByWebsite(w1.ID, boundaryTime)
	if err != nil {
		t.Fatalf("GetEligiblePublicByWebsite boundary query failed: %v", err)
	}
	foundExactStart := false
	foundExactEnd := false
	for _, p := range eligibleAtBoundary {
		if p.ID == popExactStart.ID {
			foundExactStart = true
		}
		if p.ID == popExactEnd.ID {
			foundExactEnd = true
		}
	}
	if !foundExactStart {
		t.Errorf("Boundary test failed: popup with start_time == now was not included")
	}
	if !foundExactEnd {
		t.Errorf("Boundary test failed: popup with end_time == now was not included")
	}
	if foundExactStart && foundExactEnd {
		t.Logf("PASS: Boundary test — Inclusive boundaries verified for start_time == now and end_time == now")
	}

	// DTO Contract verification
	forbiddenKeys := []string{"website_id", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	for _, item := range dataW1 {
		m := item.(map[string]interface{})
		for _, fk := range forbiddenKeys {
			if _, exists := m[fk]; exists {
				t.Errorf("Public popup DTO leaked forbidden field %q", fk)
			}
		}
	}
	t.Logf("PASS: DTO contract — Response contains only public fields, zero admin or internal metadata leaked")
}

func TestAdminCRUD_Regression(t *testing.T) {
	db, router, websiteRepo, _, adminToken := setupTestEnvironment(t)
	if db == nil {
		return
	}

	nowNano := time.Now().UnixNano()
	siteDomain := fmt.Sprintf("crud-regression-%d.com", nowNano)

	// Step 1: Admin creates website
	createSitePayload := map[string]interface{}{
		"website_name": fmt.Sprintf("CRUD Site %d", nowNano),
		"domain":       siteDomain,
		"platform":     "HTML",
	}
	siteBody, _ := json.Marshal(createSitePayload)
	reqCreateSite, _ := http.NewRequest(http.MethodPost, "/api/v1/websites", bytes.NewReader(siteBody))
	reqCreateSite.Header.Set("Authorization", "Bearer "+adminToken)
	reqCreateSite.Header.Set("Content-Type", "application/json")
	recCreateSite := httptest.NewRecorder()
	router.ServeHTTP(recCreateSite, reqCreateSite)

	if recCreateSite.Code != http.StatusOK {
		t.Fatalf("Admin Create Website failed with code %d: %s", recCreateSite.Code, recCreateSite.Body.String())
	}

	// Lookup created website to get ID and generated website_key
	createdSite, err := websiteRepo.GetByDomain(siteDomain)
	if err != nil || createdSite == nil {
		t.Fatalf("Failed to fetch created website from DB: %v", err)
	}

	defer func() {
		db.Unscoped().Where("website_id = ?", createdSite.ID).Delete(&Popup{})
		db.Unscoped().Where("id = ?", createdSite.ID).Delete(&websites.Website{})
		db.Unscoped().Where("domain LIKE 'crud-regression-%'").Delete(&websites.Website{})
	}()

	if createdSite.WebsiteKey == "" || !strings.HasPrefix(createdSite.WebsiteKey, "wg_live_") {
		t.Fatalf("Expected website_key with prefix 'wg_live_', got %q", createdSite.WebsiteKey)
	}
	t.Logf("PASS: Website created with ID %d and Key %s", createdSite.ID, createdSite.WebsiteKey)

	// Step 2: Admin gets website by ID
	reqGetSite, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/websites/%d", createdSite.ID), nil)
	reqGetSite.Header.Set("Authorization", "Bearer "+adminToken)
	recGetSite := httptest.NewRecorder()
	router.ServeHTTP(recGetSite, reqGetSite)

	if recGetSite.Code != http.StatusOK {
		t.Fatalf("Admin Get Website failed with code %d: %s", recGetSite.Code, recGetSite.Body.String())
	}
	var getSiteResp map[string]interface{}
	_ = json.Unmarshal(recGetSite.Body.Bytes(), &getSiteResp)
	siteData, ok := getSiteResp["data"].(map[string]interface{})
	if !ok || siteData["domain"] != siteDomain {
		t.Errorf("Get Website expected domain %s, got %v", siteDomain, siteData["domain"])
	}
	t.Logf("PASS: Admin Get Website verified")

	// Step 3: Admin updates website
	updateSitePayload := map[string]interface{}{
		"website_name": fmt.Sprintf("Updated CRUD Site %d", nowNano),
		"domain":       siteDomain,
		"platform":     "WordPress",
		"status":       true,
	}
	updSiteBody, _ := json.Marshal(updateSitePayload)
	reqUpdSite, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/websites/%d", createdSite.ID), bytes.NewReader(updSiteBody))
	reqUpdSite.Header.Set("Authorization", "Bearer "+adminToken)
	reqUpdSite.Header.Set("Content-Type", "application/json")
	recUpdSite := httptest.NewRecorder()
	router.ServeHTTP(recUpdSite, reqUpdSite)

	if recUpdSite.Code != http.StatusOK {
		t.Fatalf("Admin Update Website failed with code %d: %s", recUpdSite.Code, recUpdSite.Body.String())
	}
	t.Logf("PASS: Admin Update Website verified")

	// Step 4: Admin creates popup
	createPopupPayload := map[string]interface{}{
		"website_id": createdSite.ID,
		"title":      "Integration Test Popup",
		"content":    "<p>Exclusive discount banner</p>",
		"position":   "center",
		"status":     true,
	}
	popupBody, _ := json.Marshal(createPopupPayload)
	reqCreatePopup, _ := http.NewRequest(http.MethodPost, "/api/v1/popups", bytes.NewReader(popupBody))
	reqCreatePopup.Header.Set("Authorization", "Bearer "+adminToken)
	reqCreatePopup.Header.Set("Content-Type", "application/json")
	recCreatePopup := httptest.NewRecorder()
	router.ServeHTTP(recCreatePopup, reqCreatePopup)

	if recCreatePopup.Code != http.StatusOK {
		t.Fatalf("Admin Create Popup failed with code %d: %s", recCreatePopup.Code, recCreatePopup.Body.String())
	}

	var createdPopup Popup
	if err := db.Where("website_id = ? AND title = ?", createdSite.ID, "Integration Test Popup").First(&createdPopup).Error; err != nil {
		t.Fatalf("Failed to retrieve created popup from DB: %v", err)
	}
	t.Logf("PASS: Popup created with ID %d for website ID %d", createdPopup.ID, createdSite.ID)

	// Step 5: Admin gets popup by ID
	reqGetPopup, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/popups/%d", createdPopup.ID), nil)
	reqGetPopup.Header.Set("Authorization", "Bearer "+adminToken)
	recGetPopup := httptest.NewRecorder()
	router.ServeHTTP(recGetPopup, reqGetPopup)

	if recGetPopup.Code != http.StatusOK {
		t.Fatalf("Admin Get Popup failed with code %d: %s", recGetPopup.Code, recGetPopup.Body.String())
	}
	t.Logf("PASS: Admin Get Popup verified")

	// Step 6: Admin updates popup
	updatePopupPayload := map[string]interface{}{
		"title":    "Updated Integration Popup",
		"content":  "<p>Updated content</p>",
		"position": PositionBottomRight,
		"status":   true,
	}
	updPopupBody, _ := json.Marshal(updatePopupPayload)
	reqUpdPopup, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/popups/%d", createdPopup.ID), bytes.NewReader(updPopupBody))
	reqUpdPopup.Header.Set("Authorization", "Bearer "+adminToken)
	reqUpdPopup.Header.Set("Content-Type", "application/json")
	recUpdPopup := httptest.NewRecorder()
	router.ServeHTTP(recUpdPopup, reqUpdPopup)

	if recUpdPopup.Code != http.StatusOK {
		t.Fatalf("Admin Update Popup failed with code %d: %s", recUpdPopup.Code, recUpdPopup.Body.String())
	}
	t.Logf("PASS: Admin Update Popup verified")

	// Step 7: Public delivery endpoint delivers the updated popup via created website key
	reqPublic, _ := http.NewRequest(http.MethodGet, "/api/v1/public/popups?website_key="+createdSite.WebsiteKey, nil)
	recPublic := httptest.NewRecorder()
	router.ServeHTTP(recPublic, reqPublic)

	if recPublic.Code != http.StatusOK {
		t.Fatalf("Public Delivery failed with code %d: %s", recPublic.Code, recPublic.Body.String())
	}
	var pubResp map[string]interface{}
	_ = json.Unmarshal(recPublic.Body.Bytes(), &pubResp)
	pubData, ok := pubResp["data"].([]interface{})
	if !ok || len(pubData) != 1 {
		t.Fatalf("Expected 1 public popup, got %v", pubResp["data"])
	}
	item := pubData[0].(map[string]interface{})
	if item["title"] != "Updated Integration Popup" || item["position"] != PositionBottomRight {
		t.Errorf("Unexpected public popup data: %v", item)
	}

	// Verify public DTO sanitization for newly delivered popup
	forbiddenKeys := []string{"website_id", "status", "created_by", "created_at", "updated_at", "deleted_at"}
	for _, fk := range forbiddenKeys {
		if _, exists := item[fk]; exists {
			t.Errorf("Public popup DTO leaked forbidden field %q", fk)
		}
	}
	t.Logf("PASS: Public Delivery of created popup succeeded with sanitized DTO")

	// Step 8: Admin deletes popup
	reqDelPopup, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/popups/%d", createdPopup.ID), nil)
	reqDelPopup.Header.Set("Authorization", "Bearer "+adminToken)
	recDelPopup := httptest.NewRecorder()
	router.ServeHTTP(recDelPopup, reqDelPopup)

	if recDelPopup.Code != http.StatusOK {
		t.Fatalf("Admin Delete Popup failed with code %d: %s", recDelPopup.Code, recDelPopup.Body.String())
	}
	t.Logf("PASS: Admin Delete Popup verified")

	// Step 9: Public delivery returns empty array [] after popup deletion
	reqPublicAfterDel, _ := http.NewRequest(http.MethodGet, "/api/v1/public/popups?website_key="+createdSite.WebsiteKey, nil)
	recPublicAfterDel := httptest.NewRecorder()
	router.ServeHTTP(recPublicAfterDel, reqPublicAfterDel)

	if recPublicAfterDel.Code != http.StatusOK {
		t.Fatalf("Public Delivery after deletion failed with code %d: %s", recPublicAfterDel.Code, recPublicAfterDel.Body.String())
	}
	if !strings.Contains(recPublicAfterDel.Body.String(), `"data":[]`) {
		t.Errorf("Expected '\"data\":[]' after deletion, got %s", recPublicAfterDel.Body.String())
	}
	t.Logf("PASS: Public Delivery returned empty array after popup deletion")

	// Step 10: Admin deletes website
	reqDelSite, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/websites/%d", createdSite.ID), nil)
	reqDelSite.Header.Set("Authorization", "Bearer "+adminToken)
	recDelSite := httptest.NewRecorder()
	router.ServeHTTP(recDelSite, reqDelSite)

	if recDelSite.Code != http.StatusOK {
		t.Fatalf("Admin Delete Website failed with code %d: %s", recDelSite.Code, recDelSite.Body.String())
	}
	t.Logf("PASS: Admin Delete Website verified")

	// Step 11: Public delivery returns 404 after website deletion
	reqPublicAfterSiteDel, _ := http.NewRequest(http.MethodGet, "/api/v1/public/popups?website_key="+createdSite.WebsiteKey, nil)
	recPublicAfterSiteDel := httptest.NewRecorder()
	router.ServeHTTP(recPublicAfterSiteDel, reqPublicAfterSiteDel)

	if recPublicAfterSiteDel.Code != http.StatusNotFound {
		t.Fatalf("Public Delivery for deleted website expected 404, got %d", recPublicAfterSiteDel.Code)
	}
	t.Logf("PASS: Public Delivery returned 404 for deleted website")
}

func TestSystemConfigAndLogging_Regression(t *testing.T) {
	_, _, _, _, _ = setupTestEnvironment(t)
	if config.AppConfig == nil {
		t.Fatal("config.AppConfig is nil")
	}
	if config.AppConfig.AppPort == "" {
		t.Error("AppConfig.AppPort should not be empty")
	}
	if config.AppConfig.DBHost == "" {
		t.Error("AppConfig.DBHost should not be empty")
	}
	if slog.Default() == nil {
		t.Error("slog.Default() should be non-nil")
	}
	t.Logf("PASS: System configuration and structured logging verified")
}

func TestOpenAPI_SpecificationValidity(t *testing.T) {
	data, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		data, err = os.ReadFile("docs/openapi.yaml")
	}
	if err != nil {
		t.Fatalf("Failed to read openapi.yaml: %v", err)
	}

	var root map[string]interface{}
	if err := yaml.Unmarshal(data, &root); err != nil {
		t.Fatalf("docs/openapi.yaml is invalid YAML: %v", err)
	}

	paths, ok := root["paths"].(map[string]interface{})
	if !ok {
		t.Fatalf("Missing 'paths' section in openapi.yaml")
	}

	publicPopups, ok := paths["/api/v1/public/popups"].(map[string]interface{})
	if !ok {
		t.Fatalf("Missing '/api/v1/public/popups' in paths")
	}

	if _, hasGet := publicPopups["get"]; !hasGet {
		t.Errorf("Missing GET operation for /api/v1/public/popups")
	}
	if _, hasOptions := publicPopups["options"]; !hasOptions {
		t.Errorf("Missing OPTIONS operation for /api/v1/public/popups")
	}

	// Verify all $ref in document resolve to components
	components, _ := root["components"].(map[string]interface{})
	schemas, _ := components["schemas"].(map[string]interface{})
	parameters, _ := components["parameters"].(map[string]interface{})

	lines := strings.Split(string(data), "\n")
	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "$ref:") || strings.Contains(trimmed, "$ref:") {
			idx := strings.Index(trimmed, "$ref:")
			refVal := strings.Trim(trimmed[idx+5:], " '\"")
			if strings.HasPrefix(refVal, "#/components/schemas/") {
				schemaName := strings.TrimPrefix(refVal, "#/components/schemas/")
				if _, exists := schemas[schemaName]; !exists {
					t.Errorf("Line %d: unresolved schema reference %q", lineNum+1, refVal)
				}
			} else if strings.HasPrefix(refVal, "#/components/parameters/") {
				paramName := strings.TrimPrefix(refVal, "#/components/parameters/")
				if _, exists := parameters[paramName]; !exists {
					t.Errorf("Line %d: unresolved parameter reference %q", lineNum+1, refVal)
				}
			}
		}
	}
	t.Logf("PASS: docs/openapi.yaml syntax, public route specs, and all $ref references validated successfully")
}
