package main

import (
	"popup-manager-api/config"
	"popup-manager-api/database"
	"popup-manager-api/internal/users"
	"popup-manager-api/routes"

	"popup-manager-api/internal/websites"


	"github.com/gin-gonic/gin"
)

func main() {  // entry point like index.php

	config.LoadEnv() // 
    database.ConnectDatabase()
	database.MigrateDatabase()
	database.SeedDatabase()
	router := gin.Default() // create HTTP server , router receive every incoming request

	// Users Module
	userRepository := users.NewUserRepository(database.DB)
	userService := users.NewUserService(userRepository)
	userController := users.NewUserController(userService)

	// Website Module
	websiteRepository := websites.NewWebsiteRepository(database.DB)
	websiteService := websites.NewWebsiteService(websiteRepository)
	websiteController := websites.NewWebsiteController(websiteService)

	controllers := &routes.Controllers{
		User: userController,
		Website:  websiteController,
	}

	routes.SetupRoutes(router, controllers)
	router.Run(":" + config.GetEnv("APP_PORT")) // this start the server
}
