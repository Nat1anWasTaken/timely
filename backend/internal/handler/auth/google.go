package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/NathanWasTaken/timely/backend/internal/service"
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

// GoogleLogin redirects the user to Google's OAuth consent page
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

// GoogleCallback handles the callback from Google OAuth
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
	user, err := h.oauthService.FindOrCreateUserFromGoogle(googleUser)
	if err != nil {
		h.logger.Error("Failed to find or create user", zap.Error(err))
		http.Error(w, "Failed to process user", http.StatusInternalServerError)
		return
	}

	// TODO: Create JWT token or session here
	// For now, we'll set a simple session cookie
	sessionCookie := &http.Cookie{
		Name:     "user_session",
		Value:    user.Email, // In production, use a proper session token
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/",
	}
	http.SetCookie(w, sessionCookie)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"success": true,
		"message": "Successfully authenticated with Google",
		"user":    user,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("User successfully authenticated",
		zap.String("email", user.Email),
		zap.String("name", user.Username))
}
