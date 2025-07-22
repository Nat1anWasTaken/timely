package repository

import (
	"gorm.io/gorm"

	"github.com/NathanWasTaken/timely/backend/internal/model"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// FindByEmail finds a user by email address
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByGoogleID finds a user by Google ID
func (r *UserRepository) FindByGoogleID(googleID string) (*model.User, error) {
	var user model.User
	err := r.db.Where("google_id = ?", googleID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create creates a new user
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// Update updates an existing user
func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id uint64) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername finds a user by username
func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ExistsByEmail checks if a user with the given email exists
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// ExistsByUsername checks if a user with the given username exists
func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// GoogleToken repository methods

// CreateOrUpdateGoogleToken creates or updates a Google OAuth token for a user
func (r *UserRepository) CreateOrUpdateGoogleToken(token *model.GoogleToken) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Check if token already exists for this user
		var existingToken model.GoogleToken
		err := tx.Where("user_id = ?", token.UserID).First(&existingToken).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// Token doesn't exist, create new one
				return tx.Create(token).Error
			}
			// Return other database errors
			return err
		}

		// Token exists, update it
		existingToken.RefreshToken = token.RefreshToken
		existingToken.AccessToken = token.AccessToken
		existingToken.ExpiresAt = token.ExpiresAt
		// Let GORM handle UpdatedAt automatically
		return tx.Save(&existingToken).Error
	})
}

// FindGoogleTokenByUserID finds a Google OAuth token by user ID
func (r *UserRepository) FindGoogleTokenByUserID(userID uint64) (*model.GoogleToken, error) {
	var token model.GoogleToken
	err := r.db.Where("user_id = ?", userID).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// DeleteGoogleTokenByUserID deletes a Google OAuth token by user ID
func (r *UserRepository) DeleteGoogleTokenByUserID(userID uint64) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.GoogleToken{}).Error
}
