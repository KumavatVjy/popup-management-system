package database

import (
	"log"

	"popup-manager-api/internal/users"
	"popup-manager-api/internal/websites"
)

func MigrateDatabase() {

	err := DB.AutoMigrate(
		&users.User{},
		&websites.Website{},
	)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database Migration Completed")
}