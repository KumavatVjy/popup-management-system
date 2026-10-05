package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName string
	AppEnv  string
	AppPort string

	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	JWTSecret      string
	JWTExpireHours int

	AdminEmail    string
	AdminPassword string
}

var AppConfig *Config

func Load() error {
	_ = godotenv.Load()

	config := &Config{}

	config.AppName = getEnv("APP_NAME", "popup-manager")
	config.AppEnv = getEnv("APP_ENV", "development")

	appPort := getEnv("APP_PORT", "")
	if appPort == "" {
		return fmt.Errorf("APP_PORT is required")
	}
	config.AppPort = appPort

	// Database
	config.DBHost = getEnv("DB_HOST", "")
	if config.DBHost == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	config.DBPort = getEnv("DB_PORT", "")
	if config.DBPort == "" {
		return fmt.Errorf("DB_PORT is required")
	}
	config.DBName = getEnv("DB_NAME", "")
	if config.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	config.DBUser = getEnv("DB_USER", "")
	if config.DBUser == "" {
		return fmt.Errorf("DB_USER is required")
	}
	config.DBPassword = getEnv("DB_PASSWORD", "")
	// Optional depending on local setup, so we do not enforce it

	// JWT
	config.JWTSecret = getEnv("JWT_SECRET", "")
	if config.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	jwtExpireStr := getEnv("JWT_EXPIRE_HOURS", "")
	if jwtExpireStr == "" {
		return fmt.Errorf("JWT_EXPIRE_HOURS is required")
	}
	jwtExpire, err := strconv.Atoi(jwtExpireStr)
	if err != nil {
		return fmt.Errorf("JWT_EXPIRE_HOURS must be a valid integer")
	}
	config.JWTExpireHours = jwtExpire

	// Admin
	config.AdminEmail = getEnv("ADMIN_EMAIL", "")
	config.AdminPassword = getEnv("ADMIN_PASSWORD", "") // Optional

	AppConfig = config
	return nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
