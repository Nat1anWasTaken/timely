package calendar

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/NathanWasTaken/timely/backend/internal/middleware"
	"github.com/NathanWasTaken/timely/backend/internal/model"
	"github.com/NathanWasTaken/timely/backend/internal/service"
)

type CalendarHandler struct {
	calendarService *service.CalendarService
	logger          *zap.Logger
}

func NewCalendarHandler(calendarService *service.CalendarService) *CalendarHandler {
	return &CalendarHandler{
		calendarService: calendarService,
		logger:          zap.L(),
	}
}

// GetCalendarEvents retrieves all events for user's calendars within a specified time range
// @Summary Get Calendar Events
// @Description Retrieves all events for user's calendars within a specified time range (max 3 months)
// @Tags Calendar
// @Produce json
// @Security BearerAuth
// @Param start_timestamp query string true "Start timestamp in Unix format"
// @Param end_timestamp query string true "End timestamp in Unix format"
// @Param force_sync query bool false "Force sync from Google API regardless of cache"
// @Success 200 {object} model.CalendarEventsResponse "Events retrieved successfully"
// @Failure 400 {object} model.ErrorResponse "Bad Request - Invalid query parameters or time range"
// @Failure 401 {object} model.ErrorResponse "Unauthorized - Authentication required"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/calendars/events [get]
func (h *CalendarHandler) GetCalendarEvents(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by JWT middleware)
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User not found in context")
		sendErrorResponse(w, "Authentication required", "authentication_required", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	startTimestampStr := r.URL.Query().Get("start_timestamp")
	endTimestampStr := r.URL.Query().Get("end_timestamp")
	forceSync := r.URL.Query().Get("force_sync") == "true"

	// Validate query parameters
	if startTimestampStr == "" || endTimestampStr == "" {
		sendErrorResponse(w, "Start timestamp and end timestamp query parameters are required", "missing_time_range", http.StatusBadRequest)
		return
	}

	// Parse timestamps
	startTimestamp, err := strconv.ParseInt(startTimestampStr, 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse start timestamp", zap.Error(err), zap.String("start_timestamp", startTimestampStr))
		sendErrorResponse(w, "Invalid start timestamp format", "invalid_start_timestamp", http.StatusBadRequest)
		return
	}

	endTimestamp, err := strconv.ParseInt(endTimestampStr, 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse end timestamp", zap.Error(err), zap.String("end_timestamp", endTimestampStr))
		sendErrorResponse(w, "Invalid end timestamp format", "invalid_end_timestamp", http.StatusBadRequest)
		return
	}

	// Convert timestamps to time.Time
	startTime := time.Unix(startTimestamp, 0)
	endTime := time.Unix(endTimestamp, 0)

	// Validate time range
	if startTime.After(endTime) {
		sendErrorResponse(w, "Start time must be before end time", "invalid_time_range", http.StatusBadRequest)
		return
	}

	h.logger.Info("Fetching calendar events for user",
		zap.Uint64("user_id", user.ID),
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime),
		zap.Bool("force_sync", forceSync))

	// Get calendar events from service with smart sync
	calendarsWithEvents, err := h.calendarService.GetUserCalendarEventsWithSync(user.ID, startTime, endTime, forceSync)
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

// GetImportedCalendars retrieves all imported calendars for the authenticated user
// @Summary Get Imported Calendars
// @Description Retrieves all imported calendars (Google and ICS) for the authenticated user
// @Tags Calendar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.ImportedCalendarsResponse "Imported calendars retrieved successfully"
// @Failure 401 {object} model.ErrorResponse "Unauthorized - Authentication required"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/calendars [get]
func (h *CalendarHandler) GetImportedCalendars(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by JWT middleware)
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User not found in context")
		sendErrorResponse(w, "Authentication required", "authentication_required", http.StatusUnauthorized)
		return
	}

	h.logger.Info("Fetching imported calendars for user", zap.Uint64("user_id", user.ID))

	// Get all imported calendars from service
	calendars, err := h.calendarService.GetImportedCalendars(user.ID)
	if err != nil {
		h.logger.Error("Failed to get imported calendars", zap.Error(err), zap.Uint64("user_id", user.ID))
		sendErrorResponse(w, "Failed to retrieve imported calendars", "calendar_fetch_error", http.StatusInternalServerError)
		return
	}

	// Create success response
	response := model.ImportedCalendarsResponse{
		Success:   true,
		Message:   "Imported calendars retrieved successfully",
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

	h.logger.Info("Successfully retrieved imported calendars",
		zap.Uint64("user_id", user.ID),
		zap.Int("calendar_count", len(calendars)))
}

