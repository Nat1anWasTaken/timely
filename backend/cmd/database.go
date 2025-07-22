package cmd

import (
	"log"

	"github.com/NathanWasTaken/timely/backend/internal/config"
	"github.com/NathanWasTaken/timely/backend/internal/model"
)

// InitializeDatabase sets up and migrates the database
func InitializeDatabase() {
	dbConfig := config.NewDatabaseConfig()

	// Auto-migrate the schema
	if err := dbConfig.GetDB().AutoMigrate(
		&model.User{},
		&model.Account{},
		&model.Calendar{},
		&model.CalendarEvent{},
	); err != nil {
		log.Fatal("Failed to migrate database: " + err.Error())
	}

	log.Println("Database connected and migrated successfully")
}