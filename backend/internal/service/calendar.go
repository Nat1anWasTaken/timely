package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

// GetUserCalendars retrieves all calendars for a user with smart sync logic
func (s *CalendarService) GetUserCalendars(userID uint64) ([]*model.GoogleCalendar, error) {
	return s.GetUserCalendarsWithSync(userID, false)
}

// GetUserCalendarsWithSync retrieves calendars with optional force sync
func (s *CalendarService) GetUserCalendarsWithSync(userID uint64, forceSync bool) ([]*model.GoogleCalendar, error) {
	// Check if calendars exist first
	localCalendars, err := s.calendarRepo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get local calendars: %w", err)
	}

	// If no calendars exist, fetch from Google API
	if len(localCalendars) == 0 {
		s.logger.Info("No local calendars found, fetching from Google API", zap.Uint64("user_id", userID))
		return s.fetchUserCalendarsFromGoogle(userID)
	}

	// Use the reusable sync function
	synced, err := s.SyncIfNeeded(userID, forceSync)
	if err != nil {
		return nil, fmt.Errorf("failed to sync calendars: %w", err)
	}

	// Get fresh local calendars after potential sync
	localCalendars, err = s.calendarRepo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated local calendars: %w", err)
	}

	s.logger.Debug("Retrieved calendar data", 
		zap.Uint64("user_id", userID),
		zap.Bool("synced", synced),
		zap.Int("calendar_count", len(localCalendars)))

	return s.convertLocalCalendarsToGoogle(localCalendars)
}

