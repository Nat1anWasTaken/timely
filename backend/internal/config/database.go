package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseConfig struct {
	DB *gorm.DB
}

func NewDatabaseConfig() *DatabaseConfig {
	dbType := strings.ToLower(os.Getenv("DB_TYPE"))
	if dbType == "" {
		dbType = "sqlite"
	}

	var dialector gorm.Dialector
	var err error

	switch dbType {
	case "sqlite":
		dialector = sqlite.Open("timely.db")
	case "postgres":
		dsn := buildPostgresDSN()
		dialector = postgres.Open(dsn)
	case "mysql":
		dsn := buildMySQLDSN()
		dialector = mysql.Open(dsn)
	default:
		log.Fatal("Unsupported database type: " + dbType + ". Supported types: sqlite, postgres, mysql")
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Printf("Successfully connected to %s database", dbType)

	return &DatabaseConfig{
		DB: db,
	}
}

func buildPostgresDSN() string {
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	dbname := getEnvOrDefault("DB_NAME", "timely")
	user := getEnvOrDefault("DB_USER", "timely_user")
	password := os.Getenv("DB_PASSWORD")
	sslmode := getEnvOrDefault("DB_SSL_MODE", "disable")

	if password == "" {
		log.Fatal("DB_PASSWORD is required for PostgreSQL connection")
	}

	return fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		host, port, dbname, user, password, sslmode)
}

func buildMySQLDSN() string {
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "3306")
	dbname := getEnvOrDefault("DB_NAME", "timely")
	user := getEnvOrDefault("DB_USER", "timely_user")
	password := os.Getenv("DB_PASSWORD")

	if password == "" {
		log.Fatal("DB_PASSWORD is required for MySQL connection")
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbname)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (dc *DatabaseConfig) GetDB() *gorm.DB {
	return dc.DB
}
