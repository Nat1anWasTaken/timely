package service

import (
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/NathanWasTaken/timely/backend/internal/model"
	"github.com/NathanWasTaken/timely/backend/internal/repository"
	"github.com/NathanWasTaken/timely/backend/pkg/utils"
)

type UserService struct {
	userRepo *repository.UserRepository
	logger   *zap.Logger
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		logger:   zap.L(),
	}
}

// FindOrCreateGoogleUser finds an existing user by Google ID or email, or creates a new one
func (s *UserService) FindOrCreateGoogleUser(googleUser *model.GoogleUserInfo) (*model.User, error) {
	// First try to find by Google ID
	if user, err := s.userRepo.FindByGoogleID(googleUser.ID); err == nil {
		s.logger.Info("Found existing user by Google ID", zap.String("email", user.Email))
		return user, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Then try to find by email (user might exist but not linked to Google)
	if user, err := s.userRepo.FindByEmail(googleUser.Email); err == nil {
		// Link Google ID to existing user
		user.GoogleID = &googleUser.ID
		if user.Picture == nil && googleUser.Picture != "" {
			user.Picture = &googleUser.Picture
		}
		user.UpdatedAt = time.Now()

		if err := s.userRepo.Update(user); err != nil {
			return nil, err
		}

		s.logger.Info("Linked Google account to existing user", zap.String("email", user.Email))
		return user, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new user
	user := &model.User{
		ID:          utils.GenerateID(),
		Email:       googleUser.Email,
		Username:    s.generateUniqueUsername(googleUser.Name),
		DisplayName: googleUser.Name,
		Picture:     &googleUser.Picture,
		GoogleID:    &googleUser.ID,
		Password:    nil, // OAuth users don't have passwords
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	s.logger.Info("Created new user from Google OAuth", zap.String("email", user.Email))
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(email string) (*model.User, error) {
	return s.userRepo.FindByEmail(email)
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id uint64) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

// generateUniqueUsername generates a unique username from the given name
func (s *UserService) generateUniqueUsername(name string) string {
	baseUsername := name
	// TODO: if username exists, append a number
	// In a real application, you might want more sophisticated logic

	// For now, just use the name and let GORM handle uniqueness constraints
	// If there's a conflict, it will error and we can handle it appropriately
	return baseUsername
}
