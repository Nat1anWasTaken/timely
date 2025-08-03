package user

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/NathanWasTaken/timely/backend/internal/model"
	"github.com/NathanWasTaken/timely/backend/internal/service"
)

type UserEventsHandler struct {
	calendarService *service.CalendarService
	userService     *service.UserService
	logger          *zap.Logger
}

func NewUserEventsHandler(calendarService *service.CalendarService, userService *service.UserService) *UserEventsHandler {
	return &UserEventsHandler{
		calendarService: calendarService,
		userService:     userService,
		logger:          zap.L(),
	}
}

// GetPublicUserEvents retrieves public calendar events for a specific user
// @Summary Get Public User Events
// @Description Retrieves public calendar events for a specific user within a specified time range (max 6 months). No authentication required.
// @Tags User
// @Produce json
// @Param username path string true "Username"
// @Param start_timestamp query string true "Start timestamp in Unix format"
// @Param end_timestamp query string true "End timestamp in Unix format"
// @Success 200 {object} model.CalendarEventsResponse "Public events retrieved successfully"
// @Failure 400 {object} model.ErrorResponse "Bad Request - Invalid parameters or time range"
// @Failure 404 {object} model.ErrorResponse "Not Found - User not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/users/{username}/events [get]
func (h *UserEventsHandler) GetPublicUserEvents(w http.ResponseWriter, r *http.Request) {
	// Get user ID from path parameter
	username := r.PathValue("username")
	if username == "" {
		h.logger.Error("Username not provided in path")
		sendEventsErrorResponse(w, "Username is required", "missing_username", http.StatusBadRequest)
		return
	}

	// Get user by username
	user, err := h.userService.GetUserByUsername(username)
	if err != nil {
		h.logger.Error("Failed to get user by username", zap.Error(err), zap.String("username", username))
		sendEventsErrorResponse(w, "User not found", "user_not_found", http.StatusNotFound)
		return
	}

	// Parse query parameters
	startTimestampStr := r.URL.Query().Get("start_timestamp")
	endTimestampStr := r.URL.Query().Get("end_timestamp")

	// Validate query parameters
	if startTimestampStr == "" || endTimestampStr == "" {
		sendEventsErrorResponse(w, "Start timestamp and end timestamp query parameters are required", "missing_time_range", http.StatusBadRequest)
		return
	}

	// Parse timestamps
	startTimestamp, err := strconv.ParseInt(startTimestampStr, 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse start timestamp", zap.Error(err), zap.String("start_timestamp", startTimestampStr))
		sendEventsErrorResponse(w, "Invalid start timestamp format", "invalid_start_timestamp", http.StatusBadRequest)
		return
	}

	endTimestamp, err := strconv.ParseInt(endTimestampStr, 10, 64)
	if err != nil {
		h.logger.Error("Failed to parse end timestamp", zap.Error(err), zap.String("end_timestamp", endTimestampStr))
		sendEventsErrorResponse(w, "Invalid end timestamp format", "invalid_end_timestamp", http.StatusBadRequest)
		return
	}

	// Convert timestamps to time.Time
	startTime := time.Unix(startTimestamp, 0)
	endTime := time.Unix(endTimestamp, 0)

	// Validate time range
	if startTime.After(endTime) {
		sendEventsErrorResponse(w, "Start time must be before end time", "invalid_time_range", http.StatusBadRequest)
		return
	}

	h.logger.Info("Fetching public calendar events for user",
		zap.Uint64("user_id", user.ID),
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime))

	// Get public calendar events from service
	calendarsWithEvents, err := h.calendarService.GetPublicUserCalendarEvents(user.ID, startTime, endTime)
	if err != nil {
		h.logger.Error("Failed to get public calendar events", zap.Error(err), zap.Uint64("user_id", user.ID))

		// Handle specific error cases
		switch {
		case err.Error() == "time range cannot exceed 6 months":
			sendEventsErrorResponse(w, "Time range cannot exceed 6 months", "time_range_too_large", http.StatusBadRequest)
		case err.Error() == "user not found":
			sendEventsErrorResponse(w, "User not found", "user_not_found", http.StatusNotFound)
		default:
			sendEventsErrorResponse(w, "Failed to retrieve public calendar events", "calendar_events_fetch_error", http.StatusInternalServerError)
		}
		return
	}

	// Create success response
	response := model.CalendarEventsResponse{
		Success:   true,
		Message:   "Public calendar events retrieved successfully",
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

	h.logger.Info("Successfully retrieved public calendar events",
		zap.Uint64("user_id", user.ID),
		zap.Int("calendar_count", len(calendarsWithEvents)),
		zap.Int("total_events", totalEvents))
}

// sendEventsErrorResponse sends a standardized error response for user events
func sendEventsErrorResponse(w http.ResponseWriter, message, errorType string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := model.ErrorResponse{
		Success: false,
		Message: message,
		Error:   errorType,
	}

	json.NewEncoder(w).Encode(response)
}
