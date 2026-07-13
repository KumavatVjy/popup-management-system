package websites

import "gorm.io/gorm"

type WebsiteRepository interface {

	Create(website *Website) error

	GetAll() ([]Website, error)

	GetByID(id uint) (*Website, error)

	GetByDomain(domain string) (*Website, error)

	Update(website *Website) error

	Delete(id uint) error
}

type websiteRepository struct {
	db *gorm.DB
}

func NewWebsiteRepository(db *gorm.DB) WebsiteRepository {
	return &websiteRepository{
		db: db,
	}
}

// CREATE 
func (r *websiteRepository) Create(website *Website) error {

	return r.db.Create(website).Error

}

// GET ALL WEBSITES
func (r *websiteRepository) GetAll() ([]Website, error) {

	var websites []Website

	err := r.db.Find(&websites).Error

	return websites, err

}

//GET WEBSITE BY ID
func (r *websiteRepository) GetByID(id uint) (*Website, error) {

	var website Website

	err := r.db.First(&website, id).Error

	if err != nil {
		return nil, err
	}

	return &website, nil

}

// GET BY DOMAIN
func (r *websiteRepository) GetByDomain(domain string) (*Website, error) {

	var website Website

	err := r.db.
		Where("domain = ?", domain).
		First(&website).Error

	if err != nil {
		return nil, err
	}

	return &website, nil
}

//UPDATE WEBSITE
func (r *websiteRepository) Update(website *Website) error {

	return r.db.Save(website).Error

}

//DELETE WEBSITE
func (r *websiteRepository) Delete(id uint) error {

	return r.db.Delete(&Website{}, id).Error

}