package auth

import (
	"encoding/json"
	"net/http"
)

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"message": "Login endpoints available",
		"endpoints": map[string]interface{}{
			"google_oauth": map[string]string{
				"url":         "/auth/google/login",
				"method":      "GET",
				"description": "Start Google OAuth login flow",
			},
			"traditional": map[string]string{
				"url":         "/auth/login",
				"method":      "POST",
				"description": "Traditional email/password login (not implemented yet)",
			},
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
