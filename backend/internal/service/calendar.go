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
)

type CalendarService struct {
	userRepo    *repository.UserRepository
	oauthConfig *config.OAuthConfig
	logger      *zap.Logger
}

func NewCalendarService(userRepo *repository.UserRepository, oauthConfig *config.OAuthConfig) *CalendarService {
	return &CalendarService{
		userRepo:    userRepo,
		oauthConfig: oauthConfig,
		logger:      zap.L(),
	}
}

// GetUserCalendars retrieves all calendars for a user from Google Calendar API
func (s *CalendarService) GetUserCalendars(userID uint64) ([]*model.GoogleCalendar, error) {
	// Get user's Google token
	token, err := s.userRepo.FindGoogleTokenByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Google token: %w", err)
	}

	// Check if token needs refresh
	if err := s.refreshTokenIfNeeded(token); err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Fetch calendars from Google API
	calendars, err := s.fetchCalendarsFromGoogle(token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch calendars from Google: %w", err)
	}

	return calendars, nil
}

// refreshTokenIfNeeded checks if the token is expired and refreshes it if necessary
func (s *CalendarService) refreshTokenIfNeeded(token *model.GoogleToken) error {
	// Check if token is expired (with 5 minute buffer)
	if time.Now().Add(5 * time.Minute).Before(token.ExpiresAt) {
		return nil // Token is still valid
	}

	s.logger.Info("Refreshing expired Google token", zap.Uint64("user_id", token.UserID))

	// Create oauth2.Token for refresh
	oauthToken := &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.ExpiresAt,
	}

	// Refresh the token using the injected OAuth config
	ctx := context.Background()
	newToken, err := s.oauthConfig.Google.TokenSource(ctx, oauthToken).Token()
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update the token in database
	token.AccessToken = newToken.AccessToken
	token.RefreshToken = newToken.RefreshToken
	token.ExpiresAt = newToken.Expiry
	token.UpdatedAt = time.Now()

	if err := s.userRepo.CreateOrUpdateGoogleToken(token); err != nil {
		return fmt.Errorf("failed to update token in database: %w", err)
	}

	s.logger.Info("Successfully refreshed Google token", zap.Uint64("user_id", token.UserID))
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
