package migrations

import (
	"github.com/NathanWasTaken/timely/backend/internal/model"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// InitialSchema represents the initial database schema migration
var InitialSchema = &gormigrate.Migration{
	ID: "202501220001",
	Migrate: func(tx *gorm.DB) error {
		// Create all tables based on current models
		return tx.AutoMigrate(
			&model.User{},
			&model.Account{},
			&model.Calendar{},
			&model.CalendarEvent{},
		)
	},
	Rollback: func(tx *gorm.DB) error {
		// Drop all tables in reverse order to handle foreign key constraints
		return tx.Migrator().DropTable(
			&model.CalendarEvent{},
			&model.Calendar{},
			&model.Account{},
			&model.User{},
		)
	},
}