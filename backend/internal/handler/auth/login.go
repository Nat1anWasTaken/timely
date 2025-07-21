package auth

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/NathanWasTaken/timely/backend/internal/model"
	"github.com/NathanWasTaken/timely/backend/internal/service"
	"github.com/NathanWasTaken/timely/backend/pkg/utils"
)

type LoginHandler struct {
	userService *service.UserService
	logger      *zap.Logger
}

func NewLoginHandler(userService *service.UserService) *LoginHandler {
	return &LoginHandler{
		userService: userService,
		logger:      zap.L(),
	}
}

// Login handles user authentication with email and password
// @Summary User Login
// @Description Authenticate user with email and password, returns JWT token on success
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "Login credentials"
// @Success 200 {object} model.AuthResponse "Login successful"
// @Failure 400 {object} model.ErrorResponse "Bad request - Invalid request body or missing fields"
// @Failure 401 {object} model.ErrorResponse "Unauthorized - Invalid credentials"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/auth/login [post]
func (h *LoginHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var loginReq model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		h.logger.Error("Failed to decode login request", zap.Error(err))
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
	if loginReq.Email == "" || loginReq.Password == "" {
		response := model.ErrorResponse{
			Success: false,
			Message: "Email and password are required",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Authenticate user
	user, err := h.userService.AuthenticateUser(&loginReq)
	if err != nil {
		h.logger.Error("Authentication failed", zap.Error(err), zap.String("email", loginReq.Email))
		response := model.ErrorResponse{
			Success: false,
			Message: "Authentication failed",
			Error:   err.Error(),
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		h.logger.Error("Failed to generate JWT token", zap.Error(err))
		response := model.ErrorResponse{
			Success: false,
			Message: "Failed to generate authentication token",
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
		Message: "Login successful",
		Token:   token,
		User:    user,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("User logged in successfully",
		zap.String("email", user.Email),
		zap.String("username", user.Username))
}

// Login function for backward compatibility with current router
func Login(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder that will be replaced when we update the router
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"message": "Please use the new login handler",
		"note":    "This endpoint needs to be updated in the router configuration",
	}
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(response)
}
