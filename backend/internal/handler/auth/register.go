package auth

import (
	"encoding/json"
	"net/http"
)

func Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"message": "Registration endpoints available",
		"endpoints": map[string]interface{}{
			"google_oauth": map[string]string{
				"url":         "/auth/google/login",
				"method":      "GET",
				"description": "Start Google OAuth registration flow",
			},
			"traditional": map[string]string{
				"url":         "/auth/register",
				"method":      "POST",
				"description": "Traditional email/password registration (not implemented yet)",
			},
		},
		"oauth_flow": map[string]string{
			"step_1": "Visit /auth/google/login to start OAuth flow",
			"step_2": "User will be redirected to Google for authentication",
			"step_3": "Google will redirect back to /auth/google/callback",
			"step_4": "User will be registered and authenticated",
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
