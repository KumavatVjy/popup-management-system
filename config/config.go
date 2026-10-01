package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

// Loads the .env file into memory.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

// Retrieve values anywhere in the application
func GetEnv(key string) string {
	return os.Getenv(key)
}
