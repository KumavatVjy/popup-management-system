package popups

import (
	"errors"
	"strings"

	"popup-manager-api/internal/websites"
)

type PopupService interface {
	Create(request CreatePopupRequest, createdBy uint) error
	GetAll() ([]Popup, error)
	GetByID(id uint) (*Popup, error)
	GetByWebsiteID(websiteID uint) ([]Popup, error)
	Update(id uint, request UpdatePopupRequest) error
	Delete(id uint) error
}

type popupService struct {
	repository        PopupRepository
	websiteRepository websites.WebsiteRepository
}

func NewPopupService(repository PopupRepository, websiteRepository websites.WebsiteRepository) PopupService {
	return &popupService{
		repository:        repository,
		websiteRepository: websiteRepository,
	}
}

func isValidPosition(position string) bool {
	switch position {
	case PositionCenter, PositionTopBanner, PositionBottomLeft, PositionBottomRight:
		return true
	default:
		return false
	}
}

func (s *popupService) Create(request CreatePopupRequest, createdBy uint) error {
	if !isValidPosition(request.Position) {
		return errors.New("invalid popup position")
	}

	_, err := s.websiteRepository.GetByID(request.WebsiteID)
	if err != nil {
		return errors.New("website not found")
	}

	popup := Popup{
		WebsiteID: request.WebsiteID,
		Title:     strings.TrimSpace(request.Title),
		Content:   strings.TrimSpace(request.Content),
		Position:  request.Position,
		Status:    request.Status,
		StartTime: request.StartTime,
		EndTime:   request.EndTime,
		CreatedBy: createdBy,
	}

	return s.repository.Create(&popup)
}

func (s *popupService) GetAll() ([]Popup, error) {
	return s.repository.GetAll()
}

func (s *popupService) GetByID(id uint) (*Popup, error) {
	return s.repository.GetByID(id)
}

func (s *popupService) GetByWebsiteID(websiteID uint) ([]Popup, error) {
	return s.repository.GetByWebsiteID(websiteID)
}

func (s *popupService) Update(id uint, request UpdatePopupRequest) error {
	popup, err := s.repository.GetByID(id)
	if err != nil {
		return err
	}

	if !isValidPosition(request.Position) {
		return errors.New("invalid popup position")
	}

	popup.Title = strings.TrimSpace(request.Title)
	popup.Content = strings.TrimSpace(request.Content)
	popup.Position = request.Position
	popup.Status = request.Status
	popup.StartTime = request.StartTime
	popup.EndTime = request.EndTime

	return s.repository.Update(popup)
}

func (s *popupService) Delete(id uint) error {
	return s.repository.Delete(id)
}
