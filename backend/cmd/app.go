package cmd

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/NathanWasTaken/timely/backend/pkg/logger"
	"github.com/NathanWasTaken/timely/backend/pkg/utils"

	_ "github.com/NathanWasTaken/timely/backend/docs"
)

// @title Timely Backend API
// @version 1.0
// @description A comprehensive authentication service with traditional email/password and Google OAuth support
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8000
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func Run() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file: " + err.Error())
	}

	// Initialize core systems
	initializeCoreSystem()

	// Initialize database
	InitializeDatabase()

	// Setup HTTP router
	router := SetupRouter()

	// Setup graceful shutdown
	SetupGracefulShutdown(nil)

	// Start the server
	StartServer(router)
}

// initializeCoreSystem initializes core system components
func initializeCoreSystem() {
	// Initialize Snowflake ID generator
	utils.InitSnowflake(1) // Node ID 1 for this instance
	log.Println("Snowflake ID generator initialized")

	// Initialize logger
	logger.Init(os.Getenv("GO_ENV"))
}