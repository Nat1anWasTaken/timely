package repository

import (
	"time"

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

// FindByEmail finds a user by email address through Account model
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Joins("JOIN accounts ON users.id = accounts.user_id").
		Where("accounts.provider = ? AND accounts.provider_id = ?", "email", email).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByGoogleID finds a user by Google ID through Account model
func (r *UserRepository) FindByGoogleID(googleID string) (*model.User, error) {
	var user model.User
	err := r.db.Joins("JOIN accounts ON users.id = accounts.user_id").
		Where("accounts.provider = ? AND accounts.provider_id = ?", "google", googleID).
		First(&user).Error
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

// ExistsByEmail checks if a user with the given email exists through Account model
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).
		Joins("JOIN accounts ON users.id = accounts.user_id").
		Where("accounts.provider = ? AND accounts.provider_id = ?", "email", email).
		Count(&count).Error
	return count > 0, err
}

// ExistsByUsername checks if a user with the given username exists
func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// Account repository methods

// CreateAccount creates a new account for a user
func (r *UserRepository) CreateAccount(account *model.Account) error {
	return r.db.Create(account).Error
}

// FindAccountByProviderAndID finds an account by provider and provider ID
func (r *UserRepository) FindAccountByProviderAndID(provider, providerID string) (*model.Account, error) {
	var account model.Account
	err := r.db.Where("provider = ? AND provider_id = ?", provider, providerID).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// FindAccountsByUserID finds all accounts for a user
func (r *UserRepository) FindAccountsByUserID(userID uint64) ([]model.Account, error) {
	var accounts []model.Account
	err := r.db.Where("user_id = ?", userID).Find(&accounts).Error
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// FindGoogleAccountByUserID finds a Google account for a user
func (r *UserRepository) FindGoogleAccountByUserID(userID uint64) (*model.Account, error) {
	var account model.Account
	err := r.db.Where("user_id = ? AND provider = ?", userID, "google").First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// UpdateGoogleAccountTokens updates the OAuth tokens for a Google account
func (r *UserRepository) UpdateGoogleAccountTokens(userID uint64, accessToken, refreshToken string, expiry *time.Time) error {
	return r.db.Model(&model.Account{}).
		Where("user_id = ? AND provider = ?", userID, "google").
		Updates(map[string]interface{}{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expiry":        expiry,
		}).Error
}
