package database

import (
	"log/slog"
	"os"

	"popup-manager-api/config"
	"popup-manager-api/internal/users"
	"popup-manager-api/utils"
)

func SeedDatabase() {

	var count int64

	DB.Model(&users.User{}).Count(&count)

	if count > 0 {
		slog.Info("users already exist, seeder skipped")
		return
	}

	adminEmail := config.AppConfig.AdminEmail
	if adminEmail == "" {
		adminEmail = "admin@example.com"
	}

	adminPassword := config.AppConfig.AdminPassword
	if adminPassword == "" {
		slog.Info("admin seed skipped", "reason", "ADMIN_PASSWORD not configured")
		return
	}

	password, err := utils.HashPassword(adminPassword)
	if err != nil {
		slog.Error("failed to hash admin password", "error", err)
		os.Exit(1)
	}

	admin := users.User{
		Name:     "Administrator",
		Email:    adminEmail,
		Password: password,
		Role:     "Super Admin",
		Status:   true,
	}

	if err := DB.Create(&admin).Error; err != nil {
		slog.Error("failed to create default administrator", "error", err)
		return
	}

	slog.Info("admin user created", "email", adminEmail)
}
