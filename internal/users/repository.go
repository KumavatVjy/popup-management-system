package users

import "gorm.io/gorm"

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}

}

func (r *UserRepository) FindByEmail(email string) (*User, error) {

	var user User

	err := r.DB.Where("email = ?", email).First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}