// UpdateCalendar updates an existing calendar
// @Summary Update Calendar
// @Description Updates an existing calendar's properties such as summary, description, visibility, etc.
// @Tags Calendar
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Calendar ID"
// @Param request body model.CalendarUpdateRequest true "Calendar update request"
// @Success 200 {object} model.CalendarUpdateResponse "Calendar updated successfully"
// @Failure 400 {object} model.ErrorResponse "Bad Request - Invalid request body or calendar ID"
// @Failure 401 {object} model.ErrorResponse "Unauthorized - Authentication required"
// @Failure 404 {object} model.ErrorResponse "Not Found - Calendar not found or access denied"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/calendars/{id} [patch]
func (h *CalendarHandler) UpdateCalendar(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by JWT middleware)
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User not found in context")
		sendErrorResponse(w, "Authentication required", "authentication_required", http.StatusUnauthorized)
		return
	}

	// Get calendar ID from URL path
	calendarID := r.PathValue("id")
	if calendarID == "" {
		sendErrorResponse(w, "Calendar ID is required", "missing_calendar_id", http.StatusBadRequest)
		return
	}

	// Parse request body
	var updateRequest model.CalendarUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		sendErrorResponse(w, "Invalid request body", "invalid_request", http.StatusBadRequest)
		return
	}

	h.logger.Info("Updating calendar for user",
		zap.Uint64("user_id", user.ID),
		zap.String("calendar_id", calendarID))

	// Update calendar using service
	calendar, err := h.calendarService.UpdateCalendar(user.ID, calendarID, &updateRequest)
	if err != nil {
		h.logger.Error("Failed to update calendar", zap.Error(err), zap.Uint64("user_id", user.ID))

		// Handle specific error cases
		switch {
		case err.Error() == "calendar not found or access denied":
			sendErrorResponse(w, "Calendar not found or access denied", "calendar_not_found", http.StatusNotFound)
		case err.Error() == "failed to find calendar: record not found":
			sendErrorResponse(w, "Calendar not found", "calendar_not_found", http.StatusNotFound)
		default:
			sendErrorResponse(w, "Failed to update calendar", "calendar_update_error", http.StatusInternalServerError)
		}
		return
	}

	// Create success response
	response := model.CalendarUpdateResponse{
		Success:  true,
		Message:  "Calendar updated successfully",
		Calendar: calendar,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Successfully updated calendar",
		zap.Uint64("user_id", user.ID),
		zap.String("calendar_id", calendarID),
		zap.String("calendar_summary", calendar.Summary))
}

// DeleteCalendar deletes an existing calendar and all its events
// @Summary Delete Calendar
// @Description Deletes an existing calendar and all its associated events
// @Tags Calendar
// @Produce json
// @Security BearerAuth
// @Param id path string true "Calendar ID"
// @Success 200 {object} model.CalendarDeleteResponse "Calendar deleted successfully"
// @Failure 400 {object} model.ErrorResponse "Bad Request - Invalid calendar ID"
// @Failure 401 {object} model.ErrorResponse "Unauthorized - Authentication required"
// @Failure 404 {object} model.ErrorResponse "Not Found - Calendar not found or access denied"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/calendars/{id} [delete]
func (h *CalendarHandler) DeleteCalendar(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by JWT middleware)
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User not found in context")
		sendErrorResponse(w, "Authentication required", "authentication_required", http.StatusUnauthorized)
		return
	}

	// Get calendar ID from URL path
	calendarID := r.PathValue("id")
	if calendarID == "" {
		sendErrorResponse(w, "Calendar ID is required", "missing_calendar_id", http.StatusBadRequest)
		return
	}

	h.logger.Info("Deleting calendar for user",
		zap.Uint64("user_id", user.ID),
		zap.String("calendar_id", calendarID))

	// Delete calendar using service
	if err := h.calendarService.DeleteCalendar(user.ID, calendarID); err != nil {
		h.logger.Error("Failed to delete calendar", zap.Error(err), zap.Uint64("user_id", user.ID))

		// Handle specific error cases
		switch {
		case err.Error() == "calendar not found or access denied":
			sendErrorResponse(w, "Calendar not found or access denied", "calendar_not_found", http.StatusNotFound)
		case err.Error() == "failed to find calendar: record not found":
			sendErrorResponse(w, "Calendar not found", "calendar_not_found", http.StatusNotFound)
		default:
			sendErrorResponse(w, "Failed to delete calendar", "calendar_delete_error", http.StatusInternalServerError)
		}
		return
	}

	// Create success response
	response := model.CalendarDeleteResponse{
		Success: true,
		Message: "Calendar deleted successfully",
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Successfully deleted calendar",
		zap.Uint64("user_id", user.ID),
		zap.String("calendar_id", calendarID))
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
