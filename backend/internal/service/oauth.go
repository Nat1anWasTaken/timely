package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/NathanWasTaken/timely/backend/internal/config"
	"github.com/NathanWasTaken/timely/backend/internal/model"
	"github.com/NathanWasTaken/timely/backend/pkg/oauth"
)

type OAuthService struct {
	config       *config.OAuthConfig
	userService  *UserService
	stateManager *oauth.StateManager
}

func NewOAuthService(config *config.OAuthConfig, userService *UserService) (*OAuthService, error) {
	stateManager, err := oauth.NewStateManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create state manager: %w", err)
	}

	return &OAuthService{
		config:       config,
		userService:  userService,
		stateManager: stateManager,
	}, nil
}

// GenerateStateOauthCookie generates a secure signed state string for OAuth security
func (s *OAuthService) GenerateStateOauthCookie() (string, error) {
	return s.stateManager.GenerateSimpleState()
}

// GenerateStateWithPayload generates a secure signed state with custom payload
func (s *OAuthService) GenerateStateWithPayload(mode, from string) (string, error) {
	switch mode {
	case "login":
		return s.stateManager.CreateLoginState(from) // from = redirect page
	case "link":
		return s.stateManager.CreateLinkState(from) // from = user ID
	default:
		return s.stateManager.CreateLoginState(from)
	}
}

// VerifyState verifies and decodes a state parameter
func (s *OAuthService) VerifyState(encodedState string) (*oauth.StatePayload, error) {
	payload, err := s.stateManager.VerifyAndDecodeState(encodedState)
	if err != nil {
		return nil, err
	}

	// Check if state has expired (10 minutes)
	if s.stateManager.IsStateExpired(payload, 10*time.Minute) {
		return nil, fmt.Errorf("state parameter has expired")
	}

	return payload, nil
}

// GetGoogleLoginURL returns the Google OAuth login URL with state parameter
func (s *OAuthService) GetGoogleLoginURL(state string) string {
	return s.config.Google.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
}

// ExchangeCodeForToken exchanges the authorization code for an access token
func (s *OAuthService) ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.config.Google.Exchange(ctx, code)
}

// GetUserInfoFromGoogle fetches user information from Google API using the access token
func (s *OAuthService) GetUserInfoFromGoogle(ctx context.Context, token *oauth2.Token) (*model.GoogleUserInfo, error) {
	client := s.config.Google.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var userInfo model.GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info: %w", err)
	}

	return &userInfo, nil
}

// FindOrCreateUserFromGoogle finds or creates a user from Google OAuth info
func (s *OAuthService) FindOrCreateUserFromGoogle(googleUser *model.GoogleUserInfo) (*model.User, error) {
	return s.userService.FindOrCreateGoogleUser(googleUser, nil)
}

// FindOrCreateUserFromGoogleWithToken finds or creates a user from Google OAuth info and stores the OAuth token
func (s *OAuthService) FindOrCreateUserFromGoogleWithToken(googleUser *model.GoogleUserInfo, token *oauth2.Token) (*model.User, error) {
	return s.userService.FindOrCreateGoogleUser(googleUser, token)
}

// GetUserGoogleAccount retrieves the stored Google account for a user
func (s *OAuthService) GetUserGoogleAccount(userID uint64) (*model.Account, error) {
	return s.userService.GetGoogleAccountByUserID(userID)
}
