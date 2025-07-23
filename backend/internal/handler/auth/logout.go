package auth

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/NathanWasTaken/timely/backend/internal/model"
	"github.com/NathanWasTaken/timely/backend/pkg/utils"
)

type LogoutHandler struct {
	logger *zap.Logger
}

func NewLogoutHandler() *LogoutHandler {
	return &LogoutHandler{
		logger: zap.L(),
	}
}

// Logout handles user logout by clearing JWT cookie
// @Summary User Logout
// @Description Clear user session by removing JWT cookie
// @Tags Authentication
// @Accept json
// @Produce json
// @Success 200 {object} model.AuthResponse "Logout successful"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /api/auth/logout [post]
// @Security BearerAuth
func (h *LogoutHandler) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Clear JWT cookie
	clearCookie := utils.ClearJWTCookie()
	http.SetCookie(w, clearCookie)

	// Return success response
	response := model.AuthResponse{
		Success: true,
		Message: "Logout successful",
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("User logged out successfully")
}