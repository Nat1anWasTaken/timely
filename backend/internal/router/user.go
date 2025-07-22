package router

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/NathanWasTaken/timely/backend/internal/config"
	"github.com/NathanWasTaken/timely/backend/internal/handler/user"
	"github.com/NathanWasTaken/timely/backend/internal/middleware"
	"github.com/NathanWasTaken/timely/backend/internal/repository"
	"github.com/NathanWasTaken/timely/backend/internal/service"
)

func UserRouter(r chi.Router) {
	// Initialize database dependencies
	dbConfig := config.NewDatabaseConfig()
	userRepo := repository.NewUserRepository(dbConfig.GetDB())

	// Initialize services
	userService := service.NewUserService(userRepo)

	// Initialize handlers
	userHandler := user.NewUserHandler(userService)

	// User routes with JWT middleware
	r.Route("/user", func(r chi.Router) {
		// Apply JWT middleware to all user routes
		r.Use(middleware.JWTMiddleware(zap.L()))

		// RESTful user profile endpoint
		r.Get("/profile", userHandler.GetProfile)
	})
}