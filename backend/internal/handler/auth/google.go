package auth

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/NathanWasTaken/timely/backend/internal/middleware"
	"github.com/NathanWasTaken/timely/backend/internal/service"
	"github.com/NathanWasTaken/timely/backend/pkg/utils"
)

type GoogleOAuthHandler struct {
	oauthService *service.OAuthService
	userService  *service.UserService
	logger       *zap.Logger
}

func NewGoogleOAuthHandler(oauthService *service.OAuthService, userService *service.UserService) *GoogleOAuthHandler {
	return &GoogleOAuthHandler{
		oauthService: oauthService,
		userService:  userService,
		logger:       zap.L(),
	}
}

// GoogleLogin initiates Google OAuth flow
// @Summary Initiate Google OAuth Login
// @Description Redirects user to Google's OAuth consent page to begin authentication process
// @Tags OAuth
// @Produce html
// @Param mode query string false "OAuth mode: login or link"
// @Param from query string false "Original redirect page (for login mode)"
// @Success 307 "Redirect to Google OAuth consent page"
// @Failure 400 "Bad request - Authentication required for link mode"
// @Failure 500 "Internal server error"
// @Router /api/auth/google/login [get]
func (h *GoogleOAuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "login" // Default mode
	}

	from := r.URL.Query().Get("from")

	// For link mode, require authentication
	if mode == "link" {
		user, ok := middleware.GetUserFromContext(r.Context())
		if !ok {
			h.logger.Error("Authentication required for account linking")
			http.Error(w, "Authentication required for account linking", http.StatusBadRequest)
			return
		}

		// Use the authenticated user's ID for linking
		from = fmt.Sprintf("%d", user.ID)
		h.logger.Info("Account linking initiated", zap.Uint64("user_id", user.ID))
	}

	// Generate secure state parameter with payload
	state, err := h.oauthService.GenerateStateWithPayload(mode, from)
	if err != nil {
		h.logger.Error("Failed to generate state", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get Google OAuth URL
	url := h.oauthService.GetGoogleLoginURL(state)

	// Redirect to Google
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles Google OAuth callback
// @Summary Google OAuth Callback
// @Description Handles the callback from Google OAuth, exchanges code for user info and creates/updates user account
// @Tags OAuth
// @Accept json
// @Produce json
// @Param code query string true "Authorization code from Google"
// @Param state query string true "State parameter for CSRF protection"
// @Success 200 {object} map[string]interface{} "Authentication successful with user data"
// @Failure 400 "Bad request - Missing state cookie, invalid state, or missing authorization code"
// @Failure 500 "Internal server error - Token exchange or user processing failed"
// @Router /api/auth/google/callback [get]
func (h *GoogleOAuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Get state parameter from URL
	stateParam := r.URL.Query().Get("state")
	if stateParam == "" {
		h.logger.Error("State parameter not found in URL")
		http.Error(w, "State parameter not found", http.StatusBadRequest)
		return
	}

	// Verify and decode state parameter
	statePayload, err := h.oauthService.VerifyState(stateParam)
	if err != nil {
		h.logger.Error("Invalid state parameter", zap.Error(err))
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	h.logger.Info("State verification successful",
		zap.String("mode", statePayload.Mode),
		zap.String("from", statePayload.From))

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		h.logger.Error("Authorization code not found")
		http.Error(w, "Authorization code not found", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := h.oauthService.ExchangeCodeForToken(r.Context(), code)
	if err != nil {
		h.logger.Error("Failed to exchange code for token", zap.Error(err))
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	// Get user info from Google
	googleUser, err := h.oauthService.GetUserInfoFromGoogle(r.Context(), token)
	if err != nil {
		h.logger.Error("Failed to get user info from Google", zap.Error(err))
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	// Find or create user in database
	user, err := h.oauthService.FindOrCreateUserFromGoogleWithToken(googleUser, token)
	if err != nil {
		h.logger.Error("Failed to find or create user", zap.Error(err))
		http.Error(w, "Failed to process user", http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	jwtToken, err := utils.GenerateJWT(user.ID)
	if err != nil {
		h.logger.Error("Failed to generate JWT token", zap.Error(err))
		http.Error(w, "Failed to generate authentication token", http.StatusInternalServerError)
		return
	}

	// Set JWT as HttpOnly cookie
	jwtCookie := utils.CreateJWTCookie(jwtToken)
	http.SetCookie(w, jwtCookie)

	// Handle different OAuth modes
	switch statePayload.Mode {
	case "link":
		// For linking, we need to link the Google account to the existing user
		userIDStr := statePayload.From
		if userIDStr == "" {
			h.logger.Error("User ID not found in link state")
			http.Error(w, "Invalid link state", http.StatusBadRequest)
			return
		}

		// Parse user ID
		var userID uint64
		if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
			h.logger.Error("Invalid user ID in link state", zap.String("user_id", userIDStr))
			http.Error(w, "Invalid user ID in link state", http.StatusBadRequest)
			return
		}

		// Find the existing user
		existingUser, err := h.userService.GetUserByID(userID)
		if err != nil {
			h.logger.Error("Failed to find user for linking", zap.Uint64("user_id", userID), zap.Error(err))
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Check if Google account is already linked
		accounts, err := h.userService.GetAccountsByUserID(userID)
		if err != nil {
			h.logger.Error("Failed to check existing accounts", zap.Uint64("user_id", userID), zap.Error(err))
			http.Error(w, "Failed to check account status", http.StatusInternalServerError)
			return
		}

		googleLinked := false
		for _, account := range accounts {
			if account.Provider == "google" {
				googleLinked = true
				break
			}
		}

		if googleLinked {
			h.logger.Error("Google account already linked", zap.Uint64("user_id", userID))
			http.Error(w, "Google account already linked", http.StatusConflict)
			return
		}

		// Link the Google account to the existing user
		if err := h.userService.LinkGoogleAccount(userID, googleUser.ID, googleUser.Email); err != nil {
			h.logger.Error("Failed to link Google account", zap.Uint64("user_id", userID), zap.Error(err))
			http.Error(w, "Failed to link Google account", http.StatusInternalServerError)
			return
		}

		// Update user picture if not set
		if existingUser.Picture == nil && googleUser.Picture != "" {
			existingUser.Picture = &googleUser.Picture
			existingUser.UpdatedAt = time.Now()
			if err := h.userService.UpdateUser(existingUser); err != nil {
				h.logger.Error("Failed to update user picture", zap.Uint64("user_id", userID), zap.Error(err))
				// Don't fail the linking if picture update fails, just log it
			}
		}

		// Store the OAuth token for the linked account
		if err := h.userService.UpdateGoogleAccountTokens(userID, token); err != nil {
			h.logger.Error("Failed to store Google token for linked account", zap.Uint64("user_id", userID), zap.Error(err))
			// Don't fail the linking if token storage fails, just log it
		}

		h.logger.Info("Google account successfully linked",
			zap.Uint64("user_id", userID),
			zap.String("google_email", googleUser.Email))

		// Get client URL for redirect
		clientURL := os.Getenv("FRONTEND_DOMAIN")
		if clientURL == "" {
			clientURL = "http://localhost:3000" // Default for development
		}
		clientURL = strings.TrimSuffix(clientURL, "/")
		redirectURL := clientURL + "/dashboard"

		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)

	case "login":
		fallthrough
	default:
		// For login, use 'from' as redirect page
		var redirectURL string

		// Get client URL for redirect
		clientURL := os.Getenv("FRONTEND_DOMAIN")
		if clientURL == "" {
			clientURL = "http://localhost:3000" // Default for development
		}

		// Ensure clientURL ends without trailing slash
		clientURL = strings.TrimSuffix(clientURL, "/")

		// Use 'from' parameter if provided and valid, otherwise default to dashboard
		if statePayload.From != "" && strings.HasPrefix(statePayload.From, "/") {
			redirectURL = clientURL + statePayload.From
		} else {
			redirectURL = clientURL + "/dashboard"
		}

		// Redirect to appropriate page
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}

	h.logger.Info("User successfully authenticated",
		zap.String("username", user.Username),
		zap.String("display_name", user.DisplayName))
}
