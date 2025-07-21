package model

// LoginRequest represents the login request payload
// @Description Login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com" binding:"required"`     // User's email address
	Password string `json:"password" validate:"required,min=6" example:"password123" binding:"required,min=6"` // User's password (minimum 6 characters)
}

// RegisterRequest represents the registration request payload
// @Description Registration request payload
type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email" example:"user@example.com" binding:"required"`                      // User's email address
	Username    string `json:"username" validate:"required,min=3,max=50" example:"johndoe" binding:"required,min=3,max=50"`        // Username (3-50 characters)
	DisplayName string `json:"display_name" validate:"required,min=1,max=100" example:"John Doe" binding:"required,min=1,max=100"` // User's display name (1-100 characters)
	Password    string `json:"password" validate:"required,min=6" example:"password123" binding:"required,min=6"`                  // Password (minimum 6 characters)
}

// AuthResponse represents the authentication response
// @Description Successful authentication response
type AuthResponse struct {
	Success bool   `json:"success" example:"true"`                                            // Indicates if the operation was successful
	Message string `json:"message" example:"Login successful"`                                // Response message
	Token   string `json:"token,omitempty" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // JWT authentication token
	User    *User  `json:"user,omitempty"`                                                    // User information
}

// ErrorResponse represents an error response
// @Description Error response structure
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`                             // Always false for error responses
	Message string `json:"message" example:"Authentication failed"`             // Error message
	Error   string `json:"error,omitempty" example:"invalid email or password"` // Detailed error information
}
