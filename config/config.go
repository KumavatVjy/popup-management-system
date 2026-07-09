package config

import (
	"log"
	"os"
	"github.com/joho/godotenv"
)
// Loads the .env file into memory.
func LoadEnv(){
	if err:= godotenv.Load(); 
	err != nil{
		log.Fatal("Error loading .env file")
	}
}

// Retrieve values anywhere in the application
func GetEnv(key string) string {
	return os.Getenv(key)
}


