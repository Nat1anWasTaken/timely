package service

import (
	"errors"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"gorm.io/gorm"

	"github.com/NathanWasTaken/timely/backend/internal/model"
	"github.com/NathanWasTaken/timely/backend/internal/repository"
	"github.com/NathanWasTaken/timely/backend/pkg/encrypt"
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
func (s *UserService) FindOrCreateGoogleUser(googleUser *model.GoogleUserInfo, token *oauth2.Token) (*model.User, error) {
	// First try to find by Google ID
	if user, err := s.userRepo.FindByGoogleID(googleUser.ID); err == nil {
		s.logger.Info("Found existing user by Google ID", zap.String("email", user.Email))

		// Update OAuth token if provided
		if token != nil {
			if err := s.CreateOrUpdateGoogleToken(user.ID, token); err != nil {
				s.logger.Error("Failed to update Google token", zap.Error(err))
				// Don't fail the login if token storage fails, just log it
			}
		}

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

		// Store OAuth token if provided
		if token != nil {
			if err := s.CreateOrUpdateGoogleToken(user.ID, token); err != nil {
				s.logger.Error("Failed to store Google token", zap.Error(err))
				// Don't fail the login if token storage fails, just log it
			}
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

	// Store OAuth token if provided
	if token != nil {
		if err := s.CreateOrUpdateGoogleToken(user.ID, token); err != nil {
			s.logger.Error("Failed to store Google token for new user", zap.Error(err))
			// Don't fail the registration if token storage fails, just log it
		}
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

// CreateUser creates a new user with email/password authentication
func (s *UserService) CreateUser(req *model.RegisterRequest) (*model.User, error) {
	// Check if email already exists
	if exists, err := s.userRepo.ExistsByEmail(req.Email); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("user with this email already exists")
	}

	// Check if username already exists
	if exists, err := s.userRepo.ExistsByUsername(req.Username); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("username already taken")
	}

	// Hash the password
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &model.User{
		ID:          utils.GenerateID(),
		Email:       req.Email,
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Password:    &hashedPassword,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	s.logger.Info("Created new user", zap.String("email", user.Email), zap.String("username", user.Username))
	return user, nil
}

// AuthenticateUser authenticates a user with email/password
func (s *UserService) AuthenticateUser(req *model.LoginRequest) (*model.User, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	// Check if user has a password (not OAuth-only user)
	if user.Password == nil {
		return nil, errors.New("this account uses OAuth authentication")
	}

	// Verify password
	if !s.verifyPassword(req.Password, *user.Password) {
		return nil, errors.New("invalid email or password")
	}

	s.logger.Info("User authenticated successfully", zap.String("email", user.Email))
	return user, nil
}

// hashPassword hashes a password using Argon2
func (s *UserService) hashPassword(password string) (string, error) {
	return encrypt.HashPassword(password)
}

// verifyPassword verifies a password against its hash
func (s *UserService) verifyPassword(password, hash string) bool {
	return encrypt.VerifyPassword(password, hash)
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

// CreateOrUpdateGoogleToken creates or updates a Google OAuth token for a user
func (s *UserService) CreateOrUpdateGoogleToken(userID uint64, token *oauth2.Token) error {
	googleToken := &model.GoogleToken{
		ID:           utils.GenerateID(),
		UserID:       userID,
		RefreshToken: token.RefreshToken,
		AccessToken:  token.AccessToken,
		ExpiresAt:    token.Expiry,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return s.userRepo.CreateOrUpdateGoogleToken(googleToken)
}

// GetGoogleTokenByUserID retrieves a Google OAuth token by user ID
func (s *UserService) GetGoogleTokenByUserID(userID uint64) (*model.GoogleToken, error) {
	return s.userRepo.FindGoogleTokenByUserID(userID)
}

// DeleteGoogleTokenByUserID deletes a Google OAuth token by user ID
func (s *UserService) DeleteGoogleTokenByUserID(userID uint64) error {
	return s.userRepo.DeleteGoogleTokenByUserID(userID)
}
