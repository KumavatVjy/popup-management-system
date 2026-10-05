package database

import (
	"log/slog"

	"popup-manager-api/internal/popups"
	"popup-manager-api/internal/users"
	"popup-manager-api/internal/websites"
)

func MigrateDatabase() error {

	err := DB.AutoMigrate(
		&users.User{},
		&websites.Website{},
		&popups.Popup{},
	)

	if err != nil {
		return err
	}

	slog.Info("database migration completed")
	return nil
}
