package websites

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
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

func isValidDomain(domain string) bool {
	regex := regexp.MustCompile(`^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return regex.MatchString(domain)
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

	domain := strings.ToLower(strings.TrimSpace(request.Domain))
	if !isValidDomain(domain) {
		vErr.Add("domain", "invalid domain format")
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

	domain := strings.ToLower(strings.TrimSpace(request.Domain))
	if !isValidDomain(domain) {
		vErr.Add("domain", "invalid domain format")
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
