package calendar

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/NathanWasTaken/timely/backend/internal/middleware"
	"github.com/NathanWasTaken/timely/backend/internal/model"
	"github.com/NathanWasTaken/timely/backend/internal/service"
)

type GoogleCalendarHandler struct {
	calendarService *service.CalendarService
	logger          *zap.Logger
}

func NewGoogleCalendarHandler(calendarService *service.CalendarService) *GoogleCalendarHandler {
	return &GoogleCalendarHandler{
		calendarService: calendarService,
		logger:          zap.L(),
	}
}

// GetCalendars retrieves all calendars for the authenticated user
// @Summary Get User Calendars
// @Description Retrieves all Google calendars for the authenticated user
// @Tags Calendar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.CalendarListResponse "Calendars retrieved successfully"
// @Failure 401 {object} model.ErrorResponse "Unauthorized - Authentication required"
// @Failure 404 {object} model.ErrorResponse "Not Found - Google token not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/calendar/google/calendars [get]
func (h *GoogleCalendarHandler) GetCalendars(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by JWT middleware)
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User not found in context")
		sendErrorResponse(w, "Authentication required", "authentication_required", http.StatusUnauthorized)
		return
	}

	h.logger.Info("Fetching calendars for user", zap.Uint64("user_id", user.ID))

	// Get calendars from service
	calendars, err := h.calendarService.GetUserCalendars(user.ID)
	if err != nil {
		h.logger.Error("Failed to get user calendars", zap.Error(err), zap.Uint64("user_id", user.ID))

		// Handle specific error cases
		switch {
		case err.Error() == "failed to get Google token: record not found":
			sendErrorResponse(w, "Google account not connected", "google_token_not_found", http.StatusNotFound)
		default:
			sendErrorResponse(w, "Failed to retrieve calendars", "calendar_fetch_error", http.StatusInternalServerError)
		}
		return
	}

	// Create success response
	response := model.CalendarListResponse{
		Success:   true,
		Message:   "Calendars retrieved successfully",
		Calendars: calendars,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Successfully retrieved calendars",
		zap.Uint64("user_id", user.ID),
		zap.Int("calendar_count", len(calendars)))
}

// sendErrorResponse sends a standardized error response
func sendErrorResponse(w http.ResponseWriter, message, errorType string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := model.ErrorResponse{
		Success: false,
		Message: message,
		Error:   errorType,
	}

	json.NewEncoder(w).Encode(response)
}