// fetchUserCalendarsFromGoogle fetches calendars from Google API (original logic)
func (s *CalendarService) fetchUserCalendarsFromGoogle(userID uint64) ([]*model.GoogleCalendar, error) {
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

// SyncIfNeeded checks if calendars need syncing and performs sync if necessary
// Returns true if sync was performed, false if cached data was used
func (s *CalendarService) SyncIfNeeded(userID uint64, forceSync bool) (bool, error) {
	// Get local calendars to check sync status
	localCalendars, err := s.calendarRepo.FindByUserID(userID)
	if err != nil {
		return false, fmt.Errorf("failed to get local calendars: %w", err)
	}

	// Check if sync is needed
	needsSync := forceSync || len(localCalendars) == 0
	
	if !needsSync {
		for _, calendar := range localCalendars {
			if time.Since(calendar.SyncedAt) > 5*time.Minute {
				needsSync = true
				break
			}
		}
	}

	if !needsSync {
		s.logger.Debug("Using cached data - no sync needed", zap.Uint64("user_id", userID))
		return false, nil
	}

	// Perform sync
	if forceSync {
		s.logger.Info("Performing force sync", zap.Uint64("user_id", userID))
	} else {
		s.logger.Info("Performing automatic sync - cache expired", zap.Uint64("user_id", userID))
	}

	// Check if user has valid Google account before attempting sync
	account, err := s.userRepo.FindGoogleAccountByUserID(userID)
	if err != nil {
		s.logger.Error("No Google account found for user", 
			zap.Error(err),
			zap.Uint64("user_id", userID))
		return false, nil // Return cached data without error
	}

	// Validate account has required tokens
	if account.AccessToken == nil || account.RefreshToken == nil {
		s.logger.Error("Google account missing OAuth tokens", 
			zap.Uint64("user_id", userID))
		return false, nil // Return cached data without error
	}

	// Try to refresh token if needed
	if err := s.refreshTokenIfNeeded(account); err != nil {
		s.logger.Error("Failed to refresh token, using cached data", 
			zap.Error(err),
			zap.Uint64("user_id", userID))
		return false, nil // Return cached data without error
	}

	// Sync each calendar's events
	syncSuccessCount := 0
	syncAttempts := 0
	
	for _, calendar := range localCalendars {
		if calendar.SourceID != nil {
			shouldSyncCalendar := forceSync || time.Since(calendar.SyncedAt) > 5*time.Minute
			if shouldSyncCalendar {
				syncAttempts++
				if err := s.SyncCalendarEvents(userID, *calendar.SourceID); err != nil {
					s.logger.Error("Failed to sync calendar events",
						zap.Error(err),
						zap.String("calendar_source_id", *calendar.SourceID),
						zap.Uint64("user_id", userID))
					// Continue with other calendars - don't fail the entire operation
				} else {
					syncSuccessCount++
				}
			}
		}
	}

	if syncAttempts > 0 {
		s.logger.Info("Calendar sync completed",
			zap.Uint64("user_id", userID),
			zap.Int("successful_syncs", syncSuccessCount),
			zap.Int("attempted_syncs", syncAttempts))
	}

	// Return true if we attempted sync (even if some failed)
	return syncAttempts > 0, nil
}

// convertLocalCalendarsToGoogle converts local calendars to Google Calendar format
func (s *CalendarService) convertLocalCalendarsToGoogle(localCalendars []*model.Calendar) ([]*model.GoogleCalendar, error) {
	var googleCalendars []*model.GoogleCalendar
	
	for _, localCalendar := range localCalendars {
		if localCalendar.SourceID == nil {
			continue
		}

		googleCalendar := &model.GoogleCalendar{
			ID:       *localCalendar.SourceID,
			Summary:  localCalendar.Summary,
			TimeZone: localCalendar.TimeZone,
			Selected: true,
		}

		if localCalendar.Description != nil {
			googleCalendar.Description = *localCalendar.Description
		}
		if localCalendar.EventColor != nil {
			googleCalendar.BackgroundColor = *localCalendar.EventColor
		}

		googleCalendars = append(googleCalendars, googleCalendar)
	}

	return googleCalendars, nil
}

// refreshTokenIfNeeded checks if the token is expired and refreshes it if necessary
func (s *CalendarService) refreshTokenIfNeeded(account *model.Account) error {
	// Always try to refresh if token is expired or close to expiring
	needsRefresh := false
	
	if account.Expiry == nil {
		s.logger.Warn("Token has no expiry time, forcing refresh", zap.Uint64("user_id", account.UserID))
		needsRefresh = true
	} else if time.Now().Add(5*time.Minute).After(*account.Expiry) {
		s.logger.Info("Token is expired or expiring soon, refreshing", 
			zap.Uint64("user_id", account.UserID),
			zap.Time("expiry", *account.Expiry))
		needsRefresh = true
	}

	if !needsRefresh {
		s.logger.Debug("Token is still valid", zap.Uint64("user_id", account.UserID))
		return nil
	}

	s.logger.Info("Refreshing Google OAuth token", zap.Uint64("user_id", account.UserID))

	// Validate we have refresh token
	if account.RefreshToken == nil || *account.RefreshToken == "" {
		return fmt.Errorf("no refresh token available for user %d", account.UserID)
	}

	// Create oauth2.Token for refresh
	oauthToken := &oauth2.Token{
		RefreshToken: *account.RefreshToken,
	}

	// Set access token if we have one (even if expired)
	if account.AccessToken != nil {
		oauthToken.AccessToken = *account.AccessToken
	}

	// Set expiry if we have one
	if account.Expiry != nil {
		oauthToken.Expiry = *account.Expiry
	}

	// Refresh the token using the OAuth config
	ctx := context.Background()
	tokenSource := s.oauthConfig.Google.TokenSource(ctx, oauthToken)
	
	newToken, err := tokenSource.Token()
	if err != nil {
		s.logger.Error("Failed to refresh OAuth token", 
			zap.Error(err),
			zap.Uint64("user_id", account.UserID))
		return fmt.Errorf("failed to refresh token for user %d: %w", account.UserID, err)
	}

	// Validate new token
	if newToken.AccessToken == "" {
		return fmt.Errorf("received empty access token for user %d", account.UserID)
	}

	// Update the token in database
	refreshToken := newToken.RefreshToken
	if refreshToken == "" && account.RefreshToken != nil {
		// Keep existing refresh token if new one is empty
		refreshToken = *account.RefreshToken
	}

	if err := s.userRepo.UpdateGoogleAccountTokens(account.UserID, newToken.AccessToken, refreshToken, &newToken.Expiry); err != nil {
		return fmt.Errorf("failed to update token in database: %w", err)
	}

	// Update the account object with new tokens for immediate use
	account.AccessToken = &newToken.AccessToken
	account.RefreshToken = &refreshToken
	account.Expiry = &newToken.Expiry

	s.logger.Info("Successfully refreshed Google token", 
		zap.Uint64("user_id", account.UserID),
		zap.Time("new_expiry", newToken.Expiry))
	
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

// fetchEventsFromGoogleWithRetry fetches events with retry logic for authentication failures
func (s *CalendarService) fetchEventsFromGoogleWithRetry(accessToken, calendarID string, userID uint64) ([]*model.GoogleCalendarEvent, error) {
	// First attempt
	events, err := s.fetchEventsFromGoogle(accessToken, calendarID)
	if err == nil {
		return events, nil
	}

	// Check if it's an authentication error (401)
	if !s.isAuthError(err) {
		return nil, err // Not an auth error, return original error
	}

	s.logger.Warn("Authentication error detected, attempting token refresh and retry",
		zap.Error(err),
		zap.Uint64("user_id", userID),
		zap.String("calendar_id", calendarID))

	// Get fresh account info and try to refresh token
	account, err := s.userRepo.FindGoogleAccountByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account for retry: %w", err)
	}

	// Force token refresh
	if err := s.forceRefreshToken(account); err != nil {
		return nil, fmt.Errorf("failed to force refresh token: %w", err)
	}

	// Retry with new token
	s.logger.Info("Retrying event fetch with refreshed token",
		zap.Uint64("user_id", userID),
		zap.String("calendar_id", calendarID))

	events, err = s.fetchEventsFromGoogle(*account.AccessToken, calendarID)
	if err != nil {
		return nil, fmt.Errorf("retry failed: %w", err)
	}

	return events, nil
}

// isAuthError checks if an error is an authentication error
func (s *CalendarService) isAuthError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "status 401") || 
		   strings.Contains(errStr, "UNAUTHENTICATED") ||
		   strings.Contains(errStr, "Invalid Credentials")
}

