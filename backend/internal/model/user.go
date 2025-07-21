package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          uint64         `json:"id,string" gorm:"primaryKey"`
	Email       string         `json:"email" gorm:"uniqueIndex;not null"`
	Username    string         `json:"username" gorm:"uniqueIndex;not null"`
	DisplayName string         `json:"display_name" gorm:"not null"`
	Password    *string        `json:"password,omitempty"`
	Picture     *string        `json:"picture"`
	GoogleID    *string        `json:"google_id,omitempty" gorm:"uniqueIndex"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type GoogleUserInfo struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Picture    string `json:"picture"`
	Locale     string `json:"locale"`
}
