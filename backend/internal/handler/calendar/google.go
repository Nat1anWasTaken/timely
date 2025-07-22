package calendar

import (
	"encoding/json"
	"net/http"
	"time"

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
// @Router /api/calendar/google [get]
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

// ImportCalendarRequest represents the request body for importing a calendar
type ImportCalendarRequest struct {
	CalendarID string `json:"calendar_id" validate:"required"`
}

// ImportCalendarResponse represents the response for importing a calendar
type ImportCalendarResponse struct {
	Success  bool            `json:"success"`
	Message  string          `json:"message"`
	Calendar *model.Calendar `json:"calendar"`
}

// ImportCalendar imports a Google calendar to the database
// @Summary Import Google Calendar
// @Description Imports a specific Google calendar to the user's database
// @Tags Calendar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ImportCalendarRequest true "Import calendar request"
// @Success 201 {object} ImportCalendarResponse "Calendar imported successfully"
// @Failure 400 {object} model.ErrorResponse "Bad Request - Invalid request body"
// @Failure 401 {object} model.ErrorResponse "Unauthorized - Authentication required"
// @Failure 404 {object} model.ErrorResponse "Not Found - Google token not found or calendar not found"
// @Failure 409 {object} model.ErrorResponse "Conflict - Calendar already imported"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/calendar/google [post]
func (h *GoogleCalendarHandler) ImportCalendar(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by JWT middleware)
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User not found in context")
		sendErrorResponse(w, "Authentication required", "authentication_required", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req ImportCalendarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		sendErrorResponse(w, "Invalid request body", "invalid_request", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.CalendarID == "" {
		sendErrorResponse(w, "Calendar ID is required", "missing_calendar_id", http.StatusBadRequest)
		return
	}

	h.logger.Info("Importing calendar for user",
		zap.Uint64("user_id", user.ID),
		zap.String("calendar_id", req.CalendarID))

	// Import calendar using service
	calendar, err := h.calendarService.ImportCalendar(user.ID, req.CalendarID)
	if err != nil {
		h.logger.Error("Failed to import calendar", zap.Error(err), zap.Uint64("user_id", user.ID))

		// Handle specific error cases
		switch {
		case err.Error() == "failed to get Google account: record not found":
			sendErrorResponse(w, "Google account not connected", "google_token_not_found", http.StatusNotFound)
		case err.Error() == "calendar already imported":
			sendErrorResponse(w, "Calendar already imported", "calendar_already_imported", http.StatusConflict)
		case err.Error() == "Google account not properly configured with OAuth tokens":
			sendErrorResponse(w, "Google account not properly configured", "google_account_not_configured", http.StatusNotFound)
		case err.Error() == "calendar not found in list":
			sendErrorResponse(w, "Calendar not found in list", "calendar_not_found", http.StatusNotFound)
		default:
			sendErrorResponse(w, "Failed to import calendar", "calendar_import_error", http.StatusInternalServerError)
		}
		return
	}

	// Create success response
	response := ImportCalendarResponse{
		Success:  true,
		Message:  "Calendar imported successfully",
		Calendar: calendar,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Successfully imported calendar",
		zap.Uint64("user_id", user.ID),
		zap.String("calendar_id", req.CalendarID),
		zap.String("calendar_summary", calendar.Summary))
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

// GetCalendarEvents retrieves all events for user's calendars within a specified time range
// @Summary Get Calendar Events
// @Description Retrieves all events for user's calendars within a specified time range (max 3 months)
// @Tags Calendar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.CalendarEventsRequest true "Time range request"
// @Success 200 {object} model.CalendarEventsResponse "Events retrieved successfully"
// @Failure 400 {object} model.ErrorResponse "Bad Request - Invalid request body or time range"
// @Failure 401 {object} model.ErrorResponse "Unauthorized - Authentication required"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/calendar/events [post]
func (h *GoogleCalendarHandler) GetCalendarEvents(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by JWT middleware)
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User not found in context")
		sendErrorResponse(w, "Authentication required", "authentication_required", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req model.CalendarEventsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		sendErrorResponse(w, "Invalid request body", "invalid_request", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.StartTime == "" || req.EndTime == "" {
		sendErrorResponse(w, "Start time and end time are required", "missing_time_range", http.StatusBadRequest)
		return
	}

	// Parse time strings
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		h.logger.Error("Failed to parse start time", zap.Error(err), zap.String("start_time", req.StartTime))
		sendErrorResponse(w, "Invalid start time format", "invalid_start_time", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		h.logger.Error("Failed to parse end time", zap.Error(err), zap.String("end_time", req.EndTime))
		sendErrorResponse(w, "Invalid end time format", "invalid_end_time", http.StatusBadRequest)
		return
	}

	// Validate time range
	if startTime.After(endTime) {
		sendErrorResponse(w, "Start time must be before end time", "invalid_time_range", http.StatusBadRequest)
		return
	}

	h.logger.Info("Fetching calendar events for user",
		zap.Uint64("user_id", user.ID),
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime))

	// Get calendar events from service
	calendarsWithEvents, err := h.calendarService.GetUserCalendarEvents(user.ID, startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get calendar events", zap.Error(err), zap.Uint64("user_id", user.ID))

		// Handle specific error cases
		switch {
		case err.Error() == "time range cannot exceed 3 months":
			sendErrorResponse(w, "Time range cannot exceed 3 months", "time_range_too_large", http.StatusBadRequest)
		default:
			sendErrorResponse(w, "Failed to retrieve calendar events", "calendar_events_fetch_error", http.StatusInternalServerError)
		}
		return
	}

	// Create success response
	response := model.CalendarEventsResponse{
		Success:   true,
		Message:   "Calendar events retrieved successfully",
		Calendars: calendarsWithEvents,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Log total events count
	totalEvents := 0
	for _, calendar := range calendarsWithEvents {
		totalEvents += len(calendar.Events)
	}

	h.logger.Info("Successfully retrieved calendar events",
		zap.Uint64("user_id", user.ID),
		zap.Int("calendar_count", len(calendarsWithEvents)),
		zap.Int("total_events", totalEvents))
}
