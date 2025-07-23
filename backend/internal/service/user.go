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
		s.logger.Info("Found existing user by Google ID", zap.String("username", user.Username))

		// Update OAuth token if provided
		if token != nil {
			if err := s.UpdateGoogleAccountTokens(user.ID, token); err != nil {
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
		// Link Google account to existing user
		googleAccount := &model.Account{
			ID:         utils.GenerateID(),
			UserID:     user.ID,
			Provider:   "google",
			ProviderID: googleUser.ID,
			Email:      &googleUser.Email,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if err := s.userRepo.CreateAccount(googleAccount); err != nil {
			return nil, err
		}

		// Update user picture if not set
		if user.Picture == nil && googleUser.Picture != "" {
			user.Picture = &googleUser.Picture
			user.UpdatedAt = time.Now()
			if err := s.userRepo.Update(user); err != nil {
				return nil, err
			}
		}

		// Store OAuth token if provided
		if token != nil {
			if err := s.UpdateGoogleAccountTokens(user.ID, token); err != nil {
				s.logger.Error("Failed to store Google token", zap.Error(err))
				// Don't fail the login if token storage fails, just log it
			}
		}

		s.logger.Info("Linked Google account to existing user", zap.String("username", user.Username))
		return user, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new user
	user := &model.User{
		ID:          utils.GenerateID(),
		Username:    s.generateUniqueUsername(googleUser.Name),
		DisplayName: googleUser.Name,
		Picture:     &googleUser.Picture,
		Password:    nil, // OAuth users don't have passwords
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Create Google account
	googleAccount := &model.Account{
		ID:         utils.GenerateID(),
		UserID:     user.ID,
		Provider:   "google",
		ProviderID: googleUser.ID,
		Email:      &googleUser.Email,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.userRepo.CreateAccount(googleAccount); err != nil {
		return nil, err
	}

	// Store OAuth token if provided
	if token != nil {
		if err := s.UpdateGoogleAccountTokens(user.ID, token); err != nil {
			s.logger.Error("Failed to store Google token", zap.Error(err))
			// Don't fail the login if token storage fails, just log it
		}
	}

	s.logger.Info("Created new user with Google account", zap.String("username", user.Username))
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

// GetUserWithAccountsByID retrieves a user by ID with associated accounts
func (s *UserService) GetUserWithAccountsByID(id uint64) (*model.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	accounts, err := s.userRepo.FindAccountsByUserID(id)
	if err != nil {
		return nil, err
	}

	user.Accounts = accounts
	return user, nil
}

// CreateUser creates a new user with email/password authentication
func (s *UserService) CreateUser(req *model.RegisterRequest) (*model.User, error) {
	// Check if user already exists
	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user with this email already exists")
	}

	exists, err = s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username already taken")
	}

	// Hash password
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &model.User{
		ID:          utils.GenerateID(),
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Password:    &hashedPassword,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Create email account
	emailAccount := &model.Account{
		ID:         utils.GenerateID(),
		UserID:     user.ID,
		Provider:   "email",
		ProviderID: req.Email,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.userRepo.CreateAccount(emailAccount); err != nil {
		return nil, err
	}

	s.logger.Info("Created new user", zap.String("username", user.Username))
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

	s.logger.Info("User authenticated successfully", zap.String("username", user.Username))
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

// UpdateGoogleAccountTokens updates the OAuth tokens for a Google account
func (s *UserService) UpdateGoogleAccountTokens(userID uint64, token *oauth2.Token) error {
	return s.userRepo.UpdateGoogleAccountTokens(userID, token.AccessToken, token.RefreshToken, &token.Expiry)
}

// GetGoogleAccountByUserID retrieves a Google account by user ID
func (s *UserService) GetGoogleAccountByUserID(userID uint64) (*model.Account, error) {
	return s.userRepo.FindGoogleAccountByUserID(userID)
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(user *model.User) error {
	return s.userRepo.Update(user)
}

// GetAccountsByUserID retrieves all accounts for a user
func (s *UserService) GetAccountsByUserID(userID uint64) ([]model.Account, error) {
	return s.userRepo.FindAccountsByUserID(userID)
}

// LinkGoogleAccount links a Google account to an existing user
func (s *UserService) LinkGoogleAccount(userID uint64, googleID string, email string) error {
	googleAccount := &model.Account{
		ID:         utils.GenerateID(),
		UserID:     userID,
		Provider:   "google",
		ProviderID: googleID,
		Email:      &email,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return s.userRepo.CreateAccount(googleAccount)
}

