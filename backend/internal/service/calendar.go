package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/NathanWasTaken/timely/backend/internal/config"
	"github.com/NathanWasTaken/timely/backend/internal/model"
	"github.com/NathanWasTaken/timely/backend/internal/repository"
	"github.com/NathanWasTaken/timely/backend/pkg/utils"
)

type CalendarService struct {
	userRepo     *repository.UserRepository
	calendarRepo *repository.CalendarRepository
	oauthConfig  *config.OAuthConfig
	logger       *zap.Logger
}

func NewCalendarService(userRepo *repository.UserRepository, calendarRepo *repository.CalendarRepository, oauthConfig *config.OAuthConfig) *CalendarService {
	return &CalendarService{
		userRepo:     userRepo,
		calendarRepo: calendarRepo,
		oauthConfig:  oauthConfig,
		logger:       zap.L(),
	}
}

// GetUserCalendars retrieves all calendars for a user from Google Calendar API
func (s *CalendarService) GetUserCalendars(userID uint64) ([]*model.GoogleCalendar, error) {
	// Get user's Google account
	account, err := s.userRepo.FindGoogleAccountByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Google account: %w", err)
	}

	// Check if account has tokens
	if account.AccessToken == nil || account.RefreshToken == nil {
		return nil, fmt.Errorf("google account not properly configured with OAuth tokens")
	}

	// Check if token needs refresh
	if err := s.refreshTokenIfNeeded(account); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Fetch calendars from Google API
	calendars, err := s.fetchCalendarsFromGoogle(*account.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch calendars from Google: %w", err)
	}

	return calendars, nil
}

// refreshTokenIfNeeded checks if the token is expired and refreshes it if necessary
func (s *CalendarService) refreshTokenIfNeeded(account *model.Account) error {
	// Check if token is expired (with 5 minute buffer)
	if account.Expiry == nil || time.Now().Add(5*time.Minute).Before(*account.Expiry) {
		return nil // Token is still valid
	}

	s.logger.Info("Refreshing expired Google token", zap.Uint64("user_id", account.UserID))

	// Create oauth2.Token for refresh
	oauthToken := &oauth2.Token{
		AccessToken:  *account.AccessToken,
		RefreshToken: *account.RefreshToken,
		Expiry:       *account.Expiry,
	}

	// Refresh the token using the injected OAuth config
	ctx := context.Background()
	newToken, err := s.oauthConfig.Google.TokenSource(ctx, oauthToken).Token()
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update the token in database
	if err := s.userRepo.UpdateGoogleAccountTokens(account.UserID, newToken.AccessToken, newToken.RefreshToken, &newToken.Expiry); err != nil {
		return fmt.Errorf("failed to update token in database: %w", err)
	}

	s.logger.Info("Successfully refreshed Google token", zap.Uint64("user_id", account.UserID))
	return nil
}

