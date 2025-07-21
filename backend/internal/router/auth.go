package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/NathanWasTaken/timely/backend/internal/config"
	"github.com/NathanWasTaken/timely/backend/internal/handler/auth"
	"github.com/NathanWasTaken/timely/backend/internal/repository"
	"github.com/NathanWasTaken/timely/backend/internal/service"
)

func AuthRouter(r chi.Router) {
	// Initialize database dependencies
	dbConfig := config.NewDatabaseConfig()
	userRepo := repository.NewUserRepository(dbConfig.GetDB())
	userService := service.NewUserService(userRepo)

	// Initialize OAuth dependencies
	oauthConfig := config.NewOAuthConfig()
	oauthService := service.NewOAuthService(oauthConfig, userService)
	googleHandler := auth.NewGoogleOAuthHandler(oauthService, userService)

	// Public Routes
	r.Route("/auth", func(r chi.Router) {
		// Traditional auth endpoints
		r.Post("/login", auth.Login)
		r.Get("/register", auth.Register)  // Changed to GET to show available registration options
		r.Post("/register", auth.Register) // Also support POST for traditional registration

		// Google OAuth endpoints
		r.Get("/google/login", googleHandler.GoogleLogin)
		r.Get("/google/callback", googleHandler.GoogleCallback)
	})
}
