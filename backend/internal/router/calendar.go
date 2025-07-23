package router

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/NathanWasTaken/timely/backend/internal/config"
	"github.com/NathanWasTaken/timely/backend/internal/handler/calendar"
	"github.com/NathanWasTaken/timely/backend/internal/middleware"
	"github.com/NathanWasTaken/timely/backend/internal/repository"
	"github.com/NathanWasTaken/timely/backend/internal/service"
)

func CalendarRouter(r chi.Router) {
	// Initialize database dependencies
	dbConfig := config.NewDatabaseConfig()
	userRepo := repository.NewUserRepository(dbConfig.GetDB())
	calendarRepo := repository.NewCalendarRepository(dbConfig.GetDB())

	// Initialize OAuth dependencies
	oauthConfig := config.NewOAuthConfig()

	// Initialize services
	calendarService := service.NewCalendarService(userRepo, calendarRepo, oauthConfig)

	// Initialize handlers
	googleCalendarHandler := calendar.NewGoogleCalendarHandler(calendarService)
	icsHandler := calendar.NewICSHandler(calendarService)

	// Calendar routes with JWT middleware
	r.Route("/calendars", func(r chi.Router) {
		// Apply JWT middleware to all calendar routes
		r.Use(middleware.JWTMiddleware(zap.L()))

		// Google Calendar endpoints
		r.Route("/google", func(r chi.Router) {
			r.Get("/", googleCalendarHandler.GetCalendars)
			r.Post("/", googleCalendarHandler.ImportCalendar)
		})

		// ICS Calendar endpoints
		r.Route("/ics", func(r chi.Router) {
			r.Post("/", icsHandler.ImportICS)
		})

		// Calendar events endpoint
		r.Get("/events", googleCalendarHandler.GetCalendarEvents)
	})
}
