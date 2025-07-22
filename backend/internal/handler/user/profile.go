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
// @Router /api/user/profile [get]
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by JWT middleware)
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User not found in context")
		sendErrorResponse(w, "Authentication required", "authentication_required", http.StatusUnauthorized)
		return
	}

	h.logger.Info("Fetching user profile", zap.Uint64("user_id", user.ID))

	// Get full user data from service
	fullUser, err := h.userService.GetUserByID(user.ID)
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