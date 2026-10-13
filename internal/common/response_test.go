package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func performHandleErrorTest(err error) (*httptest.ResponseRecorder, APIResponse) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	HandleError(c, err)

	var resp APIResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return w, resp
}

func TestHandleError_ValidationErrors(t *testing.T) {
	vErr := NewValidationError()
	vErr.Add("domain", "invalid domain format")
	vErr.Add("website_name", "cannot be empty")

	w, resp := performHandleErrorTest(vErr)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("Expected status 422, got %d", w.Code)
	}
	if resp.Success != false {
		t.Errorf("Expected success to be false, got true")
	}
	if resp.Message != "Validation failed" {
		t.Errorf("Expected message 'Validation failed', got %q", resp.Message)
	}
	if len(resp.Errors) != 2 || resp.Errors["domain"] != "invalid domain format" {
		t.Errorf("Unexpected validation errors payload: %v", resp.Errors)
	}

	// Wrapped validation error
	wrappedVErr := fmt.Errorf("service failed: %w", vErr)
	wWrapped, respWrapped := performHandleErrorTest(wrappedVErr)

	if wWrapped.Code != http.StatusUnprocessableEntity {
		t.Errorf("Expected wrapped validation error to return 422, got %d", wWrapped.Code)
	}
	if len(respWrapped.Errors) != 2 {
		t.Errorf("Expected 2 field errors in wrapped validation, got %v", respWrapped.Errors)
	}
}

func TestHandleError_NotFoundErrors(t *testing.T) {
	// 1. Direct ErrNotFound
	w, resp := performHandleErrorTest(ErrNotFound)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for ErrNotFound, got %d", w.Code)
	}
	if resp.Message != ErrNotFound.Error() {
		t.Errorf("Expected message %q, got %q", ErrNotFound.Error(), resp.Message)
	}

	// 2. Wrapped ErrNotFound
	wWrap, _ := performHandleErrorTest(fmt.Errorf("item query: %w", ErrNotFound))
	if wWrap.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for wrapped ErrNotFound, got %d", wWrap.Code)
	}

	// 3. gorm.ErrRecordNotFound
	wGorm, respGorm := performHandleErrorTest(gorm.ErrRecordNotFound)
	if wGorm.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for gorm.ErrRecordNotFound, got %d", wGorm.Code)
	}
	if respGorm.Message != "resource not found" {
		t.Errorf("Expected 'resource not found', got %q", respGorm.Message)
	}

	// 4. Wrapped gorm.ErrRecordNotFound
	wGormWrap, _ := performHandleErrorTest(fmt.Errorf("repo find: %w", gorm.ErrRecordNotFound))
	if wGormWrap.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for wrapped gorm.ErrRecordNotFound, got %d", wGormWrap.Code)
	}

	// 5. Legacy strings
	wLegacy, respLegacy := performHandleErrorTest(errors.New("Popup not found"))
	if wLegacy.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for legacy 'Popup not found', got %d", wLegacy.Code)
	}
	if respLegacy.Message != "Popup not found" {
		t.Errorf("Expected 'Popup not found', got %q", respLegacy.Message)
	}
}

func TestHandleError_ConflictErrors(t *testing.T) {
	// Direct ErrDuplicateDomain
	w, resp := performHandleErrorTest(ErrDuplicateDomain)
	if w.Code != http.StatusConflict {
		t.Errorf("Expected 409 for ErrDuplicateDomain, got %d", w.Code)
	}
	if resp.Message != ErrDuplicateDomain.Error() {
		t.Errorf("Expected %q, got %q", ErrDuplicateDomain.Error(), resp.Message)
	}

	// Wrapped ErrDuplicateDomain
	wWrap, _ := performHandleErrorTest(fmt.Errorf("create website: %w", ErrDuplicateDomain))
	if wWrap.Code != http.StatusConflict {
		t.Errorf("Expected 409 for wrapped ErrDuplicateDomain, got %d", wWrap.Code)
	}

	// Direct ErrConflict
	wConf, _ := performHandleErrorTest(ErrConflict)
	if wConf.Code != http.StatusConflict {
		t.Errorf("Expected 409 for ErrConflict, got %d", wConf.Code)
	}
}

func TestHandleError_AuthenticationAndAuthorization(t *testing.T) {
	// ErrUnauthorized
	wAuth, _ := performHandleErrorTest(ErrUnauthorized)
	if wAuth.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for ErrUnauthorized, got %d", wAuth.Code)
	}

	// ErrInvalidCredentials
	wCred, _ := performHandleErrorTest(ErrInvalidCredentials)
	if wCred.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for ErrInvalidCredentials, got %d", wCred.Code)
	}

	// Legacy invalid email or password
	wLegacyAuth, _ := performHandleErrorTest(errors.New("invalid email or password"))
	if wLegacyAuth.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for legacy auth error, got %d", wLegacyAuth.Code)
	}

	// ErrForbidden
	wForbid, _ := performHandleErrorTest(ErrForbidden)
	if wForbid.Code != http.StatusForbidden {
		t.Errorf("Expected 403 for ErrForbidden, got %d", wForbid.Code)
	}
}

func TestHandleError_BadRequest(t *testing.T) {
	wBad, _ := performHandleErrorTest(ErrBadRequest)
	if wBad.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for ErrBadRequest, got %d", wBad.Code)
	}

	wVal, _ := performHandleErrorTest(ErrValidation)
	if wVal.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for ErrValidation, got %d", wVal.Code)
	}

	wReq, _ := performHandleErrorTest(errors.New("website_key is required"))
	if wReq.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for 'website_key is required', got %d", wReq.Code)
	}
}

func TestHandleError_AppError(t *testing.T) {
	appErr := NewAppError(http.StatusTeapot, "I'm a teapot", errors.New("underlying"))
	w, resp := performHandleErrorTest(appErr)
	if w.Code != http.StatusTeapot {
		t.Errorf("Expected 418, got %d", w.Code)
	}
	if resp.Message != "I'm a teapot" {
		t.Errorf("Expected 'I'm a teapot', got %q", resp.Message)
	}

	// AppError with status >= 500 sanitizes message
	internalAppErr := NewAppError(http.StatusInternalServerError, "sensitive internal info", errors.New("raw db fail"))
	wInt, respInt := performHandleErrorTest(internalAppErr)
	if wInt.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", wInt.Code)
	}
	if respInt.Message != "Internal server error" {
		t.Errorf("Expected safe 'Internal server error', got %q", respInt.Message)
	}
}

func TestHandleError_UnexpectedInternalErrorSanitization(t *testing.T) {
	// Raw SQL error or unknown database failure
	rawSQLErr := errors.New("Error 1064 (42000): You have an error in your SQL syntax near 'SELECT * FROM users'")
	w, resp := performHandleErrorTest(rawSQLErr)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 for unexpected error, got %d", w.Code)
	}
	if resp.Success != false {
		t.Errorf("Expected success false, got true")
	}
	if resp.Message != "Internal server error" {
		t.Errorf("Expected safe generic message 'Internal server error', got %q", resp.Message)
	}
	// Verify raw SQL error is NOT leaked in response
	if resp.Message == rawSQLErr.Error() {
		t.Fatalf("LEAK DETECTED: Raw SQL error was returned in response payload!")
	}
}
