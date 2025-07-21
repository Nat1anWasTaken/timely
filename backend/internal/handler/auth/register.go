package auth

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/NathanWasTaken/timely/backend/internal/model"
	"github.com/NathanWasTaken/timely/backend/internal/service"
	"github.com/NathanWasTaken/timely/backend/pkg/utils"
)

type RegisterHandler struct {
	userService *service.UserService
	logger      *zap.Logger
}

func NewRegisterHandler(userService *service.UserService) *RegisterHandler {
	return &RegisterHandler{
		userService: userService,
		logger:      zap.L(),
	}
}

// Register handles user registration with email and password
// @Summary User Registration
// @Description Register a new user account with email, username, display name, and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "Registration details"
// @Success 201 {object} model.AuthResponse "Registration successful"
// @Failure 400 {object} model.ErrorResponse "Bad request - Invalid request body, missing fields, or validation errors"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/auth/register [post]
func (h *RegisterHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var registerReq model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&registerReq); err != nil {
		h.logger.Error("Failed to decode register request", zap.Error(err))
		response := model.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate request
	if registerReq.Email == "" || registerReq.Username == "" || registerReq.DisplayName == "" || registerReq.Password == "" {
		response := model.ErrorResponse{
			Success: false,
			Message: "Email, username, display name, and password are required",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Additional validation
	if len(registerReq.Password) < 6 {
		response := model.ErrorResponse{
			Success: false,
			Message: "Password must be at least 6 characters long",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if len(registerReq.Username) < 3 {
		response := model.ErrorResponse{
			Success: false,
			Message: "Username must be at least 3 characters long",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create user
	user, err := h.userService.CreateUser(&registerReq)
	if err != nil {
		h.logger.Error("Registration failed", zap.Error(err), zap.String("email", registerReq.Email))
		response := model.ErrorResponse{
			Success: false,
			Message: "Registration failed",
			Error:   err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		h.logger.Error("Failed to generate JWT token", zap.Error(err))
		response := model.ErrorResponse{
			Success: false,
			Message: "Registration successful but failed to generate authentication token",
			Error:   err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Set JWT as HttpOnly cookie
	jwtCookie := utils.CreateJWTCookie(token)
	http.SetCookie(w, jwtCookie)

	// Clear password from response
	user.Password = nil

	// Return success response
	response := model.AuthResponse{
		Success: true,
		Message: "Registration successful",
		Token:   token,
		User:    user,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("User registered successfully",
		zap.String("email", user.Email),
		zap.String("username", user.Username))
}

// Register function for backward compatibility with current router
func Register(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder that will be replaced when we update the router
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"message": "Please use the new register handler",
		"note":    "This endpoint needs to be updated in the router configuration",
	}
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(response)
}
