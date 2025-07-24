package migrations

import (
	"github.com/NathanWasTaken/timely/backend/internal/model"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// AddSyncFields adds sync-related fields to the Calendar table for Google Calendar sync optimization
var AddSyncFields = &gormigrate.Migration{
	ID: "202501240001",
	Migrate: func(tx *gorm.DB) error {
		// Add sync fields to Calendar table
		// This will only add the fields if they don't already exist
		return tx.AutoMigrate(&model.Calendar{})
	},
	Rollback: func(tx *gorm.DB) error {
		// Remove sync fields from Calendar table
		if tx.Migrator().HasColumn(&model.Calendar{}, "sync_status") {
			if err := tx.Migrator().DropColumn(&model.Calendar{}, "sync_status"); err != nil {
				return err
			}
		}
		if tx.Migrator().HasColumn(&model.Calendar{}, "sync_token") {
			if err := tx.Migrator().DropColumn(&model.Calendar{}, "sync_token"); err != nil {
				return err
			}
		}
		if tx.Migrator().HasColumn(&model.Calendar{}, "last_full_sync") {
			if err := tx.Migrator().DropColumn(&model.Calendar{}, "last_full_sync"); err != nil {
				return err
			}
		}
		return nil
	},
}