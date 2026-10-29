package database

import (
	"fmt"
	"log/slog"

	"popup-manager-api/internal/popups"
	"popup-manager-api/internal/users"
	"popup-manager-api/internal/websites"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	// Existing campaigns predate device targeting; keep them visible on every device.
	if err := DB.Model(&popups.Popup{}).
		Where("show_on_desktop = ? AND show_on_tablet = ? AND show_on_mobile = ?", false, false, false).
		Updates(map[string]any{"show_on_desktop": true, "show_on_tablet": true, "show_on_mobile": true}).Error; err != nil {
		return fmt.Errorf("failed to initialize popup device visibility: %w", err)
	}
	if err := DB.Model(&popups.Popup{}).
		Where("show_image = ? AND show_content = ?", false, false).
		Updates(map[string]any{"show_image": true, "show_content": true}).Error; err != nil {
		return fmt.Errorf("failed to initialize popup image and content visibility: %w", err)
	}

	if err := migrateWebsiteOrigins(); err != nil {
		return fmt.Errorf("failed to consolidate website origins: %w", err)
	}

	slog.Info("database migration completed")
	return nil
}

// migrateWebsiteOrigins collapses legacy registrations that stored page paths
// in Domain into one website per host. Popup page targeting is inferred from
// each legacy website path before associations are moved to the canonical row.
func migrateWebsiteOrigins() error {
	var registered []websites.Website
	if err := DB.Order("id asc").Find(&registered).Error; err != nil {
		return err
	}

	groups := make(map[string][]websites.Website)
	for _, website := range registered {
		origin, err := websites.NormalizeAndValidateDomain(website.Domain)
		if err != nil {
			return fmt.Errorf("website %d has invalid domain %q: %w", website.ID, website.Domain, err)
		}
		groups[origin] = append(groups[origin], website)
	}

	for origin, group := range groups {
		primary := group[0]
		canonicalRowFound := false
		anyActive := false
		for _, candidate := range group {
			if candidate.Status {
				anyActive = true
			}
			if candidate.Domain == origin {
				primary = candidate
				canonicalRowFound = true
			}
		}
		if !canonicalRowFound && anyActive {
			for _, candidate := range group {
				if candidate.Status {
					primary = candidate
					break
				}
			}
		}

		if err := DB.Transaction(func(tx *gorm.DB) error {
			if anyActive && !primary.Status {
				if err := tx.Model(&websites.Website{}).Where("id = ?", primary.ID).Update("status", true).Error; err != nil {
					return err
				}
			}
			if primary.Domain != origin {
				if err := tx.Model(&websites.Website{}).Where("id = ?", primary.ID).Update("domain", origin).Error; err != nil {
					return err
				}
			}

			websiteIDs := make([]uint, 0, len(group))
			for _, website := range group {
				websiteIDs = append(websiteIDs, website.ID)
			}

			// Preserve legacy page scopes for campaigns without explicit targeting.
			var associatedPopups []popups.Popup
			if err := tx.Where("website_id IN ?", websiteIDs).Find(&associatedPopups).Error; err != nil {
				return err
			}
			popupPaths := make(map[uint][]string)
			popupHasRootScope := make(map[uint]bool)
			popupByID := make(map[uint]popups.Popup, len(associatedPopups))
			for _, website := range group {
				path := websites.ExtractWebsitePath(website.Domain)
				for _, popup := range associatedPopups {
					if popup.WebsiteID != website.ID {
						continue
					}
					popupByID[popup.ID] = popup
					if path == "" {
						popupHasRootScope[popup.ID] = true
					} else {
						popupPaths[popup.ID] = append(popupPaths[popup.ID], path)
					}
				}
			}

			if tx.Migrator().HasTable("popup_websites") {
				type popupWebsiteAssociation struct {
					PopupID   uint `gorm:"column:popup_id"`
					WebsiteID uint `gorm:"column:website_id"`
				}
				var associations []popupWebsiteAssociation
				if err := tx.Table("popup_websites").Where("website_id IN ?", websiteIDs).Find(&associations).Error; err != nil {
					return err
				}
				for _, association := range associations {
					path := ""
					for _, website := range group {
						if website.ID == association.WebsiteID {
							path = websites.ExtractWebsitePath(website.Domain)
							break
						}
					}
					if path == "" {
						popupHasRootScope[association.PopupID] = true
					} else {
						popupPaths[association.PopupID] = append(popupPaths[association.PopupID], path)
					}
					if _, exists := popupByID[association.PopupID]; !exists {
						var popup popups.Popup
						if err := tx.First(&popup, association.PopupID).Error; err != nil {
							return err
						}
						popupByID[association.PopupID] = popup
					}
				}

				for _, association := range associations {
					row := popupWebsiteAssociation{PopupID: association.PopupID, WebsiteID: primary.ID}
					if err := tx.Table("popup_websites").Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
						return err
					}
				}
				if err := tx.Exec("DELETE FROM popup_websites WHERE website_id IN ? AND website_id <> ?", websiteIDs, primary.ID).Error; err != nil {
					return err
				}
			}

			associatedPopups = associatedPopups[:0]
			for _, popup := range popupByID {
				associatedPopups = append(associatedPopups, popup)
			}
			for _, popup := range associatedPopups {
				if popup.TargetPages != "" && popup.TargetPages != "all" {
					continue
				}
				if popupHasRootScope[popup.ID] || len(popupPaths[popup.ID]) == 0 {
					continue
				}
				uniquePaths := make([]string, 0, len(popupPaths[popup.ID]))
				seen := make(map[string]bool)
				for _, path := range popupPaths[popup.ID] {
					if !seen[path] {
						uniquePaths = append(uniquePaths, path)
						seen[path] = true
					}
				}
				if err := tx.Model(&popups.Popup{}).Where("id = ?", popup.ID).Update("target_pages", strings.Join(uniquePaths, ", ")).Error; err != nil {
					return err
				}
			}

			if len(group) > 1 {
				duplicateIDs := make([]uint, 0, len(group)-1)
				for _, website := range group {
					if website.ID != primary.ID {
						duplicateIDs = append(duplicateIDs, website.ID)
					}
				}
				if err := tx.Model(&popups.Popup{}).Where("website_id IN ?", duplicateIDs).Update("website_id", primary.ID).Error; err != nil {
					return err
				}
				if err := tx.Where("id IN ?", duplicateIDs).Delete(&websites.Website{}).Error; err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
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
