package websites

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"popup-manager-api/internal/common"

	"gorm.io/gorm"
)

type WebsiteService interface {
	Create(request CreateWebsiteRequest, createdBy uint) error

	GetAll(params common.QueryParams) ([]Website, int64, error)

	GetByID(id uint) (*Website, error)

	Update(id uint, request UpdateWebsiteRequest) error

	Delete(id uint) error
}

type websiteService struct {
	repository WebsiteRepository
}

func NewWebsiteService(repository WebsiteRepository) WebsiteService {

	return &websiteService{
		repository: repository,
	}

}

var (
	domainHostRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*\.[a-zA-Z]{2,}$`)
	ipHostRegex     = regexp.MustCompile(`^127(?:\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$`)
)

// NormalizeAndValidateDomain returns the canonical website origin host:
// - Strips scheme (http://, https://, //)
// - Strips leading www.
// - Normalizes localhost and 127.0.0.1 (with optional port)
// - Ignores page paths so all routes on the same host share one website key
func NormalizeAndValidateDomain(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", errors.New("domain or URL cannot be empty")
	}

	// 1. Strip protocol (http://, https://, or //)
	reProtocol := regexp.MustCompile(`(?i)^(?:https?:)?//`)
	trimmed = reProtocol.ReplaceAllString(trimmed, "")

	// 2. Strip fragment (#) and query string (?) if any
	if idx := strings.Index(trimmed, "#"); idx != -1 {
		trimmed = trimmed[:idx]
	}
	if idx := strings.Index(trimmed, "?"); idx != -1 {
		trimmed = trimmed[:idx]
	}

	// A website key belongs to the origin host; routes are handled by popup targeting.
	var hostPart string
	if slashIdx := strings.Index(trimmed, "/"); slashIdx != -1 {
		hostPart = trimmed[:slashIdx]
	} else {
		hostPart = trimmed
	}

	hostPart = strings.ToLower(strings.TrimSpace(hostPart))
	// Strip leading www.
	if strings.HasPrefix(hostPart, "www.") {
		hostPart = strings.TrimPrefix(hostPart, "www.")
	}

	if hostPart == "" {
		return "", errors.New("domain or host cannot be empty")
	}

	// Validate host (with optional port)
	hostWithoutPort := hostPart
	if colonIdx := strings.LastIndex(hostPart, ":"); colonIdx != -1 {
		hostWithoutPort = hostPart[:colonIdx]
		portStr := hostPart[colonIdx+1:]
		portNum, err := strconv.Atoi(portStr)
		if err != nil || portNum < 1 || portNum > 65535 {
			return "", errors.New("invalid port number in domain or URL")
		}
	}

	// Host must be either localhost, loopback IP, or valid public domain
	isLocalhost := hostWithoutPort == "localhost" || ipHostRegex.MatchString(hostWithoutPort)
	if !isLocalhost && !domainHostRegex.MatchString(hostWithoutPort) {
		return "", errors.New("invalid domain format")
	}

	return hostPart, nil
}

// ExtractWebsitePath extracts a normalized route from a previously registered URL.
// It is used during migration to preserve page-specific popup assignments.
func ExtractWebsitePath(input string) string {
	value := strings.TrimSpace(input)
	value = regexp.MustCompile(`(?i)^(?:https?:)?//`).ReplaceAllString(value, "")
	if index := strings.IndexAny(value, "?#"); index >= 0 {
		value = value[:index]
	}
	slashIndex := strings.Index(value, "/")
	if slashIndex < 0 {
		return ""
	}
	path := regexp.MustCompile(`/+`).ReplaceAllString(value[slashIndex:], "/")
	path = strings.TrimRight(path, "/")
	if path == "" {
		return ""
	}
	return path
}

func isValidPlatform(platform string) bool {

	switch platform {

	case PlatformWordPress,
		PlatformReact,
		PlatformPython,
		PlatformLaravel,
		PlatformNodeJS,
		PlatformHTML,
		PlatformJava,
		PlatformVue,
		PlatformGo,
		PlatformOther:

		return true

	default:

		return false
	}

}

// GenerateWebsiteKey generates a cryptographically secure random website key.
// Format: wg_live_<48 hex chars> (total length: 56 characters, fitting VARCHAR(64)).
func GenerateWebsiteKey() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to read secure random bytes: %w", err)
	}
	return "wg_live_" + hex.EncodeToString(bytes), nil
}

func isWebsiteKeyCollision(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate entry") && strings.Contains(msg, "website_key")
}

func (s *websiteService) Create(request CreateWebsiteRequest, createdBy uint) error {

	vErr := common.NewValidationError()

	websiteName := strings.TrimSpace(request.WebsiteName)
	if websiteName == "" {
		vErr.Add("website_name", "website name cannot be empty")
	}

	// Validate Platform
	if !isValidPlatform(request.Platform) {
		vErr.Add("platform", "invalid platform")
	}

	domain, dErr := NormalizeAndValidateDomain(request.Domain)
	if dErr != nil {
		vErr.Add("domain", dErr.Error())
	}

	if vErr.HasErrors() {
		return vErr
	}

	// Check duplicate domain
	existing, err := s.repository.GetByDomain(domain)
	if err != nil && !errors.Is(err, common.ErrNotFound) && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existing != nil {
		return common.ErrDuplicateDomain
	}

	const maxRetries = 5
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		key, err := GenerateWebsiteKey()
		if err != nil {
			return errors.New("failed to generate website key")
		}

		website := Website{
			WebsiteName: websiteName,
			Domain:      domain,
			Platform:    request.Platform,
			WebsiteKey:  key,
			Status:      true,
			CreatedBy:   createdBy,
		}

		err = s.repository.Create(&website)
		if err == nil {
			return nil
		}

		if isWebsiteKeyCollision(err) {
			lastErr = err
			slog.Warn("website key collision occurred, retrying with a new key", "attempt", attempt+1)
			continue
		}

		return err
	}

	slog.Error("failed to generate unique website key after max retries", "error", lastErr)
	return errors.New("failed to generate unique website key")
}

func (s *websiteService) GetAll(params common.QueryParams) ([]Website, int64, error) {

	return s.repository.GetAll(params)

}

func (s *websiteService) GetByID(id uint) (*Website, error) {

	return s.repository.GetByID(id)

}

func (s *websiteService) Update(id uint, request UpdateWebsiteRequest) error {

	website, err := s.repository.GetByID(id)

	if err != nil {
		return err
	}

	vErr := common.NewValidationError()

	websiteName := strings.TrimSpace(request.WebsiteName)
	if websiteName == "" {
		vErr.Add("website_name", "website name cannot be empty")
	}

	if !isValidPlatform(request.Platform) {
		vErr.Add("platform", "invalid platform")
	}

	domain, dErr := NormalizeAndValidateDomain(request.Domain)
	if dErr != nil {
		vErr.Add("domain", dErr.Error())
	}

	if vErr.HasErrors() {
		return vErr
	}

	// Check duplicate domain for update
	existing, err := s.repository.GetByDomain(domain)
	if err != nil && !errors.Is(err, common.ErrNotFound) && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existing != nil && existing.ID != id {
		return common.ErrDuplicateDomain
	}

	website.WebsiteName = websiteName
	website.Domain = domain
	website.Platform = request.Platform
	website.Status = request.Status

	if strings.TrimSpace(website.WebsiteKey) == "" {
		key, err := GenerateWebsiteKey()
		if err != nil {
			return errors.New("failed to generate website key")
		}
		website.WebsiteKey = key
	}

	return s.repository.Update(website)
}

func (s *websiteService) Delete(id uint) error {
	return s.repository.Delete(id)
}
