package main

import (
	"log"
	"log/slog"
	"os"

	"popup-manager-api/config"
	"popup-manager-api/database"
	"popup-manager-api/internal/common"
	"popup-manager-api/internal/popups"
	"popup-manager-api/internal/users"
	"popup-manager-api/internal/websites"
	"popup-manager-api/routes"

	"github.com/gin-gonic/gin"
)

func main() { // entry point like index.php

	if err := config.Load(); err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
	}

	common.InitLogger(config.AppConfig.AppEnv)

	if err := database.ConnectDatabase(); err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}

	if err := database.MigrateDatabase(); err != nil {
		slog.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	if err := database.SeedDatabase(); err != nil {
		slog.Error("database seeding failed", "error", err)
		os.Exit(1)
	}
	router := gin.Default() // create HTTP server , router receive every incoming request

	// Users Module
	userRepository := users.NewUserRepository(database.DB)
	userService := users.NewUserService(userRepository)
	userController := users.NewUserController(userService)

	// Website Module
	websiteRepository := websites.NewWebsiteRepository(database.DB)
	websiteService := websites.NewWebsiteService(websiteRepository)
	websiteController := websites.NewWebsiteController(websiteService)

	// Popup Module
	popupRepository := popups.NewPopupRepository(database.DB)
	popupService := popups.NewPopupService(popupRepository, websiteRepository)
	popupController := popups.NewPopupController(popupService)

	controllers := &routes.Controllers{
		User:    userController,
		Website: websiteController,
		Popup:   popupController,
	}

	routes.SetupRoutes(router, controllers)

	slog.Info("server started", "port", config.AppConfig.AppPort, "env", config.AppConfig.AppEnv)
	router.Run(":" + config.AppConfig.AppPort) // this start the server
}
