package user

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/NathanWasTaken/timely/backend/internal/middleware"
	"github.com/NathanWasTaken/timely/backend/internal/model"
	"github.com/NathanWasTaken/timely/backend/internal/service"
)

type UserHandler struct {
	userService *service.UserService
	logger      *zap.Logger
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      zap.L(),
	}
}

// GetProfile retrieves the current user's profile information
// @Summary Get User Profile
// @Description Retrieves the authenticated user's profile information from JWT token
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.UserProfileResponse "User profile retrieved successfully"
// @Failure 401 {object} model.ErrorResponse "Unauthorized - Authentication required"
// @Failure 404 {object} model.ErrorResponse "Not Found - User not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/users/me [get]
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by JWT middleware)
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User not found in context")
		sendErrorResponse(w, "Authentication required", "authentication_required", http.StatusUnauthorized)
		return
	}

	h.logger.Info("Fetching user profile", zap.Uint64("user_id", user.ID))

	// Get full user data with accounts from service
	fullUser, err := h.userService.GetUserWithAccountsByID(user.ID)
	if err != nil {
		h.logger.Error("Failed to get user profile", zap.Error(err), zap.Uint64("user_id", user.ID))
		sendErrorResponse(w, "User not found", "user_not_found", http.StatusNotFound)
		return
	}

	// Remove sensitive information before sending response
	fullUser.Password = nil

	// Create success response
	response := model.UserProfileResponse{
		Success: true,
		Message: "User profile retrieved successfully",
		User:    fullUser,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Successfully retrieved user profile", zap.Uint64("user_id", user.ID))
}

// GetPublicProfile retrieves a user's public profile information by username
// @Summary Get Public User Profile
// @Description Retrieves public profile information for a specific user by username. No authentication required.
// @Tags User
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} model.PublicUserProfileResponse "Public user profile retrieved successfully"
// @Failure 404 {object} model.ErrorResponse "Not Found - User not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/users/{username} [get]
func (h *UserHandler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	// Get username from path parameter
	username := r.PathValue("username")
	if username == "" {
		h.logger.Error("Username not provided in path")
		sendErrorResponse(w, "Username is required", "missing_username", http.StatusBadRequest)
		return
	}

	h.logger.Info("Fetching public user profile", zap.String("username", username))

	// Get user by username from service
	user, err := h.userService.GetUserByUsername(username)
	if err != nil {
		h.logger.Error("Failed to get user by username", zap.Error(err), zap.String("username", username))
		sendErrorResponse(w, "User not found", "user_not_found", http.StatusNotFound)
		return
	}

	// Create public profile (exclude sensitive information)
	publicProfile := &model.PublicUserProfile{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Picture:     user.Picture,
		CreatedAt:   user.CreatedAt,
	}

	// Create success response
	response := model.PublicUserProfileResponse{
		Success: true,
		Message: "Public user profile retrieved successfully",
		User:    publicProfile,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Successfully retrieved public user profile", zap.String("username", username), zap.Uint64("user_id", user.ID))
}

// UpdateProfile updates the current user's profile information
// @Summary Update User Profile
// @Description Updates the authenticated user's profile information (username and/or display name)
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.UpdateUserProfileRequest true "Profile update request"
// @Success 200 {object} model.UpdateUserProfileResponse "User profile updated successfully"
// @Failure 400 {object} model.ErrorResponse "Bad Request - Invalid input or username already taken"
// @Failure 401 {object} model.ErrorResponse "Unauthorized - Authentication required"
// @Failure 404 {object} model.ErrorResponse "Not Found - User not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/users/me [patch]
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by JWT middleware)
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User not found in context")
		sendErrorResponse(w, "Authentication required", "authentication_required", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req model.UpdateUserProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		sendErrorResponse(w, "Invalid request body", "invalid_request", http.StatusBadRequest)
		return
	}

	h.logger.Info("Updating user profile", zap.Uint64("user_id", user.ID))

	// Get current user data
	currentUser, err := h.userService.GetUserByID(user.ID)
	if err != nil {
		h.logger.Error("Failed to get current user", zap.Error(err), zap.Uint64("user_id", user.ID))
		sendErrorResponse(w, "User not found", "user_not_found", http.StatusNotFound)
		return
	}

	// Track if any changes are made
	hasChanges := false

	// Update username if provided
	if req.Username != nil && *req.Username != currentUser.Username {
		// Validate new username
		if *req.Username == "" {
			sendErrorResponse(w, "Username cannot be empty", "invalid_username", http.StatusBadRequest)
			return
		}

		// Check if username already exists
		if exists, err := h.userService.GetUserByUsername(*req.Username); err == nil && exists.ID != user.ID {
			sendErrorResponse(w, "Username already taken", "username_taken", http.StatusBadRequest)
			return
		}

		currentUser.Username = *req.Username
		hasChanges = true
	}

	// Update display name if provided
	if req.DisplayName != nil && *req.DisplayName != currentUser.DisplayName {
		if *req.DisplayName == "" {
			sendErrorResponse(w, "Display name cannot be empty", "invalid_display_name", http.StatusBadRequest)
			return
		}

		currentUser.DisplayName = *req.DisplayName
		hasChanges = true
	}

	// If no changes, return current user
	if !hasChanges {
		response := model.UpdateUserProfileResponse{
			Success: true,
			Message: "No changes detected",
			User:    currentUser,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Update user in database
	if err := h.userService.UpdateUser(currentUser); err != nil {
		h.logger.Error("Failed to update user", zap.Error(err), zap.Uint64("user_id", user.ID))
		sendErrorResponse(w, "Failed to update profile", "update_failed", http.StatusInternalServerError)
		return
	}

	// Remove sensitive information before sending response
	currentUser.Password = nil

	// Create success response
	response := model.UpdateUserProfileResponse{
		Success: true,
		Message: "User profile updated successfully",
		User:    currentUser,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Successfully updated user profile", zap.Uint64("user_id", user.ID))
}

// sendErrorResponse sends a standardized error response
func sendErrorResponse(w http.ResponseWriter, message, errorType string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := model.ErrorResponse{
		Success: false,
		Message: message,
		Error:   errorType,
	}

	json.NewEncoder(w).Encode(response)
}
