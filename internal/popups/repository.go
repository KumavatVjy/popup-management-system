package popups

import (
	"popup-manager-api/internal/common"

	"gorm.io/gorm"
)

type PopupRepository interface {
	Create(popup *Popup) error
	GetAll(params common.QueryParams) ([]Popup, int64, error)
	GetByID(id uint) (*Popup, error)
	GetByWebsiteID(websiteID uint, params common.QueryParams) ([]Popup, int64, error)
	Update(popup *Popup) error
	Delete(id uint) error
}

type popupRepository struct {
	db *gorm.DB
}

func NewPopupRepository(db *gorm.DB) PopupRepository {
	return &popupRepository{
		db: db,
	}
}

func (r *popupRepository) Create(popup *Popup) error {
	return r.db.Create(popup).Error
}

func (r *popupRepository) GetAll(params common.QueryParams) ([]Popup, int64, error) {
	var popups []Popup
	var total int64

	db := r.db.Model(&Popup{})

	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		db = db.Where("title LIKE ? OR content LIKE ?", searchTerm, searchTerm)
	}

	db.Count(&total)

	offset := (params.Page - 1) * params.Limit
	err := db.Order(params.Sort + " " + params.Order).Offset(offset).Limit(params.Limit).Find(&popups).Error

	return popups, total, err
}

func (r *popupRepository) GetByID(id uint) (*Popup, error) {
	var popup Popup
	err := r.db.First(&popup, id).Error
	if err != nil {
		return nil, err
	}
	return &popup, nil
}

func (r *popupRepository) GetByWebsiteID(websiteID uint, params common.QueryParams) ([]Popup, int64, error) {
	var popups []Popup
	var total int64

	db := r.db.Model(&Popup{}).Where("website_id = ?", websiteID)

	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		db = db.Where("title LIKE ? OR content LIKE ?", searchTerm, searchTerm)
	}

	db.Count(&total)

	offset := (params.Page - 1) * params.Limit
	err := db.Order(params.Sort + " " + params.Order).Offset(offset).Limit(params.Limit).Find(&popups).Error

	return popups, total, err
}

func (r *popupRepository) Update(popup *Popup) error {
	return r.db.Save(popup).Error
}

func (r *popupRepository) Delete(id uint) error {
	return r.db.Delete(&Popup{}, id).Error
}
