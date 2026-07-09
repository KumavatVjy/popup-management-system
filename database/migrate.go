package database

import (
	"log"

	"popup-manager-api/internal/users"
)

func MigrateDatabase() {

	err := DB.AutoMigrate(
		&users.User{},
	)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database Migration Completed")
}