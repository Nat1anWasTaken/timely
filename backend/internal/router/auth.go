package router

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/NathanWasTaken/timely/backend/internal/config"
	"github.com/NathanWasTaken/timely/backend/internal/handler/auth"
	"github.com/NathanWasTaken/timely/backend/internal/middleware"
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
	oauthService, err := service.NewOAuthService(oauthConfig, userService)
	if err != nil {
		panic(fmt.Sprintf("Failed to create OAuth service: %v", err))
	}
	googleHandler := auth.NewGoogleOAuthHandler(oauthService, userService)

	// Initialize traditional auth handlers
	loginHandler := auth.NewLoginHandler(userService)
	registerHandler := auth.NewRegisterHandler(userService)
	logoutHandler := auth.NewLogoutHandler()

	// Public Routes
	r.Route("/auth", func(r chi.Router) {
		// Optional JWT middleware - must be defined before routes
		r.Use(middleware.OptionalJWTMiddleware(zap.L()))

		// Traditional auth endpoints
		r.Post("/login", loginHandler.Login)
		r.Post("/register", registerHandler.Register) // Handle registration
		r.Post("/logout", logoutHandler.Logout)

		// Google OAuth endpoints
		r.Get("/google/login", googleHandler.GoogleLogin)
		r.Get("/google/callback", googleHandler.GoogleCallback)
	})
}
