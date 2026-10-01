package websites

import (
	"errors"
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

func (s *websiteService) Create(request CreateWebsiteRequest, createdBy uint) error {

	// Validate Platform
	if !isValidPlatform(request.Platform) {
		return errors.New("invalid platform")
	}

	domain := strings.ToLower(strings.TrimSpace(request.Domain))
	if !isValidDomain(domain) {
		return errors.New("invalid domain format")
	}

	// Check duplicate domain
	existing, err := s.repository.GetByDomain(domain)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existing != nil {
		return errors.New("website domain already exists")
	}

	website := Website{
		WebsiteName: strings.TrimSpace(request.WebsiteName),
		Domain:      domain,
		Platform:    request.Platform,
		Status:      true,
		CreatedBy:   createdBy,
	}

	return s.repository.Create(&website)
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

	if !isValidPlatform(request.Platform) {
		return errors.New("invalid platform")
	}

	domain := strings.ToLower(strings.TrimSpace(request.Domain))
	if !isValidDomain(domain) {
		return errors.New("invalid domain format")
	}

	// Check duplicate domain for update
	existing, err := s.repository.GetByDomain(domain)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existing != nil && existing.ID != id {
		return errors.New("website domain already exists")
	}

	website.WebsiteName = strings.TrimSpace(request.WebsiteName)
	website.Domain = domain
	website.Platform = request.Platform
	website.Status = request.Status

	return s.repository.Update(website)
}

func (s *websiteService) Delete(id uint) error {
	return s.repository.Delete(id)
}
