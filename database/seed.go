package database

import (
	"log"

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

	password, err := utils.HashPassword("Admin@123")

	if err != nil {
		log.Fatal(err)
	}

	admin := users.User{
		Name:     "Administrator",
		Email:    "vijay.kumavat@definedge.com",
		Password: password,
		Role:     "Super Admin",
		Status:   true,
	}

	DB.Create(&admin)

	log.Println("Default administrator created successfully.")
}
