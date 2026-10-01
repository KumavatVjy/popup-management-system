package popups

import "gorm.io/gorm"

type PopupRepository interface {
	Create(popup *Popup) error
	GetAll() ([]Popup, error)
	GetByID(id uint) (*Popup, error)
	GetByWebsiteID(websiteID uint) ([]Popup, error)
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

func (r *popupRepository) GetAll() ([]Popup, error) {
	var popups []Popup
	err := r.db.Find(&popups).Error
	return popups, err
}

func (r *popupRepository) GetByID(id uint) (*Popup, error) {
	var popup Popup
	err := r.db.First(&popup, id).Error
	if err != nil {
		return nil, err
	}
	return &popup, nil
}

func (r *popupRepository) GetByWebsiteID(websiteID uint) ([]Popup, error) {
	var popups []Popup
	err := r.db.Where("website_id = ?", websiteID).Find(&popups).Error
	return popups, err
}

func (r *popupRepository) Update(popup *Popup) error {
	return r.db.Save(popup).Error
}

func (r *popupRepository) Delete(id uint) error {
	return r.db.Delete(&Popup{}, id).Error
}
