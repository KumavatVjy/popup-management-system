package database

import (
	"errors"
	"log/slog"

	"popup-manager-api/config"
	"popup-manager-api/internal/users"
	"popup-manager-api/utils"

	"gorm.io/gorm"
)

func SeedDatabase() error {
	adminEmail := config.AppConfig.AdminEmail
	if adminEmail == "" {
		adminEmail = "admin@example.com"
	}

	adminPassword := config.AppConfig.AdminPassword
	if adminPassword == "" {
		slog.Info("admin seed skipped", "reason", "ADMIN_PASSWORD not configured")
		return nil
	}

	var existing users.User
	err := DB.Where("email = ?", adminEmail).First(&existing).Error
	if err == nil {
		slog.Info("admin user already exists", "email", adminEmail)
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Admin does not exist: create it
	password, err := utils.HashPassword(adminPassword)
	if err != nil {
		return err
	}

	admin := users.User{
		Name:     "Administrator",
		Email:    adminEmail,
		Password: password,
		Role:     "Super Admin",
		Status:   true,
	}

	if err := DB.Create(&admin).Error; err != nil {
		return err
	}

	slog.Info("admin user created", "email", adminEmail)
	return nil
}
