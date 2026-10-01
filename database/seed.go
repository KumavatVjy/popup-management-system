package database

import (
	"log"

	"popup-manager-api/config"
	"popup-manager-api/internal/users"
	"popup-manager-api/utils"
)

func SeedDatabase() {

	var count int64

	DB.Model(&users.User{}).Count(&count)

	if count > 0 {
		log.Println("Users already exist. Seeder skipped.")
		return
	}

	adminEmail := config.GetEnv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@example.com"
	}

	adminPassword := config.GetEnv("ADMIN_PASSWORD")
	if adminPassword == "" {
		log.Println("ADMIN_PASSWORD not set. Seeder skipped.")
		return
	}

	password, err := utils.HashPassword(adminPassword)
	if err != nil {
		log.Fatal(err)
	}

	admin := users.User{
		Name:     "Administrator",
		Email:    adminEmail,
		Password: password,
		Role:     "Super Admin",
		Status:   true,
	}

	if err := DB.Create(&admin).Error; err != nil {
		log.Println("Failed to create default administrator:", err)
		return
	}

	log.Println("Default administrator created successfully.")
}
