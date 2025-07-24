package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"github.com/NathanWasTaken/timely/backend/internal/config"
	"github.com/NathanWasTaken/timely/backend/internal/model"
	"github.com/NathanWasTaken/timely/backend/internal/repository"
	"github.com/NathanWasTaken/timely/backend/pkg/utils"
)

type CalendarService struct {
	userRepo         *repository.UserRepository
	calendarRepo     *repository.CalendarRepository
	oauthConfig      *config.OAuthConfig
	syncTokenManager *SyncTokenManager
	logger           *zap.Logger
}

// SyncTokenManager handles sync token operations for efficient incremental sync
type SyncTokenManager struct {
	calendarRepo *repository.CalendarRepository
	logger       *zap.Logger
}

func NewSyncTokenManager(calendarRepo *repository.CalendarRepository) *SyncTokenManager {
	return &SyncTokenManager{
		calendarRepo: calendarRepo,
		logger:       zap.L(),
	}
}

// GetSyncToken retrieves the sync token for a calendar
func (stm *SyncTokenManager) GetSyncToken(calendarID uint64) (string, error) {
	calendar, err := stm.calendarRepo.FindByID(fmt.Sprintf("%d", calendarID))
	if err != nil {
		return "", fmt.Errorf("failed to find calendar: %w", err)
	}
	
	if calendar.SyncToken == nil {
		return "", nil
	}
	
	return *calendar.SyncToken, nil
}

// StoreSyncToken stores a sync token for a calendar
func (stm *SyncTokenManager) StoreSyncToken(calendarID uint64, syncToken string) error {
	return stm.calendarRepo.UpdateSyncToken(calendarID, syncToken)
}

// ClearSyncToken clears the sync token for a calendar (forcing full sync next time)
func (stm *SyncTokenManager) ClearSyncToken(calendarID uint64) error {
	return stm.calendarRepo.UpdateSyncToken(calendarID, "")
}

// ShouldPerformFullSync determines if a full sync is needed
func (stm *SyncTokenManager) ShouldPerformFullSync(calendar *model.Calendar, forceSync bool) bool {
	// Force full sync if explicitly requested
	if forceSync {
		return true
	}
	
	// Perform full sync if calendar has never been synced
	if calendar.SyncStatus == model.CalendarSyncStatusNeverSynced {
		return true
	}
	
	// Perform full sync if no sync token exists
	if calendar.SyncToken == nil || *calendar.SyncToken == "" {
		return true
	}
	
	// Perform full sync if last full sync was more than 24 hours ago
	if calendar.LastFullSync == nil || time.Since(*calendar.LastFullSync) > 24*time.Hour {
		stm.logger.Info("Performing full sync due to 24-hour threshold",
			zap.Uint64("calendar_id", calendar.ID),
			zap.Time("last_full_sync", func() time.Time {
				if calendar.LastFullSync != nil {
					return *calendar.LastFullSync
				}
				return time.Time{}
			}()))
		return true
	}
	
	return false
}

// UpdateSyncMetadata updates sync-related metadata after a successful sync
func (stm *SyncTokenManager) UpdateSyncMetadata(calendarID uint64, syncToken string, isFullSync bool) error {
	now := time.Now()
	var status model.CalendarSyncStatus
	var lastFullSync *time.Time
	
	if isFullSync {
		status = model.CalendarSyncStatusFullSyncComplete
		lastFullSync = &now
	} else {
		status = model.CalendarSyncStatusIncrementalSync
		// Don't update lastFullSync for incremental syncs
	}
	
	return stm.calendarRepo.UpdateSyncMetadata(calendarID, status, syncToken, lastFullSync, now)
}

