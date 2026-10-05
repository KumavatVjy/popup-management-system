package database

import (
	"log/slog"
	"os"

	"popup-manager-api/internal/popups"
	"popup-manager-api/internal/users"
	"popup-manager-api/internal/websites"
)

func MigrateDatabase() {

	err := DB.AutoMigrate(
		&users.User{},
		&websites.Website{},
		&popups.Popup{},
	)

	if err != nil {
		slog.Error("database migration failed", "error", err)
		os.Exit(1)
	}

	slog.Info("database migration completed")
}
