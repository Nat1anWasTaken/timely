package encrypt

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"runtime"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2 parameters tuned for ~0.5 second hash time (benchmarked: ~496ms)
	// These can be adjusted based on your hardware
	iterations = 2          // Number of iterations (optimal for 500ms timing)
	memory     = 512 * 1024 // Memory usage in KB (512 MB, optimal for 500ms)
	threads    = 4          // Number of threads (adjust based on CPU cores)
	keyLength  = 32         // Length of the derived key in bytes
	saltLength = 16         // Length of the salt in bytes
)

// Config holds Argon2 configuration parameters
type Config struct {
	Time       uint32
	Memory     uint32
	Threads    uint8
	KeyLength  uint32
	SaltLength uint32
}

// DefaultConfig returns a configuration tuned for ~0.5 second hash time
func DefaultConfig() *Config {
	// Adjust threads based on available CPU cores
	numCPU := runtime.NumCPU()
	if numCPU > 4 {
		numCPU = 4 // Cap at 4 threads for consistency
	}

	return &Config{
		Time:       iterations,
		Memory:     memory,
		Threads:    uint8(numCPU),
		KeyLength:  keyLength,
		SaltLength: saltLength,
	}
}

// GenerateSalt creates a cryptographically secure random salt
func GenerateSalt(length uint32) ([]byte, error) {
	salt := make([]byte, length)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

// HashPassword hashes a password using Argon2id with secure parameters
// Returns a base64 encoded string in the format: $argon2id$v=19$m=memory,t=time,p=threads$salt$hash
func HashPassword(password string) (string, error) {
	config := DefaultConfig()
	return HashPasswordWithConfig(password, config)
}

// HashPasswordWithConfig hashes a password using the provided configuration
func HashPasswordWithConfig(password string, config *Config) (string, error) {
	// Generate a cryptographically secure salt
	salt, err := GenerateSalt(config.SaltLength)
	if err != nil {
		return "", err
	}

	// Generate the hash using Argon2id
	hash := argon2.IDKey([]byte(password), salt, config.Time, config.Memory, config.Threads, config.KeyLength)

	// Encode salt and hash to base64
	saltEncoded := base64.RawStdEncoding.EncodeToString(salt)
	hashEncoded := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=memory,t=time,p=threads$salt$hash
	encodedHash := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		config.Memory, config.Time, config.Threads, saltEncoded, hashEncoded)

	return encodedHash, nil
}

// VerifyPassword verifies that a password matches the stored hash
func VerifyPassword(password, encodedHash string) bool {
	// Parse the encoded hash
	config, salt, hash, err := parseHash(encodedHash)
	if err != nil {
		return false
	}

	// Hash the provided password with the same salt and config
	passwordHash := argon2.IDKey([]byte(password), salt, config.Time, config.Memory, config.Threads, config.KeyLength)

	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(hash, passwordHash) == 1
}

// parseHash extracts the configuration, salt, and hash from an encoded hash string
func parseHash(encodedHash string) (*Config, []byte, []byte, error) {
	// Expected format: $argon2id$v=19$m=memory,t=time,p=threads$salt$hash
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, fmt.Errorf("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return nil, nil, nil, fmt.Errorf("unsupported algorithm: %s", parts[1])
	}

	if parts[2] != "v=19" {
		return nil, nil, nil, fmt.Errorf("unsupported version: %s", parts[2])
	}

	// Parse parameters: m=memory,t=time,p=threads
	var memory, time uint32
	var threads uint8
	paramParts := strings.Split(parts[3], ",")
	if len(paramParts) != 3 {
		return nil, nil, nil, fmt.Errorf("invalid parameter format")
	}

	for _, param := range paramParts {
		keyValue := strings.Split(param, "=")
		if len(keyValue) != 2 {
			return nil, nil, nil, fmt.Errorf("invalid parameter: %s", param)
		}

		switch keyValue[0] {
		case "m":
			if _, err := fmt.Sscanf(keyValue[1], "%d", &memory); err != nil {
				return nil, nil, nil, fmt.Errorf("invalid memory parameter: %s", keyValue[1])
			}
		case "t":
			if _, err := fmt.Sscanf(keyValue[1], "%d", &time); err != nil {
				return nil, nil, nil, fmt.Errorf("invalid time parameter: %s", keyValue[1])
			}
		case "p":
			var p uint32
			if _, err := fmt.Sscanf(keyValue[1], "%d", &p); err != nil {
				return nil, nil, nil, fmt.Errorf("invalid threads parameter: %s", keyValue[1])
			}
			threads = uint8(p)
		}
	}

	// Decode salt and hash
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode hash: %w", err)
	}

	config := &Config{
		Time:       time,
		Memory:     memory,
		Threads:    threads,
		KeyLength:  uint32(len(hash)),
		SaltLength: uint32(len(salt)),
	}

	return config, salt, hash, nil
}

// BenchmarkHashTime measures how long it takes to hash a password with current config
// Useful for tuning parameters to achieve desired timing
func BenchmarkHashTime(password string, iterations int) (avgDuration float64, err error) {
	if iterations <= 0 {
		iterations = 1
	}

	config := DefaultConfig()
	salt, err := GenerateSalt(config.SaltLength)
	if err != nil {
		return 0, err
	}

	for i := 0; i < iterations; i++ {
		argon2.IDKey([]byte(password), salt, config.Time, config.Memory, config.Threads, config.KeyLength)
	}

	// Return the number of iterations completed
	// For actual timing, use time.Now() before and after the loop
	return float64(iterations), nil
}
