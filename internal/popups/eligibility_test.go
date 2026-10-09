package popups

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"popup-manager-api/config"
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
	popupRepo := NewPopupRepository(db)
	popupService := NewPopupService(popupRepo, websiteRepo)
	popupCtrl := NewPopupController(popupService)

	router := gin.New()
	router.Use(gin.Recovery())

	api := router.Group("/api/v1")
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
