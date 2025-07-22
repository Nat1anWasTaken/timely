package calendar

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/NathanWasTaken/timely/backend/internal/middleware"
	"github.com/NathanWasTaken/timely/backend/internal/model"
)

// GetCalendarEvents retrieves all events for user's calendars within a specified time range
// @Summary Get Calendar Events
// @Description Retrieves all events for user's calendars within a specified time range (max 3 months)
// @Tags Calendar
// @Produce json
// @Security BearerAuth
// @Param start_timestamp query string true "Start timestamp in Unix format"
// @Param end_timestamp query string true "End timestamp in Unix format"
// @Success 200 {object} model.CalendarEventsResponse "Events retrieved successfully"
// @Failure 400 {object} model.ErrorResponse "Bad Request - Invalid query parameters or time range"
// @Failure 401 {object} model.ErrorResponse "Unauthorized - Authentication required"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/calendar/events [get]
func (h *GoogleCalendarHandler) GetCalendarEvents(w http.ResponseWriter, r *http.Request) {
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
