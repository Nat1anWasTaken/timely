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
	calendarRepo := repository.NewCalendarRepository(dbConfig.GetDB())

	// Initialize OAuth dependencies
	oauthConfig := config.NewOAuthConfig()

	// Initialize services
	userService := service.NewUserService(userRepo)
	calendarService := service.NewCalendarService(userRepo, calendarRepo, oauthConfig)

	// Initialize handlers
	userHandler := user.NewUserHandler(userService)
	userEventsHandler := user.NewUserEventsHandler(calendarService, userService)

	// User routes
	r.Route("/users", func(r chi.Router) {
		// Authenticated user profile endpoint (requires JWT)
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTMiddleware(zap.L()))
			r.Get("/me", userHandler.GetProfile)
		})

		// Public endpoints (no authentication required)
		r.Get("/{username}", userHandler.GetPublicProfile)
		r.Get("/{username}/events", userEventsHandler.GetPublicUserEvents)
	})
}
