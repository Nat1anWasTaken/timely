package auth

import (
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"

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
// @Success 307 "Redirect to Google OAuth consent page"
// @Failure 500 "Internal server error"
// @Router /api/auth/google/login [get]
func (h *GoogleOAuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate state parameter for CSRF protection
	state, err := h.oauthService.GenerateStateOauthCookie()
	if err != nil {
		h.logger.Error("Failed to generate state", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set state cookie
	cookie := &http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(10 * time.Minute),
		Path:     "/",
	}
	http.SetCookie(w, cookie)

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
	// Verify state parameter
	stateCookie, err := r.Cookie("oauthstate")
	if err != nil {
		h.logger.Error("State cookie not found", zap.Error(err))
		http.Error(w, "State cookie not found", http.StatusBadRequest)
		return
	}

	stateParam := r.URL.Query().Get("state")
	if stateParam != stateCookie.Value {
		h.logger.Error("Invalid state parameter")
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Clear the state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauthstate",
		Value:    "",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(-time.Hour),
		Path:     "/",
	})

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
	jwtToken, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		h.logger.Error("Failed to generate JWT token", zap.Error(err))
		http.Error(w, "Failed to generate authentication token", http.StatusInternalServerError)
		return
	}

	// Set JWT as HttpOnly cookie
	jwtCookie := utils.CreateJWTCookie(jwtToken)
	http.SetCookie(w, jwtCookie)

	// Get client URL for redirect
	clientURL := os.Getenv("FRONTEND_DOMAIN")
	if clientURL == "" {
		clientURL = "http://localhost:3000" // Default for development
	}

	// Ensure clientURL ends without trailing slash
	clientURL = strings.TrimSuffix(clientURL, "/")

	dashboardURL := clientURL + "/dashboard"

	// Redirect to client dashboard
	http.Redirect(w, r, dashboardURL, http.StatusTemporaryRedirect)

	h.logger.Info("User successfully authenticated",
		zap.String("email", user.Email),
		zap.String("name", user.Username))
}