func NewCalendarService(userRepo *repository.UserRepository, calendarRepo *repository.CalendarRepository, oauthConfig *config.OAuthConfig) *CalendarService {
	return &CalendarService{
		userRepo:         userRepo,
		calendarRepo:     calendarRepo,
		oauthConfig:      oauthConfig,
		syncTokenManager: NewSyncTokenManager(calendarRepo),
		logger:           zap.L(),
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

// GetUserCalendarsFromGoogle fetches calendars from Google API without syncing
func (s *CalendarService) GetUserCalendarsFromGoogle(userID uint64) ([]*model.GoogleCalendar, error) {
	return s.fetchUserCalendarsFromGoogle(userID)
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
			if time.Since(calendar.SyncedAt) > 1*time.Minute {
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
			shouldSyncCalendar := forceSync || time.Since(calendar.SyncedAt) > 1*time.Minute
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
	} else if time.Now().Add(5 * time.Minute).After(*account.Expiry) {
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
		SyncStatus:  model.CalendarSyncStatusNeverSynced,
		SyncToken:   nil,
		LastFullSync: nil,
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

// fetchEventsFromGoogle calls the Google Calendar API to get events for a specific calendar (full sync)
func (s *CalendarService) fetchEventsFromGoogle(accessToken, calendarID string) ([]*model.GoogleCalendarEvent, error) {
	response, err := s.fetchEventsFromGoogleWithResponse(accessToken, calendarID, "", false)
	if err != nil {
		return nil, err
	}
	return response.Items, nil
}


// fetchEventsFromGoogleWithResponse calls the Google Calendar API and returns the full response including sync tokens
func (s *CalendarService) fetchEventsFromGoogleWithResponse(accessToken, calendarID, syncToken string, isIncrementalSync bool) (*model.GoogleCalendarEventsResponse, error) {
	// Create HTTP client with OAuth2 transport
	ctx := context.Background()
	oauthToken := &oauth2.Token{AccessToken: accessToken}
	client := s.oauthConfig.Google.Client(ctx, oauthToken)

	// Call Google Calendar API events endpoint
	url := fmt.Sprintf("https://www.googleapis.com/calendar/v3/calendars/%s/events", calendarID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	
	if isIncrementalSync && syncToken != "" {
		// For incremental sync, use sync token
		q.Add("syncToken", syncToken)
	} else {
		// For full sync, use time range
		now := time.Now()
		timeMin := now.AddDate(0, 0, -30).Format(time.RFC3339)
		timeMax := now.AddDate(1, 0, 0).Format(time.RFC3339)
		q.Add("timeMin", timeMin)
		q.Add("timeMax", timeMax)
		q.Add("singleEvents", "true")
		q.Add("orderBy", "startTime")
	}
	
	// CRITICAL: Include deleted events in response for proper sync
	q.Add("showDeleted", "true")
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

	return &eventsResponse, nil
}

// fetchAllEventsWithSyncToken fetches all events using sync token, handling pagination
func (s *CalendarService) fetchAllEventsWithSyncToken(accessToken, calendarID, syncToken string) (*model.GoogleCalendarEventsResponse, error) {
	var allEvents []*model.GoogleCalendarEvent
	var nextSyncToken string
	pageToken := ""
	
	for {
		response, err := s.fetchEventsPageWithSyncToken(accessToken, calendarID, syncToken, pageToken)
		if err != nil {
			return nil, err
		}
		
		// Accumulate events
		allEvents = append(allEvents, response.Items...)
		
		// Store the sync token (only present on the last page)
		if response.NextSyncToken != "" {
			nextSyncToken = response.NextSyncToken
		}
		
		// Check if there are more pages
		if response.NextPageToken == "" {
			break
		}
		
		pageToken = response.NextPageToken
	}
	
	// Return consolidated response
	return &model.GoogleCalendarEventsResponse{
		Kind:          "calendar#events",
		Items:         allEvents,
		NextSyncToken: nextSyncToken,
	}, nil
}

// fetchEventsPageWithSyncToken fetches a single page of events with sync token
func (s *CalendarService) fetchEventsPageWithSyncToken(accessToken, calendarID, syncToken, pageToken string) (*model.GoogleCalendarEventsResponse, error) {
	// Create HTTP client with OAuth2 transport
	ctx := context.Background()
	oauthToken := &oauth2.Token{AccessToken: accessToken}
	client := s.oauthConfig.Google.Client(ctx, oauthToken)

	// Call Google Calendar API events endpoint
	url := fmt.Sprintf("https://www.googleapis.com/calendar/v3/calendars/%s/events", calendarID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	
	if syncToken != "" {
		q.Add("syncToken", syncToken)
	}
	
	if pageToken != "" {
		q.Add("pageToken", pageToken)
	}
	
	// CRITICAL: Include deleted events in response for proper sync
	q.Add("showDeleted", "true")
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

	return &eventsResponse, nil
}

// isEventDeleted checks if a Google Calendar event is marked as deleted
func (s *CalendarService) isEventDeleted(googleEvent *model.GoogleCalendarEvent) bool {
	// Google Calendar API returns deleted events with status "cancelled"
	// Also check for empty/minimal events which can indicate deletion
	isDeleted := googleEvent.Status == "cancelled"
	
	s.logger.Debug("Checking event deletion status",
		zap.String("event_id", googleEvent.ID),
		zap.String("status", googleEvent.Status),
		zap.String("summary", googleEvent.Summary),
		zap.Bool("is_deleted", isDeleted))
	
	return isDeleted
}

// processSyncResponse processes the sync response and categorizes events for CRUD operations
func (s *CalendarService) processSyncResponse(response *model.GoogleCalendarEventsResponse, calendarID uint64) (*SyncEventChanges, error) {
	s.logger.Info("Processing sync response",
		zap.Uint64("calendar_id", calendarID),
		zap.Int("total_events_from_google", len(response.Items)),
		zap.String("next_sync_token", func() string {
			if len(response.NextSyncToken) > 20 {
				return response.NextSyncToken[:20] + "..."
			}
			return response.NextSyncToken
		}()))

	changes := &SyncEventChanges{
		ToCreate: []*model.CalendarEvent{},
		ToUpdate: []*model.CalendarEvent{},
		ToDelete: []string{},
	}
	
	// Get existing events for comparison
	existingEvents, err := s.calendarRepo.FindEventsByCalendarID(calendarID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing events: %w", err)
	}
	
	// Create a map of existing events by source ID for efficient lookup
	existingEventMap := make(map[string]*model.CalendarEvent)
	for _, event := range existingEvents {
		existingEventMap[event.SourceID] = event
	}
	
	// Process each event from Google Calendar
	for _, googleEvent := range response.Items {
		if s.isEventDeleted(googleEvent) {
			// Event is deleted - add to delete list if it exists locally
			if existingEvent, exists := existingEventMap[googleEvent.ID]; exists {
				changes.ToDelete = append(changes.ToDelete, existingEvent.SourceID)
				s.logger.Info("Event marked for deletion - found locally",
					zap.String("event_id", googleEvent.ID),
					zap.String("event_title", googleEvent.Summary),
					zap.String("status", googleEvent.Status),
					zap.Uint64("calendar_id", calendarID))
			} else {
				s.logger.Info("Deleted event not found locally - skipping",
					zap.String("event_id", googleEvent.ID),
					zap.String("event_title", googleEvent.Summary),
					zap.String("status", googleEvent.Status),
					zap.Uint64("calendar_id", calendarID))
			}
			continue
		}
		
		// Skip events without summary (some system events)
		if googleEvent.Summary == "" {
			continue
		}
		
		// Convert Google event to our format
		event, err := s.convertGoogleEventToCalendarEvent(googleEvent, calendarID)
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
			changes.ToUpdate = append(changes.ToUpdate, event)
		} else {
			// Create new event
			changes.ToCreate = append(changes.ToCreate, event)
		}
	}
	
	// Log sync operation summary
	s.logger.Info("Sync response processing completed",
		zap.Uint64("calendar_id", calendarID),
		zap.Int("existing_events_count", len(existingEventMap)),
		zap.Int("events_to_create", len(changes.ToCreate)),
		zap.Int("events_to_update", len(changes.ToUpdate)),
		zap.Int("events_to_delete", len(changes.ToDelete)))
	
	return changes, nil
}

// SyncEventChanges represents the changes to apply during sync
type SyncEventChanges struct {
	ToCreate []*model.CalendarEvent
	ToUpdate []*model.CalendarEvent
	ToDelete []string // Source IDs of events to delete
}

// applySyncChanges applies the sync changes to the database
func (s *CalendarService) applySyncChanges(changes *SyncEventChanges, calendarID uint64) error {
	// Create new events
	if len(changes.ToCreate) > 0 {
		if err := s.calendarRepo.CreateEvents(changes.ToCreate); err != nil {
			s.logger.Error("Failed to create new events",
				zap.Error(err),
				zap.Int("event_count", len(changes.ToCreate)),
				zap.Uint64("calendar_id", calendarID))
			return fmt.Errorf("failed to create new events: %w", err)
		}
		s.logger.Info("Created new events",
			zap.Int("event_count", len(changes.ToCreate)),
			zap.Uint64("calendar_id", calendarID))
	}

	// Update existing events
	if len(changes.ToUpdate) > 0 {
		if err := s.calendarRepo.UpdateEvents(changes.ToUpdate); err != nil {
			s.logger.Error("Failed to update existing events",
				zap.Error(err),
				zap.Int("event_count", len(changes.ToUpdate)),
				zap.Uint64("calendar_id", calendarID))
			return fmt.Errorf("failed to update existing events: %w", err)
		}
		s.logger.Info("Updated existing events",
			zap.Int("event_count", len(changes.ToUpdate)),
			zap.Uint64("calendar_id", calendarID))
	}

	// Delete events
	if len(changes.ToDelete) > 0 {
		deletionErrors := []string{}
		successfulDeletions := 0
		
		for _, sourceID := range changes.ToDelete {
			s.logger.Info("Attempting to delete event",
				zap.String("source_id", sourceID),
				zap.Uint64("calendar_id", calendarID))
				
			if err := s.calendarRepo.DeleteEventsBySourceID(sourceID); err != nil {
				deletionErrors = append(deletionErrors, fmt.Sprintf("sourceID:%s error:%v", sourceID, err))
				s.logger.Error("Failed to delete event",
					zap.Error(err),
					zap.String("source_id", sourceID),
					zap.Uint64("calendar_id", calendarID))
			} else {
				successfulDeletions++
				s.logger.Info("Successfully deleted event",
					zap.String("source_id", sourceID),
					zap.Uint64("calendar_id", calendarID))
			}
		}
		
		s.logger.Info("Event deletion summary",
			zap.Int("attempted_deletions", len(changes.ToDelete)),
			zap.Int("successful_deletions", successfulDeletions),
			zap.Int("failed_deletions", len(deletionErrors)),
			zap.Uint64("calendar_id", calendarID))
			
		// Log errors if any occurred, but don't fail the entire sync
		if len(deletionErrors) > 0 {
			s.logger.Warn("Some event deletions failed",
				zap.Strings("deletion_errors", deletionErrors),
				zap.Uint64("calendar_id", calendarID))
		}
	}

	return nil
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
		ID:          utils.GenerateID(),
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

		// Apply calendar event redaction to event titles
		s.applyEventRedaction(calendarEvents, calendar)

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

// SyncCalendarEvents synchronizes events for a specific calendar using Google's best practices
func (s *CalendarService) SyncCalendarEvents(userID uint64, calendarID string) error {
	return s.SyncCalendarEventsWithForce(userID, calendarID, false)
}

// SyncCalendarEventsWithForce synchronizes events for a specific calendar with optional force sync
func (s *CalendarService) SyncCalendarEventsWithForce(userID uint64, calendarID string, forceSync bool) error {
	s.logger.Info("Starting calendar event sync",
		zap.Uint64("user_id", userID),
		zap.String("calendar_id", calendarID),
		zap.Bool("force_sync", forceSync))

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

	// Get local calendar
	localCalendar, err := s.calendarRepo.FindBySourceID(calendarID)
	if err != nil {
		return fmt.Errorf("failed to find local calendar: %w", err)
	}

	// Determine sync strategy using SyncTokenManager
	shouldPerformFullSync := s.syncTokenManager.ShouldPerformFullSync(localCalendar, forceSync)
	
	var response *model.GoogleCalendarEventsResponse
	var isFullSync bool
	
	if shouldPerformFullSync {
		// Perform full sync
		s.logger.Info("Performing full sync with deleted events", 
			zap.Uint64("calendar_id", localCalendar.ID),
			zap.String("google_calendar_id", calendarID),
			zap.String("reason", "full sync required"))
		
		response, err = s.fetchEventsFromGoogleWithResponse(*account.AccessToken, calendarID, "", false)
		if err != nil {
			return fmt.Errorf("failed to fetch events from Google (full sync): %w", err)
		}
		isFullSync = true
		
		// Clear existing sync token since we're doing a full sync
		if err := s.syncTokenManager.ClearSyncToken(localCalendar.ID); err != nil {
			s.logger.Warn("Failed to clear sync token", zap.Error(err))
		}
		
		s.logger.Info("Full sync API request completed",
			zap.Int("events_received", len(response.Items)),
			zap.String("next_sync_token", func() string {
				if len(response.NextSyncToken) > 20 {
					return response.NextSyncToken[:20] + "..."
				}
				return response.NextSyncToken
			}()))
	} else {
		// Perform incremental sync
		syncToken, err := s.syncTokenManager.GetSyncToken(localCalendar.ID)
		if err != nil {
			s.logger.Error("Failed to get sync token, falling back to full sync", zap.Error(err))
			// Fallback to full sync
			response, err = s.fetchEventsFromGoogleWithResponse(*account.AccessToken, calendarID, "", false)
			if err != nil {
				return fmt.Errorf("failed to fetch events from Google (fallback full sync): %w", err)
			}
			isFullSync = true
		} else {
			s.logger.Info("Performing incremental sync with deleted events", 
				zap.Uint64("calendar_id", localCalendar.ID),
				zap.String("google_calendar_id", calendarID),
				zap.String("sync_token", syncToken[:min(len(syncToken), 20)]+"..."))
			
			response, err = s.fetchAllEventsWithSyncToken(*account.AccessToken, calendarID, syncToken)
			if err != nil {
				// Check if it's a sync token invalidation error (410)
				if s.isSyncTokenInvalidError(err) {
					s.logger.Warn("Sync token invalid, performing full sync", 
						zap.Error(err),
						zap.Uint64("calendar_id", localCalendar.ID))
					
					// Clear invalid sync token and perform full sync
					if clearErr := s.syncTokenManager.ClearSyncToken(localCalendar.ID); clearErr != nil {
						s.logger.Warn("Failed to clear invalid sync token", zap.Error(clearErr))
					}
					
					response, err = s.fetchEventsFromGoogleWithResponse(*account.AccessToken, calendarID, "", false)
					if err != nil {
						return fmt.Errorf("failed to fetch events from Google (recovery full sync): %w", err)
					}
					isFullSync = true
				} else {
					return fmt.Errorf("failed to fetch events from Google (incremental sync): %w", err)
				}
			} else {
				isFullSync = false
				s.logger.Info("Incremental sync API request completed",
					zap.Int("events_received", len(response.Items)),
					zap.String("next_sync_token", func() string {
						if len(response.NextSyncToken) > 20 {
							return response.NextSyncToken[:20] + "..."
						}
						return response.NextSyncToken
					}()))
			}
		}
	}

	// Process the sync response and get changes
	changes, err := s.processSyncResponse(response, localCalendar.ID)
	if err != nil {
		return fmt.Errorf("failed to process sync response: %w", err)
	}

	// Apply changes to database
	if err := s.applySyncChanges(changes, localCalendar.ID); err != nil {
		return fmt.Errorf("failed to apply sync changes: %w", err)
	}

	// Update sync metadata
	if response.NextSyncToken != "" {
		if err := s.syncTokenManager.UpdateSyncMetadata(localCalendar.ID, response.NextSyncToken, isFullSync); err != nil {
			s.logger.Error("Failed to update sync metadata", zap.Error(err))
			// Don't fail the sync operation for metadata update failures
		}
	}

	s.logger.Info("Successfully synced calendar events",
		zap.Uint64("user_id", userID),
		zap.String("calendar_id", calendarID),
		zap.Uint64("local_calendar_id", localCalendar.ID),
		zap.Bool("is_full_sync", isFullSync),
		zap.Int("created_events", len(changes.ToCreate)),
		zap.Int("updated_events", len(changes.ToUpdate)),
		zap.Int("deleted_events", len(changes.ToDelete)))

	return nil
}

// isSyncTokenInvalidError checks if an error indicates sync token invalidation (410 Gone)
func (s *CalendarService) isSyncTokenInvalidError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "status 410") || strings.Contains(errStr, "Gone")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ImportICSCalendar creates a calendar from ICS data and imports all events
func (s *CalendarService) ImportICSCalendar(userID uint64, calendarName string, icsCalendar *ics.Calendar, icsEvents []*ics.VEvent) (*model.Calendar, int, error) {
	s.logger.Info("Importing ICS calendar",
		zap.Uint64("user_id", userID),
		zap.String("calendar_name", calendarName),
		zap.Int("events_count", len(icsEvents)))

	// Create the calendar
	calendar := &model.Calendar{
		ID:           utils.GenerateID(),
		UserID:       userID,
		Source:       model.SourceICS,
		Summary:      calendarName,
		TimeZone:     "UTC", // Default timezone, could be extracted from ICS if available
		Visibility:   model.CalendarVisibilityPrivate,
		SyncedAt:     time.Now(),
		SyncStatus:   model.CalendarSyncStatusFullSyncComplete, // ICS imports are considered complete
		SyncToken:    nil, // ICS calendars don't use sync tokens
		LastFullSync: func() *time.Time { now := time.Now(); return &now }(),
	}

	// Save calendar to database
	if err := s.calendarRepo.Create(calendar); err != nil {
		return nil, 0, fmt.Errorf("failed to create calendar: %w", err)
	}

	// Convert and import events
	var calendarEvents []*model.CalendarEvent
	successCount := 0

	for _, icsEvent := range icsEvents {
		event, err := s.convertICSEventToCalendarEvent(icsEvent, calendar.ID)
		if err != nil {
			s.logger.Error("Failed to convert ICS event",
				zap.Error(err),
				zap.String("event_id", icsEvent.Id()))
			continue
		}

		calendarEvents = append(calendarEvents, event)
		successCount++
	}

	// Batch create events
	if len(calendarEvents) > 0 {
		if err := s.calendarRepo.CreateEvents(calendarEvents); err != nil {
			s.logger.Error("Failed to create events", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to create events: %w", err)
		}
	}

	s.logger.Info("Successfully imported ICS calendar",
		zap.Uint64("user_id", userID),
		zap.String("calendar_name", calendarName),
		zap.Uint64("calendar_id", calendar.ID),
		zap.Int("imported_events", successCount))

	return calendar, successCount, nil
}

// convertICSEventToCalendarEvent converts an ICS event to our internal CalendarEvent format
func (s *CalendarService) convertICSEventToCalendarEvent(icsEvent *ics.VEvent, calendarID uint64) (*model.CalendarEvent, error) {
	// Extract basic event information
	summary := icsEvent.GetProperty(ics.ComponentPropertySummary)
	if summary == nil {
		return nil, fmt.Errorf("event missing summary")
	}

	// Extract start time
	dtStart := icsEvent.GetProperty(ics.ComponentPropertyDtStart)
	if dtStart == nil {
		return nil, fmt.Errorf("event missing start time")
	}

	// Extract end time
	dtEnd := icsEvent.GetProperty(ics.ComponentPropertyDtEnd)
	if dtEnd == nil {
		return nil, fmt.Errorf("event missing end time")
	}

	// Parse start time
	startTime, err := s.parseICSDateTime(dtStart.Value, dtStart.ICalParameters)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start time: %w", err)
	}

	// Parse end time
	endTime, err := s.parseICSDateTime(dtEnd.Value, dtEnd.ICalParameters)
	if err != nil {
		return nil, fmt.Errorf("failed to parse end time: %w", err)
	}

	// Determine if it's an all-day event
	allDay := false
	if valueParams, exists := dtStart.ICalParameters["VALUE"]; exists && len(valueParams) > 0 {
		allDay = valueParams[0] == "DATE"
	}

	// Extract optional fields
	description := ""
	if desc := icsEvent.GetProperty(ics.ComponentPropertyDescription); desc != nil {
		description = desc.Value
	}

	location := ""
	if loc := icsEvent.GetProperty(ics.ComponentPropertyLocation); loc != nil {
		location = loc.Value
	}

	// Create calendar event
	event := &model.CalendarEvent{
		ID:          utils.GenerateID(),
		SourceID:    icsEvent.Id(),
		CalendarID:  calendarID,
		Title:       summary.Value,
		Start:       startTime,
		End:         endTime,
		AllDay:      allDay,
		Location:    location,
		Description: description,
		Visibility:  model.CalendarEventVisibilityInherited,
	}

	return event, nil
}

// parseICSDateTime parses ICS date/time strings
func (s *CalendarService) parseICSDateTime(value string, params map[string][]string) (time.Time, error) {
	// Check if it's a DATE value (all-day event)
	if valueParams, exists := params["VALUE"]; exists && len(valueParams) > 0 && valueParams[0] == "DATE" {
		// Parse date only: YYYYMMDD
		return time.Parse("20060102", value)
	}

	// Check for timezone information
	if tzidParams, exists := params["TZID"]; exists && len(tzidParams) > 0 {
		tzid := tzidParams[0]
		// Parse with timezone: YYYYMMDDTHHMMSS with TZID
		loc, err := time.LoadLocation(tzid)
		if err != nil {
			// If timezone loading fails, use UTC
			s.logger.Warn("Failed to load timezone, using UTC",
				zap.String("tzid", tzid),
				zap.Error(err))
			loc = time.UTC
		}

		parsedTime, err := time.ParseInLocation("20060102T150405", value, loc)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse datetime with timezone: %w", err)
		}
		return parsedTime, nil
	}

	// Check if it ends with 'Z' (UTC time)
	if strings.HasSuffix(value, "Z") {
		return time.Parse("20060102T150405Z", value)
	}

	// Default: parse as local time
	return time.Parse("20060102T150405", value)
}

// GetImportedCalendars retrieves all imported calendars for a user
func (s *CalendarService) GetImportedCalendars(userID uint64) ([]*model.Calendar, error) {
	s.logger.Info("Getting imported calendars for user", zap.Uint64("user_id", userID))

	// Get all calendars for the user
	calendars, err := s.calendarRepo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get imported calendars: %w", err)
	}

	s.logger.Debug("Retrieved imported calendars",
		zap.Uint64("user_id", userID),
		zap.Int("calendar_count", len(calendars)))

	return calendars, nil
}

// UpdateCalendar updates an existing calendar with new data
func (s *CalendarService) UpdateCalendar(userID uint64, calendarID string, updateRequest *model.CalendarUpdateRequest) (*model.Calendar, error) {
	s.logger.Info("Updating calendar",
		zap.Uint64("user_id", userID),
		zap.String("calendar_id", calendarID))

	// Find the calendar and verify ownership
	calendar, err := s.calendarRepo.FindByID(calendarID)
	if err != nil {
		return nil, fmt.Errorf("failed to find calendar: %w", err)
	}

	// Verify user owns this calendar
	if calendar.UserID != userID {
		return nil, fmt.Errorf("calendar not found or access denied")
	}

	// Update fields if provided
	updated := false
	if updateRequest.Summary != nil {
		calendar.Summary = *updateRequest.Summary
		updated = true
	}
	if updateRequest.Description != nil {
		calendar.Description = updateRequest.Description
		updated = true
	}
	if updateRequest.EventRedaction != nil {
		calendar.EventRedaction = updateRequest.EventRedaction
		updated = true
	}
	if updateRequest.EventColor != nil {
		calendar.EventColor = updateRequest.EventColor
		updated = true
	}
	if updateRequest.Visibility != nil {
		calendar.Visibility = *updateRequest.Visibility
		updated = true
	}
	if updateRequest.TimeZone != nil {
		calendar.TimeZone = *updateRequest.TimeZone
		updated = true
	}

	if !updated {
		s.logger.Info("No fields to update", zap.String("calendar_id", calendarID))
		return calendar, nil
	}

	// Save updated calendar
	if err := s.calendarRepo.Update(calendar); err != nil {
		return nil, fmt.Errorf("failed to update calendar: %w", err)
	}

	s.logger.Info("Successfully updated calendar",
		zap.Uint64("user_id", userID),
		zap.String("calendar_id", calendarID),
		zap.String("summary", calendar.Summary))

	return calendar, nil
}

// DeleteCalendar deletes a calendar and all its events
func (s *CalendarService) DeleteCalendar(userID uint64, calendarID string) error {
	s.logger.Info("Deleting calendar",
		zap.Uint64("user_id", userID),
		zap.String("calendar_id", calendarID))

	// Find the calendar and verify ownership
	calendar, err := s.calendarRepo.FindByID(calendarID)
	if err != nil {
		return fmt.Errorf("failed to find calendar: %w", err)
	}

	// Verify user owns this calendar
	if calendar.UserID != userID {
		return fmt.Errorf("calendar not found or access denied")
	}

	// Delete all events for this calendar first
	if err := s.calendarRepo.DeleteEventsByCalendarID(calendar.ID); err != nil {
		s.logger.Error("Failed to delete calendar events",
			zap.Error(err),
			zap.Uint64("calendar_id", calendar.ID))
		return fmt.Errorf("failed to delete calendar events: %w", err)
	}

	// Delete the calendar itself
	if err := s.calendarRepo.Delete(calendarID); err != nil {
		return fmt.Errorf("failed to delete calendar: %w", err)
	}

	s.logger.Info("Successfully deleted calendar and its events",
		zap.Uint64("user_id", userID),
		zap.String("calendar_id", calendarID),
		zap.String("summary", calendar.Summary))

	return nil
}

// GetPublicUserCalendarEvents retrieves public calendar events for a user within a specified time range
func (s *CalendarService) GetPublicUserCalendarEvents(userID uint64, startTime, endTime time.Time) ([]*model.CalendarWithEvents, error) {
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

	// Check if user exists (if no calendars found, user might not exist)
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		s.logger.Error("User not found", zap.Uint64("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("user not found")
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
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

	// Group events by calendar ID and filter for public visibility
	eventsByCalendar := make(map[uint64][]*model.CalendarEvent)
	calendarMap := make(map[uint64]*model.Calendar)

	// Create calendar lookup map
	for _, calendar := range calendars {
		calendarMap[calendar.ID] = calendar
	}

	// Filter events for public visibility
	for _, event := range events {
		calendar := calendarMap[event.CalendarID]
		if calendar == nil {
			continue
		}

		// Check if event is public based on visibility rules
		isPublic := false
		switch event.Visibility {
		case model.CalendarEventVisibilityPublic:
			// Event is explicitly public
			isPublic = true
		case model.CalendarEventVisibilityInherited:
			// Event inherits calendar visibility
			isPublic = calendar.Visibility == model.CalendarVisibilityPublic
		case model.CalendarEventVisibilityPrivate:
			// Event is explicitly private
			isPublic = false
		}

		if isPublic {
			eventsByCalendar[event.CalendarID] = append(eventsByCalendar[event.CalendarID], event)
		}
	}

	// Create nested response structure with only calendars that have public events
	var calendarsWithEvents []*model.CalendarWithEvents
	for _, calendar := range calendars {
		calendarEvents := eventsByCalendar[calendar.ID]

		// Only include calendars that have public events or are explicitly public
		if len(calendarEvents) > 0 || calendar.Visibility == model.CalendarVisibilityPublic {
			if calendarEvents == nil {
				calendarEvents = []*model.CalendarEvent{}
			}

			// Apply calendar event redaction to event titles
			s.applyEventRedaction(calendarEvents, calendar)

			calendarWithEvents := &model.CalendarWithEvents{
				Calendar: calendar,
				Events:   calendarEvents,
			}
			calendarsWithEvents = append(calendarsWithEvents, calendarWithEvents)
		}
	}

	// Count total public events
	totalPublicEvents := 0
	for _, calendarWithEvents := range calendarsWithEvents {
		totalPublicEvents += len(calendarWithEvents.Events)
	}

	s.logger.Info("Successfully retrieved public calendar events",
		zap.Uint64("user_id", userID),
		zap.Int("calendar_count", len(calendarsWithEvents)),
		zap.Int("total_public_events", totalPublicEvents),
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime))

	return calendarsWithEvents, nil
}

// applyEventRedaction applies the calendar's event redaction to event titles if redaction is set
func (s *CalendarService) applyEventRedaction(events []*model.CalendarEvent, calendar *model.Calendar) {
	// Only apply redaction if it's set and not empty
	if calendar.EventRedaction == nil || *calendar.EventRedaction == "" {
		return
	}

	// Apply redaction as prefix to each event title
	redaction := *calendar.EventRedaction
	for _, event := range events {
		if event.Title != "" {
			event.Title = fmt.Sprintf("[%s] %s", redaction, event.Title)
		}
	}
}
