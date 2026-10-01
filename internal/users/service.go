package users

import (
	"errors"

	"popup-manager-api/utils"
)

type UserService struct {
	Repository *UserRepository
}

func NewUserService(repository *UserRepository) *UserService {
	return &UserService{
		Repository: repository,
	}
}

func (s *UserService) Login(email, password string) (*User, error) {

	user, err := s.Repository.FindByEmail(email)

	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !utils.CheckPassword(password, user.Password) {
		return nil, errors.New("invalid email or password")
	}

	if !user.Status {
		return nil, errors.New("user account is disabled")
	}

	return user, nil
}
