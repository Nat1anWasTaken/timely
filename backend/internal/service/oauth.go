package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/NathanWasTaken/timely/backend/internal/config"
	"github.com/NathanWasTaken/timely/backend/internal/model"
)

type OAuthService struct {
	config      *config.OAuthConfig
	userService *UserService
}

func NewOAuthService(config *config.OAuthConfig, userService *UserService) *OAuthService {
	return &OAuthService{
		config:      config,
		userService: userService,
	}
}

// GenerateStateOauthCookie generates a random state string for OAuth security
func (s *OAuthService) GenerateStateOauthCookie() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
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

// GetUserGoogleToken retrieves the stored Google OAuth token for a user
func (s *OAuthService) GetUserGoogleToken(userID uint64) (*model.GoogleToken, error) {
	return s.userService.GetGoogleTokenByUserID(userID)
}