// fetchCalendarsFromGoogle calls the Google Calendar API to get user's calendars
func (s *CalendarService) fetchCalendarsFromGoogle(accessToken string) ([]*model.GoogleCalendar, error) {
	// Create HTTP client with OAuth2 transport
	ctx := context.Background()
	oauthToken := &oauth2.Token{AccessToken: accessToken}
	client := s.oauthConfig.Google.Client(ctx, oauthToken)

	// Call Google Calendar API
	url := "https://www.googleapis.com/calendar/v3/users/me/calendarList"
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call Google Calendar API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google Calendar API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var calendarList struct {
		Items []*model.GoogleCalendar `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&calendarList); err != nil {
		return nil, fmt.Errorf("failed to decode Google Calendar API response: %w", err)
	}

	return calendarList.Items, nil
}

// ImportCalendar imports a Google calendar to the database along with its events
func (s *CalendarService) ImportCalendar(userID uint64, calendarID string) (*model.Calendar, error) {
	// Get user's Google account
	account, err := s.userRepo.FindGoogleAccountByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Google account: %w", err)
	}

	// Check if account has tokens
	if account.AccessToken == nil || account.RefreshToken == nil {
		return nil, fmt.Errorf("google account not properly configured with OAuth tokens")
	}

	// Check if token needs refresh
	if err := s.refreshTokenIfNeeded(account); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Fetch specific calendar from Google API
	googleCalendar, err := s.fetchCalendarFromGoogle(*account.AccessToken, calendarID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch calendar from Google: %w", err)
	}

	// Check if calendar already exists for this user
	exists, err := s.calendarRepo.ExistsByUserIDAndSourceID(userID, calendarID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if calendar exists: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("calendar already imported")
	}

	s.logger.Info("Importing Google calendar",
		zap.String("calendar_id", calendarID),
		zap.String("summary", googleCalendar.Summary))

	// Convert Google calendar to our Calendar model
	calendar := &model.Calendar{
		ID:          utils.GenerateID(),
		UserID:      userID,
		SourceID:    &calendarID,
		Source:      model.SourceGoogle,
		Summary:     googleCalendar.Summary,
		TimeZone:    googleCalendar.TimeZone,
		EventColor:  &googleCalendar.BackgroundColor,
		Description: &googleCalendar.Description,
		Visibility:  model.CalendarVisibilityPrivate,
		SyncedAt:    time.Now(),
	}

	// Save calendar to database
	if err := s.calendarRepo.Create(calendar); err != nil {
		return nil, fmt.Errorf("failed to save calendar to database: %w", err)
	}

	// Fetch events from Google Calendar API
	s.logger.Info("Fetching events for calendar",
		zap.String("calendar_id", calendarID),
		zap.Uint64("db_calendar_id", calendar.ID))

	googleEvents, err := s.fetchEventsFromGoogle(*account.AccessToken, calendarID)
	if err != nil {
		s.logger.Error("Failed to fetch events from Google",
			zap.Error(err),
			zap.String("calendar_id", calendarID))
		// Don't fail the import if events can't be fetched, just log the error
	} else {
		// Convert and store events
		var events []*model.CalendarEvent
		for _, googleEvent := range googleEvents {
			// Skip events without summary (cancelled events, etc.)
			if googleEvent.Summary == "" {
				continue
			}

			event, err := s.convertGoogleEventToCalendarEvent(googleEvent, calendar.ID)
			if err != nil {
				s.logger.Error("Failed to convert Google event",
					zap.Error(err),
					zap.String("event_id", googleEvent.ID))
				continue
			}

			// Check if event already exists
			exists, err := s.calendarRepo.ExistsEventBySourceID(googleEvent.ID)
			if err != nil {
				s.logger.Error("Failed to check if event exists",
					zap.Error(err),
					zap.String("event_id", googleEvent.ID))
				continue
			}

			if !exists {
				events = append(events, event)
			}
		}

		// Store events in batches
		if len(events) > 0 {
			if err := s.calendarRepo.CreateEvents(events); err != nil {
				s.logger.Error("Failed to store events",
					zap.Error(err),
					zap.Int("event_count", len(events)))
			} else {
				s.logger.Info("Successfully stored events",
					zap.Int("event_count", len(events)),
					zap.Uint64("calendar_id", calendar.ID))
			}
		} else {
			s.logger.Info("No new events to store", zap.Uint64("calendar_id", calendar.ID))
		}
	}

	s.logger.Info("Successfully imported calendar with events",
		zap.Uint64("user_id", userID),
		zap.String("calendar_id", calendarID),
		zap.String("summary", googleCalendar.Summary),
		zap.Uint64("db_calendar_id", calendar.ID))

	return calendar, nil
}

// fetchCalendarFromGoogle calls the Google Calendar API to get a specific calendar
func (s *CalendarService) fetchCalendarFromGoogle(accessToken, calendarID string) (*model.GoogleCalendar, error) {
	// Create HTTP client with OAuth2 transport
	ctx := context.Background()
	oauthToken := &oauth2.Token{AccessToken: accessToken}
	client := s.oauthConfig.Google.Client(ctx, oauthToken)

	// Call Google Calendar API list endpoint to get calendar with color info
	url := "https://www.googleapis.com/calendar/v3/users/me/calendarList"
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call Google Calendar API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google Calendar API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var calendarList struct {
		Items []model.GoogleCalendar `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&calendarList); err != nil {
		return nil, fmt.Errorf("failed to decode Google Calendar API response: %w", err)
	}

	// Find the specific calendar in the list
	for _, calendar := range calendarList.Items {
		if calendar.ID == calendarID {
			return &calendar, nil
		}
	}

	return nil, fmt.Errorf("calendar not found in list")
}

