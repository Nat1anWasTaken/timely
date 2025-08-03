package utils

import (
	"regexp"
	"strings"
)

// ValidateUsername checks if a username follows Instagram's rules:
// - Only letters (A-Z, case insensitive), numbers (0-9), underscore (_), and dot (.)
// - Cannot start or end with a dot
// - Cannot have consecutive dots
func ValidateUsername(username string) bool {
	if username == "" {
		return false
	}

	// Check for consecutive dots
	if strings.Contains(username, "..") {
		return false
	}

	// Check if starts or ends with dot
	if strings.HasPrefix(username, ".") || strings.HasSuffix(username, ".") {
		return false
	}

	// Regular expression to match only A-Z, a-z, 0-9, _, and .
	validUsernamePattern := regexp.MustCompile(`^[A-Za-z0-9_.]+$`)
	return validUsernamePattern.MatchString(username)
}

// SanitizeUsername converts a string to a valid username format following Instagram rules
// It removes invalid characters and fixes dot-related issues
func SanitizeUsername(input string) string {
	if input == "" {
		return ""
	}

	// Remove invalid characters, keeping only A-Z, a-z, 0-9, _, and .
	sanitizePattern := regexp.MustCompile(`[^A-Za-z0-9_.]`)
	input = sanitizePattern.ReplaceAllString(input, "")

	// Remove consecutive dots
	for strings.Contains(input, "..") {
		input = strings.ReplaceAll(input, "..", ".")
	}

	// Remove leading and trailing dots
	input = strings.Trim(input, ".")

	return input
}
