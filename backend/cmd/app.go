package cmd

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"github.com/NathanWasTaken/timely/backend/internal/config"
	"github.com/NathanWasTaken/timely/backend/internal/model"
	"github.com/NathanWasTaken/timely/backend/internal/router"
	"github.com/NathanWasTaken/timely/backend/pkg/utils"
)

func Run() {
	r := chi.NewRouter()

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file: " + err.Error())
	}

	// Initialize Snowflake ID generator
	utils.InitSnowflake(1) // Node ID 1 for this instance
	log.Println("Snowflake ID generator initialized")

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

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	r.Route("/api", func(r chi.Router) {
		router.AuthRouter(r)
	})

	http.ListenAndServe(":8000", r)
}

func DatabaseInit() {
	dbConfig := config.NewDatabaseConfig()

	// Auto-migrate the schema
	if err := dbConfig.GetDB().AutoMigrate(&model.User{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}
