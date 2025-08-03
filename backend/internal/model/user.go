package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
// @Description User account information
type User struct {
	ID          uint64         `json:"id,string" gorm:"primaryKey" example:"123456789"`        // Unique user identifier
	Username    string         `json:"username" gorm:"uniqueIndex;not null" example:"johndoe"` // Username
	DisplayName string         `json:"display_name" gorm:"not null" example:"John Doe"`        // User's display name
	Password    *string        `json:"password,omitempty"`                                     // Password hash (excluded from responses)
	Picture     *string        `json:"picture" example:"https://example.com/avatar.jpg"`       // Profile picture URL
	CreatedAt   time.Time      `json:"created_at" example:"2024-01-01T00:00:00Z"`              // Account creation timestamp
	UpdatedAt   time.Time      `json:"updated_at" example:"2024-01-01T00:00:00Z"`              // Last update timestamp
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`                                         // Soft delete timestamp (excluded from responses)
	Accounts    []Account      `json:"accounts,omitempty" gorm:"foreignKey:UserID"`            // Associated OAuth accounts
}

// Account represents an OAuth account linked to a user
// @Description OAuth account information
type Account struct {
	ID           uint64         `json:"id,string" gorm:"primaryKey"`
	UserID       uint64         `json:"user_id,string" gorm:"not null"`
	Provider     string         `json:"provider" gorm:"not null"`    // e.g. "google", "github"
	ProviderID   string         `json:"provider_id" gorm:"not null"` // e.g. Google sub, GitHub ID
	Email        *string        `json:"email,omitempty"`             // e.g. example@gmail.com
	AccessToken  *string        `json:"-" gorm:"type:text"`          // NEVER expose
	RefreshToken *string        `json:"-" gorm:"type:text"`          // NEVER expose
	Expiry       *time.Time     `json:"expiry,omitempty"`            // Access token expiry
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
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

// UserProfileResponse represents the response for user profile endpoint
// @Description User profile response
type UserProfileResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"User profile retrieved successfully"`
	User    *User  `json:"user"`
}

// PublicUserProfile represents public user information
// @Description Public user profile information
type PublicUserProfile struct {
	ID          uint64    `json:"id,string" example:"123456789"`                        // Unique user identifier
	Username    string    `json:"username" example:"johndoe"`                           // Username
	DisplayName string    `json:"display_name" example:"John Doe"`                      // User's display name
	Picture     *string   `json:"picture" example:"https://example.com/avatar.jpg"`     // Profile picture URL
	CreatedAt   time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`            // Account creation timestamp
}

// PublicUserProfileResponse represents the response for public user profile endpoint
// @Description Public user profile response
type PublicUserProfileResponse struct {
	Success bool               `json:"success" example:"true"`
	Message string             `json:"message" example:"Public user profile retrieved successfully"`
	User    *PublicUserProfile `json:"user"`
}

// UpdateUserProfileRequest represents the request for updating user profile
// @Description User profile update request
type UpdateUserProfileRequest struct {
	Username    *string `json:"username,omitempty" example:"newusername"`    // New username (optional)
	DisplayName *string `json:"display_name,omitempty" example:"New Name"`   // New display name (optional)
}

// UpdateUserProfileResponse represents the response for updating user profile
// @Description User profile update response
type UpdateUserProfileResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"User profile updated successfully"`
	User    *User  `json:"user"`
}
