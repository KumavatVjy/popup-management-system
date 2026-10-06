package database

import (
	"fmt"
	"log/slog"

	"popup-manager-api/internal/popups"
	"popup-manager-api/internal/users"
	"popup-manager-api/internal/websites"
)

func MigrateDatabase() error {

	if err := migrateWebsitesKey(); err != nil {
		return fmt.Errorf("failed to migrate website keys: %w", err)
	}

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

func migrateWebsitesKey() error {
	// If websites table does not exist yet (e.g. fresh database), AutoMigrate will create it
	if !DB.Migrator().HasTable("websites") {
		return nil
	}

	// 1. Ensure website_key column exists (added initially as nullable to prevent duplicate empty string collisions on existing rows)
	hasColumn := DB.Migrator().HasColumn(&websites.Website{}, "WebsiteKey") || DB.Migrator().HasColumn(&websites.Website{}, "website_key")
	if !hasColumn {
		if err := DB.Exec("ALTER TABLE websites ADD COLUMN website_key VARCHAR(64) NULL").Error; err != nil {
			return fmt.Errorf("failed to add website_key column: %w", err)
		}
		slog.Info("added website_key column to websites table")
	}

	// 2. Backfill existing records that have empty or null website_key
	var unkeyedWebsites []websites.Website
	if err := DB.Unscoped().Where("website_key IS NULL OR website_key = ''").Find(&unkeyedWebsites).Error; err != nil {
		return fmt.Errorf("failed to query unkeyed websites: %w", err)
	}

	if len(unkeyedWebsites) > 0 {
		slog.Info("backfilling website_key for existing websites", "count", len(unkeyedWebsites))
		for _, w := range unkeyedWebsites {
			key, err := websites.GenerateWebsiteKey()
			if err != nil {
				return fmt.Errorf("failed to generate key for website %d: %w", w.ID, err)
			}
			if err := DB.Model(&websites.Website{}).Unscoped().Where("id = ?", w.ID).Update("website_key", key).Error; err != nil {
				return fmt.Errorf("failed to backfill website %d key: %w", w.ID, err)
			}
			slog.Info("backfilled website_key for website", "website_id", w.ID, "domain", w.Domain)
		}
	}

	// 3. Verify safety check: ensure no record has empty/null website_key
	var emptyCount int64
	if err := DB.Model(&websites.Website{}).Unscoped().Where("website_key IS NULL OR website_key = ''").Count(&emptyCount).Error; err != nil {
		return fmt.Errorf("failed to count unkeyed websites: %w", err)
	}
	if emptyCount > 0 {
		return fmt.Errorf("safety check failed: %d websites still have empty website_key", emptyCount)
	}

	// 4. Ensure NOT NULL constraint is applied
	if err := DB.Exec("ALTER TABLE websites MODIFY COLUMN website_key VARCHAR(64) NOT NULL").Error; err != nil {
		return fmt.Errorf("failed to set website_key to NOT NULL: %w", err)
	}

	return nil
}
