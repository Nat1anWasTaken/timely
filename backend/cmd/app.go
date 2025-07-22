package cmd

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/NathanWasTaken/timely/backend/internal/config"
	"github.com/NathanWasTaken/timely/backend/internal/model"
	"github.com/NathanWasTaken/timely/backend/internal/router"
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
	r := chi.NewRouter()

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file: " + err.Error())
	}

	// Initialize Snowflake ID generator
	utils.InitSnowflake(1) // Node ID 1 for this instance
	log.Println("Snowflake ID generator initialized")

	logger.Init(os.Getenv("GO_ENV"))

	DatabaseInit()

	log.Println("Database connected and migrated successfully")

	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE", "HEAD", "OPTION"},
		AllowedHeaders:   []string{"User-Agent", "Content-Type", "Accept", "Accept-Encoding", "Accept-Language", "Cache-Control", "Connection", "DNT", "Host", "Origin", "Pragma", "Referer"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Health check endpoint
	// @Summary Health Check
	// @Description Check if the API service is running
	// @Tags System
	// @Produce plain
	// @Success 200 {string} string "OK"
	// @Router /health [get]
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Home endpoint
	// @Summary API Information
	// @Description Get basic API information
	// @Tags System
	// @Produce plain
	// @Success 200 {string} string "Hello World!"
	// @Router / [get]
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	// Swagger documentation
	if os.Getenv("GO_ENV") == "development" {
		r.Get("/swagger/*", httpSwagger.WrapHandler)
	}

	r.Route("/api", func(r chi.Router) {
		router.AuthRouter(r)
		router.CalendarRouter(r)
	})
	utils.PrintLogo()

	http.ListenAndServe(":8000", r)
}

func DatabaseInit() {
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
}
