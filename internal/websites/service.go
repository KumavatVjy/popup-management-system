package websites

import (
	"errors"
	"strings"
)

type WebsiteService interface {

	Create(request CreateWebsiteRequest, createdBy uint) error

	GetAll() ([]Website, error)

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


func isValidPlatform(platform string) bool {

	switch platform {

	case PlatformWordPress,
		PlatformReact,
		PlatformPython,
		PlatformLaravel,
		PlatformNodeJS,
		PlatformHTML,
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

	// Check duplicate domain
	existing, _ := s.repository.GetByDomain(strings.ToLower(strings.TrimSpace(request.Domain)))

	if existing != nil {
		return errors.New("website domain already exists")
	}

	website := Website{
		WebsiteName: strings.TrimSpace(request.WebsiteName),
		Domain:      strings.ToLower(strings.TrimSpace(request.Domain)),
		Platform:    request.Platform,
		Status:      true,
		CreatedBy:   createdBy,
	}

	return s.repository.Create(&website)
}

func (s *websiteService) GetAll() ([]Website, error) {

	return s.repository.GetAll()

}

func (s *websiteService) GetByID(id uint) (*Website, error) {

	return s.repository.GetByID(id)

}

func (s *websiteService) Update(id uint, request UpdateWebsiteRequest) error {

	return nil

}

func (s *websiteService) Delete(id uint) error {

	return s.repository.Delete(id)

}