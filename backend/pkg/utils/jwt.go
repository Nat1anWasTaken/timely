package utils

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-this-in-production" // Default secret for development
	}
	jwtSecret = []byte(secret)
}

// Claims represents the JWT claims structure
type Claims struct {
	UserID uint64 `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a new JWT token for the given user
func GenerateJWT(userID uint64, email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token expires in 24 hours

	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "timely-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// CreateJWTCookie creates an HTTP cookie with JWT token configured for the client domain
func CreateJWTCookie(token string) *http.Cookie {
	clientDomain := os.Getenv("FRONTEND_DOMAIN")

	// Extract domain from FRONTEND_DOMAIN URL (e.g., "http://localhost:3000" -> "localhost")
	var domain string
	if clientDomain != "" {
		// Remove protocol
		if strings.HasPrefix(clientDomain, "http://") {
			clientDomain = clientDomain[7:]
		} else if strings.HasPrefix(clientDomain, "https://") {
			clientDomain = clientDomain[8:]
		}

		// Extract domain part (remove port if present)
		parts := strings.Split(clientDomain, ":")
		domain = parts[0]

		// Don't set domain for localhost (browser requirement)
		if domain == "localhost" || domain == "127.0.0.1" {
			domain = ""
		}
	}

	// Determine if we should use Secure flag (true for HTTPS)
	secure := strings.HasPrefix(os.Getenv("FRONTEND_DOMAIN"), "https://")

	return &http.Cookie{
		Name:     "access_token",
		Value:    token,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour), // 1 day expiration
		Path:     "/",
		Domain:   domain,
	}
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Make sure the token is signed with the correct method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("token is not valid")
	}

	return claims, nil
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) string {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}