// forceRefreshToken forces a token refresh regardless of expiry time
func (s *CalendarService) forceRefreshToken(account *model.Account) error {
	s.logger.Info("Force refreshing OAuth token", zap.Uint64("user_id", account.UserID))

	// Validate we have refresh token
	if account.RefreshToken == nil || *account.RefreshToken == "" {
		return fmt.Errorf("no refresh token available for user %d", account.UserID)
	}

	// Create oauth2.Token for refresh
	oauthToken := &oauth2.Token{
		RefreshToken: *account.RefreshToken,
	}

	// Refresh the token using the OAuth config
	ctx := context.Background()
	tokenSource := s.oauthConfig.Google.TokenSource(ctx, oauthToken)
	
	newToken, err := tokenSource.Token()
	if err != nil {
		s.logger.Error("Failed to force refresh OAuth token", 
			zap.Error(err),
			zap.Uint64("user_id", account.UserID))
		return fmt.Errorf("failed to force refresh token for user %d: %w", account.UserID, err)
	}

	// Validate new token
	if newToken.AccessToken == "" {
		return fmt.Errorf("received empty access token for user %d", account.UserID)
	}

	// Update the token in database
	refreshToken := newToken.RefreshToken
	if refreshToken == "" && account.RefreshToken != nil {
		// Keep existing refresh token if new one is empty
		refreshToken = *account.RefreshToken
	}

	if err := s.userRepo.UpdateGoogleAccountTokens(account.UserID, newToken.AccessToken, refreshToken, &newToken.Expiry); err != nil {
		return fmt.Errorf("failed to update token in database: %w", err)
	}

	// Update the account object with new tokens for immediate use
	account.AccessToken = &newToken.AccessToken
	account.RefreshToken = &refreshToken
	account.Expiry = &newToken.Expiry

	s.logger.Info("Successfully force refreshed Google token", 
		zap.Uint64("user_id", account.UserID),
		zap.Time("new_expiry", newToken.Expiry))
	
	return nil
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

// GetUserCalendarEvents retrieves all events for a user's calendars within a specified time range with smart sync
func (s *CalendarService) GetUserCalendarEvents(userID uint64, startTime, endTime time.Time) ([]*model.CalendarWithEvents, error) {
	return s.GetUserCalendarEventsWithSync(userID, startTime, endTime, false)
}

// GetUserCalendarEventsWithSync retrieves events with optional force sync
func (s *CalendarService) GetUserCalendarEventsWithSync(userID uint64, startTime, endTime time.Time, forceSync bool) ([]*model.CalendarWithEvents, error) {
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

	// Use the reusable sync function
	synced, err := s.SyncIfNeeded(userID, forceSync)
	if err != nil {
		s.logger.Error("Failed to sync calendars", zap.Error(err))
		// Continue with cached data even if sync fails
	}

	// Extract calendar IDs
	var calendarIDs []uint64
	for _, calendar := range calendars {
		calendarIDs = append(calendarIDs, calendar.ID)
	}

	// Get events for all calendars within time range (either freshly synced or cached)
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
		zap.Time("end_time", endTime),
		zap.Bool("synced", synced))

	return calendarsWithEvents, nil
}

