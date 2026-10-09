package popups

import (
	"errors"
	"strings"
	"time"

	"popup-manager-api/internal/common"
	"popup-manager-api/internal/websites"
)

type PopupService interface {
	Create(request CreatePopupRequest, createdBy uint) error
	GetAll(params common.QueryParams) ([]Popup, int64, error)
	GetByID(id uint) (*Popup, error)
	GetByWebsiteID(websiteID uint, params common.QueryParams) ([]Popup, int64, error)
	GetPublicPopups(websiteKey string) ([]Popup, error)
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
	vErr := common.NewValidationError()

	title := strings.TrimSpace(request.Title)
	if title == "" {
		vErr.Add("title", "title cannot be empty")
	}

	content := strings.TrimSpace(request.Content)
	if content == "" {
		vErr.Add("content", "content cannot be empty")
	}

	if !isValidPosition(request.Position) {
		vErr.Add("position", "invalid popup position")
	}

	if request.StartTime != nil && request.EndTime != nil {
		if !request.StartTime.Before(*request.EndTime) {
			vErr.Add("start_time", "start time must be before end time")
		}
	}

	if vErr.HasErrors() {
		return vErr
	}

	_, err := s.websiteRepository.GetByID(request.WebsiteID)
	if err != nil {
		return common.ErrNotFound
	}

	popup := Popup{
		WebsiteID: request.WebsiteID,
		Title:     title,
		Content:   content,
		Position:  request.Position,
		Status:    request.Status,
		StartTime: request.StartTime,
		EndTime:   request.EndTime,
		CreatedBy: createdBy,
	}

	return s.repository.Create(&popup)
}

func (s *popupService) GetAll(params common.QueryParams) ([]Popup, int64, error) {
	return s.repository.GetAll(params)
}

func (s *popupService) GetByID(id uint) (*Popup, error) {
	return s.repository.GetByID(id)
}

func (s *popupService) GetByWebsiteID(websiteID uint, params common.QueryParams) ([]Popup, int64, error) {
	return s.repository.GetByWebsiteID(websiteID, params)
}

func (s *popupService) GetPublicPopups(websiteKey string) ([]Popup, error) {
	websiteKey = strings.TrimSpace(websiteKey)
	if websiteKey == "" {
		return nil, errors.New("website_key is required")
	}

	website, err := s.websiteRepository.GetByKey(websiteKey)
	if err != nil {
		return nil, err
	}

	return s.repository.GetEligiblePublicByWebsite(website.ID, time.Now())
}

func (s *popupService) Update(id uint, request UpdatePopupRequest) error {
	popup, err := s.repository.GetByID(id)
	if err != nil {
		return err
	}

	vErr := common.NewValidationError()

	title := strings.TrimSpace(request.Title)
	if title == "" {
		vErr.Add("title", "title cannot be empty")
	}

	content := strings.TrimSpace(request.Content)
	if content == "" {
		vErr.Add("content", "content cannot be empty")
	}

	if !isValidPosition(request.Position) {
		vErr.Add("position", "invalid popup position")
	}

	if request.StartTime != nil && request.EndTime != nil {
		if !request.StartTime.Before(*request.EndTime) {
			vErr.Add("start_time", "start time must be before end time")
		}
	}

	if vErr.HasErrors() {
		return vErr
	}

	popup.Title = title
	popup.Content = content
	popup.Position = request.Position
	popup.Status = request.Status
	popup.StartTime = request.StartTime
	popup.EndTime = request.EndTime

	return s.repository.Update(popup)
}

func (s *popupService) Delete(id uint) error {
	return s.repository.Delete(id)
}
