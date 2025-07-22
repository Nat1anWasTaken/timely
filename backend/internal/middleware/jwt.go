package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/NathanWasTaken/timely/backend/internal/model"
	"github.com/NathanWasTaken/timely/backend/pkg/utils"
)

// UserContextKey is the key used to store user information in the request context
type UserContextKey string

const (
	// UserContextKeyValue is the context key value for user information
	UserContextKeyValue UserContextKey = "user"
)

// UserInfo represents the user information stored in the request context
type UserInfo struct {
	ID    uint64 `json:"user_id"`
	Email string `json:"email"`
}

// JWTMiddleware creates a middleware that validates JWT tokens
func JWTMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Extract token from cookie first, then from Authorization header
			token := extractTokenFromRequest(r)
			if token == "" {
				sendAuthError(w, "No authentication token provided", http.StatusUnauthorized, logger)
				return
			}

			// Validate the token
			claims, err := utils.ValidateJWT(token)
			if err != nil {
				logger.Error("JWT validation failed", zap.Error(err))
				sendAuthError(w, "Invalid or expired authentication token", http.StatusUnauthorized, logger)
				return
			}

			// Create user info from claims
			userInfo := &UserInfo{
				ID:    claims.UserID,
				Email: claims.Email,
			}

			// Add user info to request context
			ctx := context.WithValue(r.Context(), UserContextKeyValue, userInfo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}


// extractTokenFromRequest extracts JWT token from cookie or Authorization header
func extractTokenFromRequest(r *http.Request) string {
	// First try to get token from cookie
	if cookie, err := r.Cookie("access_token"); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Then try to get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	return utils.ExtractTokenFromHeader(authHeader)
}

// sendAuthError sends an authentication error response
func sendAuthError(w http.ResponseWriter, message string, statusCode int, logger *zap.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := model.ErrorResponse{
		Success: false,
		Message: message,
		Error:   "authentication_required",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode auth error response", zap.Error(err))
	}
}

// GetUserFromContext extracts user information from the request context
func GetUserFromContext(ctx context.Context) (*UserInfo, bool) {
	user, ok := ctx.Value(UserContextKeyValue).(*UserInfo)
	return user, ok
}

// RequireAuth is a helper function that can be used in handlers to ensure authentication
func RequireAuth(w http.ResponseWriter, r *http.Request, logger *zap.Logger) (*UserInfo, bool) {
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		sendAuthError(w, "Authentication required", http.StatusUnauthorized, logger)
		return nil, false
	}
	return user, true
}
