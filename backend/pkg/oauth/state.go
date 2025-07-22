package oauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// StatePayload represents the data structure for OAuth state parameter
type StatePayload struct {
	CSRF    string    `json:"csrf"`    // CSRF token for security
	Mode    string    `json:"mode"`    // OAuth mode: "login", "link", etc.
	From    string    `json:"from"`    // User ID for linking or redirect page for login
	Created time.Time `json:"created"` // Timestamp for expiration checks
	Nonce   string    `json:"nonce"`   // Random nonce for uniqueness
}

// StateManager handles OAuth state parameter generation and verification
type StateManager struct {
	secret []byte
}

// NewStateManager creates a new state manager with secret from environment
func NewStateManager() (*StateManager, error) {
	secret := os.Getenv("OAUTH_STATE_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("OAUTH_STATE_SECRET environment variable is required")
	}

	// Ensure secret is at least 32 bytes for security
	if len(secret) < 32 {
		return nil, fmt.Errorf("OAUTH_STATE_SECRET must be at least 32 characters long")
	}

	return &StateManager{
		secret: []byte(secret),
	}, nil
}

// GenerateState creates a signed state parameter from a payload
func (sm *StateManager) GenerateState(payload *StatePayload) (string, error) {
	// Set creation time if not provided
	if payload.Created.IsZero() {
		payload.Created = time.Now()
	}

	// Marshal payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal state payload: %w", err)
	}

	// Create HMAC signature
	h := hmac.New(sha256.New, sm.secret)
	h.Write(payloadBytes)
	signature := h.Sum(nil)

	// Combine payload and signature
	combined := append(payloadBytes, signature...)

	// Base64 URL encode the result
	encoded := base64.URLEncoding.EncodeToString(combined)

	return encoded, nil
}

// VerifyAndDecodeState verifies the signature and decodes the state parameter
func (sm *StateManager) VerifyAndDecodeState(encodedState string) (*StatePayload, error) {
	// Base64 URL decode
	combined, err := base64.URLEncoding.DecodeString(encodedState)
	if err != nil {
		return nil, fmt.Errorf("invalid state encoding: %w", err)
	}

	// Check minimum length (JSON payload + 32-byte signature)
	if len(combined) < 64 {
		return nil, fmt.Errorf("state parameter too short")
	}

	// Split payload and signature
	payloadBytes := combined[:len(combined)-32]
	signature := combined[len(combined)-32:]

	// Verify HMAC signature
	h := hmac.New(sha256.New, sm.secret)
	h.Write(payloadBytes)
	expectedSignature := h.Sum(nil)

	if !hmac.Equal(signature, expectedSignature) {
		return nil, fmt.Errorf("invalid state signature")
	}

	// Unmarshal payload
	var payload StatePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state payload: %w", err)
	}

	return &payload, nil
}

// GenerateSimpleState generates a simple random state for backward compatibility
func (sm *StateManager) GenerateSimpleState() (string, error) {
	// Generate a simple payload with just CSRF and nonce
	payload := &StatePayload{
		CSRF:    generateRandomString(16),
		Mode:    "login",
		From:    "",
		Created: time.Now(),
	}

	return sm.GenerateState(payload)
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// IsStateExpired checks if a state payload has expired
func (sm *StateManager) IsStateExpired(payload *StatePayload, maxAge time.Duration) bool {
	return time.Since(payload.Created) > maxAge
}

// CreateLoginState creates a state for login flow with redirect page
func (sm *StateManager) CreateLoginState(redirectPage string) (string, error) {
	payload := &StatePayload{
		CSRF:    generateRandomString(16),
		Mode:    "login",
		From:    redirectPage, // Redirect page for login flow
		Created: time.Now(),
		Nonce:   generateRandomString(8),
	}

	return sm.GenerateState(payload)
}

// CreateLinkState creates a state for account linking flow with user ID
func (sm *StateManager) CreateLinkState(userID string) (string, error) {
	payload := &StatePayload{
		CSRF:    generateRandomString(16),
		Mode:    "link",
		From:    userID, // User ID for linking flow
		Created: time.Now(),
		Nonce:   generateRandomString(8),
	}

	return sm.GenerateState(payload)
}
