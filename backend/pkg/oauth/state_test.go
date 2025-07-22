package oauth

import (
	"os"
	"testing"
	"time"
)

func TestStateManager(t *testing.T) {
	// Set up test environment
	os.Setenv("OAUTH_STATE_SECRET", "test-secret-key-that-is-long-enough-for-security-32-chars")

	// Create state manager
	sm, err := NewStateManager()
	if err != nil {
		t.Fatalf("Failed to create state manager: %v", err)
	}

	t.Run("Generate and verify simple state", func(t *testing.T) {
		// Generate simple state
		state, err := sm.GenerateSimpleState()
		if err != nil {
			t.Fatalf("Failed to generate simple state: %v", err)
		}

		// Verify state
		payload, err := sm.VerifyAndDecodeState(state)
		if err != nil {
			t.Fatalf("Failed to verify state: %v", err)
		}

		// Check payload fields
		if payload.Mode != "login" {
			t.Errorf("Expected mode 'login', got '%s'", payload.Mode)
		}
		if payload.CSRF == "" {
			t.Error("CSRF token should not be empty")
		}
		if payload.Nonce == "" {
			t.Error("Nonce should not be empty")
		}
		if payload.Created.IsZero() {
			t.Error("Created timestamp should not be zero")
		}
	})

	t.Run("Generate and verify login state", func(t *testing.T) {
		// Generate login state
		state, err := sm.CreateLoginState("/dashboard")
		if err != nil {
			t.Fatalf("Failed to generate login state: %v", err)
		}

		// Verify state
		payload, err := sm.VerifyAndDecodeState(state)
		if err != nil {
			t.Fatalf("Failed to verify state: %v", err)
		}

		// Check payload fields
		if payload.Mode != "login" {
			t.Errorf("Expected mode 'login', got '%s'", payload.Mode)
		}
		if payload.From != "/dashboard" {
			t.Errorf("Expected from '/dashboard', got '%s'", payload.From)
		}
	})

	t.Run("Generate and verify link state", func(t *testing.T) {
		// Generate link state
		state, err := sm.CreateLinkState("123456789")
		if err != nil {
			t.Fatalf("Failed to generate link state: %v", err)
		}

		// Verify state
		payload, err := sm.VerifyAndDecodeState(state)
		if err != nil {
			t.Fatalf("Failed to verify state: %v", err)
		}

		// Check payload fields
		if payload.Mode != "link" {
			t.Errorf("Expected mode 'link', got '%s'", payload.Mode)
		}
		if payload.From != "123456789" {
			t.Errorf("Expected from '123456789', got '%s'", payload.From)
		}
	})

	t.Run("Tampered state should fail", func(t *testing.T) {
		// Generate valid state
		state, err := sm.GenerateSimpleState()
		if err != nil {
			t.Fatalf("Failed to generate state: %v", err)
		}

		// Tamper with state (add a character)
		tamperedState := state + "x"

		// Verify tampered state should fail
		_, err = sm.VerifyAndDecodeState(tamperedState)
		if err == nil {
			t.Error("Tampered state should fail verification")
		}
	})

	t.Run("Expired state detection", func(t *testing.T) {
		// Create payload with old timestamp
		payload := &StatePayload{
			CSRF:    "test-csrf",
			Mode:    "login",
			From:    "",
			Created: time.Now().Add(-15 * time.Minute), // 15 minutes ago
			Nonce:   "test-nonce",
		}

		// Generate state
		state, err := sm.GenerateState(payload)
		if err != nil {
			t.Fatalf("Failed to generate state: %v", err)
		}

		// Verify state (should succeed)
		decodedPayload, err := sm.VerifyAndDecodeState(state)
		if err != nil {
			t.Fatalf("Failed to verify state: %v", err)
		}

		// Check if expired (10 minute threshold)
		if !sm.IsStateExpired(decodedPayload, 10*time.Minute) {
			t.Error("State should be expired")
		}
	})
}

func TestStateManagerErrors(t *testing.T) {
	t.Run("Missing secret should fail", func(t *testing.T) {
		// Clear environment variable
		os.Unsetenv("OAUTH_STATE_SECRET")

		_, err := NewStateManager()
		if err == nil {
			t.Error("Should fail when OAUTH_STATE_SECRET is not set")
		}
	})

	t.Run("Short secret should fail", func(t *testing.T) {
		// Set short secret
		os.Setenv("OAUTH_STATE_SECRET", "short")

		_, err := NewStateManager()
		if err == nil {
			t.Error("Should fail when OAUTH_STATE_SECRET is too short")
		}
	})
}
