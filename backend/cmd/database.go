package cmd

import (
	"log"

	"github.com/go-gormigrate/gormigrate/v2"

	"github.com/NathanWasTaken/timely/backend/internal/config"
	"github.com/NathanWasTaken/timely/backend/internal/migrations"
)

// InitializeDatabase sets up and migrates the database
func InitializeDatabase() {
	dbConfig := config.NewDatabaseConfig()
	db := dbConfig.GetDB()

	// Initialize gormigrate with migrations
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		migrations.InitialSchema,
		migrations.AddSyncFields,
	})

	// Run migrations
	if err := m.Migrate(); err != nil {
		log.Fatal("Failed to migrate database: " + err.Error())
	}

	log.Println("Database connected and migrated successfully")
}
