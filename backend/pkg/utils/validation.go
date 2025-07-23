package utils

import (
	"regexp"
	"strings"
)

// ValidateUsername checks if a username contains only lowercase letters (a-z), numbers (0-9), underscore (_), and dot (.)
func ValidateUsername(username string) bool {
	if username == "" {
		return false
	}

	// Regular expression to match only a-z, 0-9, _, and .
	validUsernamePattern := regexp.MustCompile(`^[a-z0-9_.]+$`)
	return validUsernamePattern.MatchString(username)
}

// SanitizeUsername converts a string to a valid username format
// It converts to lowercase and removes invalid characters, keeping only a-z, 0-9, _, and .
func SanitizeUsername(input string) string {
	if input == "" {
		return ""
	}

	// Convert to lowercase first
	input = strings.ToLower(input)

	// Remove invalid characters, keeping only a-z, 0-9, _, and .
	sanitizePattern := regexp.MustCompile(`[^a-z0-9_.]`)
	input = sanitizePattern.ReplaceAllString(input, "")

	return input
}
