package calendar

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	ics "github.com/arran4/golang-ical"
	"go.uber.org/zap"

	"github.com/NathanWasTaken/timely/backend/internal/middleware"
	"github.com/NathanWasTaken/timely/backend/internal/model"
	"github.com/NathanWasTaken/timely/backend/internal/service"
)

type ICSHandler struct {
	calendarService *service.CalendarService
	logger          *zap.Logger
}

func NewICSHandler(calendarService *service.CalendarService) *ICSHandler {
	return &ICSHandler{
		calendarService: calendarService,
		logger:          zap.L(),
	}
}

// ImportICSRequest represents the request body for importing an ICS file
type ImportICSRequest struct {
	CalendarName string `json:"calendar_name,omitempty"`
	ICSData      string `json:"ics_data" validate:"required"`
}

// ImportICSResponse represents the response for importing an ICS file
type ImportICSResponse struct {
	Success     bool            `json:"success"`
	Message     string          `json:"message"`
	Calendar    *model.Calendar `json:"calendar"`
	EventsCount int             `json:"events_count"`
}

// ImportICS imports an ICS file and creates a calendar with events
// @Summary Import ICS File
// @Description Imports an ICS file via JSON body or file upload. Calendar name is extracted from ICS properties (X-WR-CALNAME) or falls back to "Untitled Calendar"
// @Tags Calendar
// @Accept json,multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param request body ImportICSRequest false "Import ICS request (JSON) - calendar_name is optional"
// @Param calendar_name formData string false "Calendar name override (optional - will use ICS properties if not provided)"
// @Param ics_file formData file true "ICS file to upload (required for file upload)"
// @Success 201 {object} ImportICSResponse "ICS file imported successfully"
// @Failure 400 {object} model.ErrorResponse "Bad Request - Invalid request body or ICS data"
// @Failure 401 {object} model.ErrorResponse "Unauthorized - Authentication required"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/calendars/ics [post]
func (h *ICSHandler) ImportICS(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by JWT middleware)
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User not found in context")
		sendErrorResponse(w, "Authentication required", "authentication_required", http.StatusUnauthorized)
		return
	}

	var providedCalendarName, icsData string
	var err error

	// Check Content-Type to determine if it's JSON or multipart form
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Handle file upload
		providedCalendarName, icsData, err = h.handleFileUpload(r)
		if err != nil {
			h.logger.Error("Failed to handle file upload", zap.Error(err))
			sendErrorResponse(w, err.Error(), "file_upload_error", http.StatusBadRequest)
			return
		}
	} else {
		// Handle JSON request
		providedCalendarName, icsData, err = h.handleJSONRequest(r)
		if err != nil {
			h.logger.Error("Failed to handle JSON request", zap.Error(err))
			sendErrorResponse(w, err.Error(), "json_request_error", http.StatusBadRequest)
			return
		}
	}

	// Parse ICS data first to extract calendar name
	calendar, events, err := h.parseICSData(icsData)
	if err != nil {
		h.logger.Error("Failed to parse ICS data", zap.Error(err))
		sendErrorResponse(w, "Failed to parse ICS data: "+err.Error(), "invalid_ics_data", http.StatusBadRequest)
		return
	}

	// Determine final calendar name: provided name takes priority, then ICS properties, then fallback
	calendarName := providedCalendarName
	if calendarName == "" {
		calendarName = h.extractCalendarName(calendar)
	}

	h.logger.Info("Importing ICS file for user",
		zap.Uint64("user_id", user.ID),
		zap.String("calendar_name", calendarName),
		zap.String("provided_name", providedCalendarName))

	// Create calendar and import events
	createdCalendar, eventsCount, err := h.calendarService.ImportICSCalendar(user.ID, calendarName, calendar, events)
	if err != nil {
		h.logger.Error("Failed to import ICS calendar", zap.Error(err))
		sendErrorResponse(w, "Failed to import ICS calendar", "calendar_import_error", http.StatusInternalServerError)
		return
	}

	// Create success response
	response := ImportICSResponse{
		Success:     true,
		Message:     "ICS file imported successfully",
		Calendar:    createdCalendar,
		EventsCount: eventsCount,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Successfully imported ICS file",
		zap.Uint64("user_id", user.ID),
		zap.String("calendar_name", calendarName),
		zap.Int("events_count", eventsCount))
}

// handleJSONRequest handles JSON request body for ICS import
func (h *ICSHandler) handleJSONRequest(r *http.Request) (string, string, error) {
	var req ImportICSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return "", "", fmt.Errorf("invalid request body: %w", err)
	}

	if req.ICSData == "" {
		return "", "", fmt.Errorf("ICS data is required")
	}

	// Calendar name is now optional - can be extracted from ICS
	return req.CalendarName, req.ICSData, nil
}

// handleFileUpload handles multipart form file upload for ICS import
func (h *ICSHandler) handleFileUpload(r *http.Request) (string, string, error) {
	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		return "", "", fmt.Errorf("failed to parse form data: %w", err)
	}

	// Get calendar name (now optional)
	calendarName := r.FormValue("calendar_name")

	// Get uploaded file
	file, header, err := r.FormFile("ics_file")
	if err != nil {
		return "", "", fmt.Errorf("ICS file is required: %w", err)
	}
	defer file.Close()

	h.logger.Info("Processing ICS file upload",
		zap.String("filename", header.Filename))

	// Read file content
	icsData, err := io.ReadAll(file)
	if err != nil {
		return "", "", fmt.Errorf("failed to read ICS file: %w", err)
	}

	return calendarName, string(icsData), nil
}

// parseICSData parses ICS data and extracts calendar and event information
func (h *ICSHandler) parseICSData(icsData string) (*ics.Calendar, []*ics.VEvent, error) {
	// Parse the ICS data
	cal, err := ics.ParseCalendar(strings.NewReader(icsData))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse ICS data: %w", err)
	}

	// Extract events
	events := cal.Events()
	if len(events) == 0 {
		return nil, nil, fmt.Errorf("no events found in ICS file")
	}

	h.logger.Debug("Parsed ICS data",
		zap.Int("events_count", len(events)))

	return cal, events, nil
}

// extractCalendarName extracts calendar name from ICS properties
func (h *ICSHandler) extractCalendarName(cal *ics.Calendar) string {
	// Try to get X-WR-CALNAME property (common non-standard property for calendar name)
	for _, prop := range cal.CalendarProperties {
		if prop.IANAToken == "X-WR-CALNAME" && prop.Value != "" {
			return prop.Value
		}
	}

	// Try to get other calendar-level properties
	for _, prop := range cal.CalendarProperties {
		// Try NAME property
		if prop.IANAToken == "NAME" && prop.Value != "" {
			return prop.Value
		}
		// Try SUMMARY property
		if prop.IANAToken == "SUMMARY" && prop.Value != "" {
			return prop.Value
		}
	}

	// Try to extract from PRODID as last resort (clean it up)
	for _, prop := range cal.CalendarProperties {
		if prop.IANAToken == "PRODID" && prop.Value != "" {
			prodId := strings.TrimSpace(prop.Value)
			// Clean up common PRODID patterns
			if !strings.HasPrefix(prodId, "-//") && !strings.Contains(prodId, "//") {
				return prodId
			}
		}
	}

	// Default fallback
	return "Untitled Calendar"
}