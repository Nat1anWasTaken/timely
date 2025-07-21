package config

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseConfig struct {
	DB *gorm.DB
}

func NewDatabaseConfig() *DatabaseConfig {
	// For now, using SQLite for simplicity
	// You can change this to PostgreSQL, MySQL, etc. later
	// TODO: Support other databases
	db, err := gorm.Open(sqlite.Open("timely.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	return &DatabaseConfig{
		DB: db,
	}
}

func (dc *DatabaseConfig) GetDB() *gorm.DB {
	return dc.DB
}
