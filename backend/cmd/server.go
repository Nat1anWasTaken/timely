package cmd

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/NathanWasTaken/timely/backend/internal/router"
	"github.com/NathanWasTaken/timely/backend/pkg/utils"
)

// SetupRouter configures and returns the HTTP router
func SetupRouter() *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE", "HEAD", "OPTION"},
		AllowedHeaders:   []string{"User-Agent", "Content-Type", "Accept", "Accept-Encoding", "Accept-Language", "Cache-Control", "Connection", "DNT", "Host", "Origin", "Pragma", "Referer"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// System endpoints
	setupSystemRoutes(r)

	// API routes
	r.Route("/api", func(r chi.Router) {
		router.AuthRouter(r)
		router.CalendarRouter(r)
		router.UserRouter(r)
	})

	return r
}

// setupSystemRoutes configures system-level routes
func setupSystemRoutes(r *chi.Mux) {
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

	// Swagger documentation (development only)
	if os.Getenv("GO_ENV") == "development" {
		r.Get("/swagger/*", httpSwagger.WrapHandler)
	}
}

// SetupGracefulShutdown configures graceful shutdown handling
func SetupGracefulShutdown(cleanup func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down gracefully...")
		
		// Run cleanup function
		if cleanup != nil {
			cleanup()
		}
		
		os.Exit(0)
	}()
}

// StartServer starts the HTTP server
func StartServer(router *chi.Mux) {
	utils.PrintLogo()
	log.Println("Server starting on port 8000...")
	
	if err := http.ListenAndServe(":8000", router); err != nil {
		log.Fatal("Server failed to start: " + err.Error())
	}
}