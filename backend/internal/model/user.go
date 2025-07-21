package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
// @Description User account information
type User struct {
	ID          uint64         `json:"id,string" gorm:"primaryKey" example:"123456789"`              // Unique user identifier
	Email       string         `json:"email" gorm:"uniqueIndex;not null" example:"user@example.com"` // User's email address
	Username    string         `json:"username" gorm:"uniqueIndex;not null" example:"johndoe"`       // Username
	DisplayName string         `json:"display_name" gorm:"not null" example:"John Doe"`              // User's display name
	Password    *string        `json:"password,omitempty"`                                           // Password hash (excluded from responses)
	Picture     *string        `json:"picture" example:"https://example.com/avatar.jpg"`             // Profile picture URL
	GoogleID    *string        `json:"google_id,omitempty" gorm:"uniqueIndex" example:"123456789"`   // Google OAuth ID
	CreatedAt   time.Time      `json:"created_at" example:"2024-01-01T00:00:00Z"`                    // Account creation timestamp
	UpdatedAt   time.Time      `json:"updated_at" example:"2024-01-01T00:00:00Z"`                    // Last update timestamp
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`                                               // Soft delete timestamp (excluded from responses)
}

// GoogleUserInfo represents user information from Google OAuth
// @Description Google OAuth user information
type GoogleUserInfo struct {
	ID         string `json:"id" example:"123456789"`                                  // Google user ID
	Email      string `json:"email" example:"user@gmail.com"`                          // Google account email
	Name       string `json:"name" example:"John Doe"`                                 // Full name
	GivenName  string `json:"given_name" example:"John"`                               // First name
	FamilyName string `json:"family_name" example:"Doe"`                               // Last name
	Picture    string `json:"picture" example:"https://lh3.googleusercontent.com/..."` // Profile picture URL
	Locale     string `json:"locale" example:"en"`                                     // User's locale
}