// fetchEventsFromGoogle calls the Google Calendar API to get events for a specific calendar
func (s *CalendarService) fetchEventsFromGoogle(accessToken, calendarID string) ([]*model.GoogleCalendarEvent, error) {
	// Create HTTP client with OAuth2 transport
	ctx := context.Background()
	oauthToken := &oauth2.Token{AccessToken: accessToken}
	client := s.oauthConfig.Google.Client(ctx, oauthToken)

	// Calculate time range (last 30 days to next 365 days)
	now := time.Now()
	timeMin := now.AddDate(0, 0, -30).Format(time.RFC3339)
	timeMax := now.AddDate(1, 0, 0).Format(time.RFC3339)

	// Call Google Calendar API events endpoint
	url := fmt.Sprintf("https://www.googleapis.com/calendar/v3/calendars/%s/events", calendarID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("timeMin", timeMin)
	q.Add("timeMax", timeMax)
	q.Add("singleEvents", "true")
	q.Add("orderBy", "startTime")
	q.Add("maxResults", "2500") // Maximum allowed by Google
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Google Calendar API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google Calendar API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var eventsResponse model.GoogleCalendarEventsResponse
	if err := json.NewDecoder(resp.Body).Decode(&eventsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode Google Calendar API response: %w", err)
	}

	return eventsResponse.Items, nil
}

// convertGoogleEventToCalendarEvent converts a Google Calendar event to our CalendarEvent model
func (s *CalendarService) convertGoogleEventToCalendarEvent(googleEvent *model.GoogleCalendarEvent, calendarID uint64) (*model.CalendarEvent, error) {
	// Parse start time
	var startTime time.Time
	var allDay bool

	if googleEvent.Start.DateTime != "" {
		// Timed event
		parsedTime, err := time.Parse(time.RFC3339, googleEvent.Start.DateTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start time: %w", err)
		}
		startTime = parsedTime
		allDay = false
	} else if googleEvent.Start.Date != "" {
		// All-day event
		parsedDate, err := time.Parse("2006-01-02", googleEvent.Start.Date)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start date: %w", err)
		}
		startTime = parsedDate
		allDay = true
	} else {
		return nil, fmt.Errorf("event has no start time")
	}

	// Parse end time
	var endTime time.Time
	if googleEvent.End.DateTime != "" {
		// Timed event
		parsedTime, err := time.Parse(time.RFC3339, googleEvent.End.DateTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end time: %w", err)
		}
		endTime = parsedTime
	} else if googleEvent.End.Date != "" {
		// All-day event
		parsedDate, err := time.Parse("2006-01-02", googleEvent.End.Date)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end date: %w", err)
		}
		endTime = parsedDate
	} else {
		return nil, fmt.Errorf("event has no end time")
	}

	// Create calendar event
	event := &model.CalendarEvent{
		SourceID:    googleEvent.ID,
		CalendarID:  calendarID,
		Title:       googleEvent.Summary,
		Start:       startTime,
		End:         endTime,
		AllDay:      allDay,
		EventColor:  googleEvent.ColorID,
		Location:    googleEvent.Location,
		Description: googleEvent.Description,
		Visibility:  model.CalendarEventVisibilityInherited,
	}

	return event, nil
}

// GetUserCalendarEvents retrieves all events for a user's calendars within a specified time range
func (s *CalendarService) GetUserCalendarEvents(userID uint64, startTime, endTime time.Time) ([]*model.CalendarWithEvents, error) {
	// Validate time range (max 3 months)
	threeMonths := startTime.AddDate(0, 3, 0)
	if endTime.After(threeMonths) {
		return nil, fmt.Errorf("time range cannot exceed 3 months")
	}

	// Get all user's calendars
	calendars, err := s.calendarRepo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user calendars: %w", err)
	}

	if len(calendars) == 0 {
		return []*model.CalendarWithEvents{}, nil
	}

	// Extract calendar IDs
	var calendarIDs []uint64
	for _, calendar := range calendars {
		calendarIDs = append(calendarIDs, calendar.ID)
	}

	// Get events for all calendars within time range
	events, err := s.calendarRepo.FindEventsByCalendarIDsAndTimeRange(calendarIDs, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get calendar events: %w", err)
	}

	// Group events by calendar ID
	eventsByCalendar := make(map[uint64][]*model.CalendarEvent)
	for _, event := range events {
		eventsByCalendar[event.CalendarID] = append(eventsByCalendar[event.CalendarID], event)
	}

	// Create nested response structure
	var calendarsWithEvents []*model.CalendarWithEvents
	for _, calendar := range calendars {
		calendarEvents := eventsByCalendar[calendar.ID]
		if calendarEvents == nil {
			calendarEvents = []*model.CalendarEvent{}
		}

		calendarWithEvents := &model.CalendarWithEvents{
			Calendar: calendar,
			Events:   calendarEvents,
		}
		calendarsWithEvents = append(calendarsWithEvents, calendarWithEvents)
	}

	s.logger.Info("Successfully retrieved calendar events",
		zap.Uint64("user_id", userID),
		zap.Int("calendar_count", len(calendars)),
		zap.Int("total_events", len(events)),
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime))

	return calendarsWithEvents, nil
}