// FindCalendarBySourceID finds a calendar by its source ID (Google Calendar ID)
func (s *CalendarService) FindCalendarBySourceID(sourceID string) (*model.Calendar, error) {
	calendar, err := s.calendarRepo.FindBySourceID(sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find calendar by source ID: %w", err)
	}
	return calendar, nil
}

// SyncCalendarEvents synchronizes events for a specific calendar
func (s *CalendarService) SyncCalendarEvents(userID uint64, calendarID string) error {
	s.logger.Info("Starting calendar event sync",
		zap.Uint64("user_id", userID),
		zap.String("calendar_id", calendarID))

	// Get user's Google account
	account, err := s.userRepo.FindGoogleAccountByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get Google account: %w", err)
	}

	// Check if account has tokens
	if account.AccessToken == nil || account.RefreshToken == nil {
		return fmt.Errorf("google account not properly configured with OAuth tokens")
	}

	// Check if token needs refresh
	if err := s.refreshTokenIfNeeded(account); err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Fetch latest events from Google Calendar API with retry on auth failure
	googleEvents, err := s.fetchEventsFromGoogleWithRetry(*account.AccessToken, calendarID, userID)
	if err != nil {
		return fmt.Errorf("failed to fetch events from Google: %w", err)
	}

	// Get local calendar
	localCalendar, err := s.calendarRepo.FindBySourceID(calendarID)
	if err != nil {
		return fmt.Errorf("failed to find local calendar: %w", err)
	}

	// Get existing events for comparison
	existingEvents, err := s.calendarRepo.FindEventsByCalendarID(localCalendar.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing events: %w", err)
	}

	// Create a map of existing events by source ID for efficient lookup
	existingEventMap := make(map[string]*model.CalendarEvent)
	for _, event := range existingEvents {
		existingEventMap[event.SourceID] = event
	}

	var eventsToCreate []*model.CalendarEvent
	var eventsToUpdate []*model.CalendarEvent

	// Process Google events
	for _, googleEvent := range googleEvents {
		// Skip events without summary (cancelled events, etc.)
		if googleEvent.Summary == "" {
			continue
		}

		// Convert Google event to our format
		event, err := s.convertGoogleEventToCalendarEvent(googleEvent, localCalendar.ID)
		if err != nil {
			s.logger.Error("Failed to convert Google event",
				zap.Error(err),
				zap.String("event_id", googleEvent.ID))
			continue
		}

		// Check if event exists locally
		if existingEvent, exists := existingEventMap[googleEvent.ID]; exists {
			// Update existing event
			event.ID = existingEvent.ID // Preserve local ID
			eventsToUpdate = append(eventsToUpdate, event)
		} else {
			// Create new event
			eventsToCreate = append(eventsToCreate, event)
		}
	}

	// Create new events
	if len(eventsToCreate) > 0 {
		if err := s.calendarRepo.CreateEvents(eventsToCreate); err != nil {
			s.logger.Error("Failed to create new events",
				zap.Error(err),
				zap.Int("event_count", len(eventsToCreate)))
		} else {
			s.logger.Info("Created new events",
				zap.Int("event_count", len(eventsToCreate)),
				zap.Uint64("calendar_id", localCalendar.ID))
		}
	}

	// Update existing events
	if len(eventsToUpdate) > 0 {
		if err := s.calendarRepo.UpdateEvents(eventsToUpdate); err != nil {
			s.logger.Error("Failed to update existing events",
				zap.Error(err),
				zap.Int("event_count", len(eventsToUpdate)))
		} else {
			s.logger.Info("Updated existing events",
				zap.Int("event_count", len(eventsToUpdate)),
				zap.Uint64("calendar_id", localCalendar.ID))
		}
	}

	// Update calendar sync timestamp
	if err := s.calendarRepo.UpdateSyncedAt(localCalendar.ID, time.Now()); err != nil {
		s.logger.Error("Failed to update calendar sync timestamp", zap.Error(err))
	}

	s.logger.Info("Successfully synced calendar events",
		zap.Uint64("user_id", userID),
		zap.String("calendar_id", calendarID),
		zap.Int("created_events", len(eventsToCreate)),
		zap.Int("updated_events", len(eventsToUpdate)))

	return nil
}
