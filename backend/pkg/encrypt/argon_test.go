package encrypt

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt(16)
	if err != nil {
		t.Fatalf("Failed to generate salt: %v", err)
	}

	if len(salt1) != 16 {
		t.Errorf("Expected salt length 16, got %d", len(salt1))
	}

	salt2, err := GenerateSalt(16)
	if err != nil {
		t.Fatalf("Failed to generate second salt: %v", err)
	}

	// Salts should be different
	if string(salt1) == string(salt2) {
		t.Error("Generated salts should be different")
	}
}

func TestHashPassword(t *testing.T) {
	password := "testPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Check hash format
	if !strings.HasPrefix(hash, "$argon2id$v=19$") {
		t.Errorf("Hash should start with $argon2id$v=19$, got: %s", hash[:20])
	}

	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		t.Errorf("Hash should have 6 parts separated by $, got %d parts", len(parts))
	}

	fmt.Printf("Generated hash: %s\n", hash)
}

func TestVerifyPassword(t *testing.T) {
	password := "testPassword123!"
	wrongPassword := "wrongPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Test correct password
	if !VerifyPassword(password, hash) {
		t.Error("Password verification should succeed for correct password")
	}

	// Test wrong password
	if VerifyPassword(wrongPassword, hash) {
		t.Error("Password verification should fail for wrong password")
	}

	// Test empty password
	if VerifyPassword("", hash) {
		t.Error("Password verification should fail for empty password")
	}
}

func TestVerifyPasswordWithInvalidHash(t *testing.T) {
	password := "testPassword123!"

	// Test various invalid hash formats
	invalidHashes := []string{
		"",
		"invalid",
		"$argon2$v=19$m=262144,t=10,p=4$salt$hash",   // wrong algorithm
		"$argon2id$v=18$m=262144,t=10,p=4$salt$hash", // wrong version
		"$argon2id$v=19$m=262144$salt$hash",          // missing parameters
		"$argon2id$v=19$m=262144,t=10,p=4$salt",      // missing hash
	}

	for _, invalidHash := range invalidHashes {
		if VerifyPassword(password, invalidHash) {
			t.Errorf("Password verification should fail for invalid hash: %s", invalidHash)
		}
	}
}

func TestHashPasswordConsistency(t *testing.T) {
	password := "testPassword123!"

	// Generate multiple hashes for the same password
	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Hashes should be different (due to different salts)
	if hash1 == hash2 {
		t.Error("Hashes should be different due to different salts")
	}

	// But both should verify the same password
	if !VerifyPassword(password, hash1) {
		t.Error("First hash should verify the password")
	}

	if !VerifyPassword(password, hash2) {
		t.Error("Second hash should verify the password")
	}
}

func TestCustomConfig(t *testing.T) {
	password := "testPassword123!"

	// Create a custom config with lower parameters for faster testing
	config := &Config{
		Time:       2,
		Memory:     64 * 1024, // 64 MB
		Threads:    2,
		KeyLength:  32,
		SaltLength: 16,
	}

	hash, err := HashPasswordWithConfig(password, config)
	if err != nil {
		t.Fatalf("Failed to hash password with custom config: %v", err)
	}

	if !VerifyPassword(password, hash) {
		t.Error("Password verification should succeed with custom config")
	}

	// Check that the hash contains the correct parameters
	if !strings.Contains(hash, "m=65536,t=2,p=2") {
		t.Errorf("Hash should contain custom parameters, got: %s", hash)
	}
}

func TestParseHash(t *testing.T) {
	// Create a known hash
	password := "testPassword123!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Parse the hash
	config, salt, hashBytes, err := parseHash(hash)
	if err != nil {
		t.Fatalf("Failed to parse hash: %v", err)
	}

	// Verify the parsed components
	if config.KeyLength != keyLength {
		t.Errorf("Expected key length %d, got %d", keyLength, config.KeyLength)
	}

	if config.SaltLength != saltLength {
		t.Errorf("Expected salt length %d, got %d", saltLength, config.SaltLength)
	}

	if len(salt) != int(config.SaltLength) {
		t.Errorf("Salt length mismatch: expected %d, got %d", config.SaltLength, len(salt))
	}

	if len(hashBytes) != int(config.KeyLength) {
		t.Errorf("Hash length mismatch: expected %d, got %d", config.KeyLength, len(hashBytes))
	}
}

// TestHashTiming measures the actual time taken to hash a password
func TestHashTiming(t *testing.T) {
	password := "testPassword123!"

	fmt.Println("\nTiming test for Argon2 hashing:")

	// Test with default config
	start := time.Now()
	hash, err := HashPassword(password)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	fmt.Printf("Default config hash time: %v\n", duration)
	fmt.Printf("Hash: %s\n", hash)

	// Verify the hash works
	verifyStart := time.Now()
	if !VerifyPassword(password, hash) {
		t.Error("Password verification failed")
	}
	verifyDuration := time.Since(verifyStart)

	fmt.Printf("Verification time: %v\n", verifyDuration)

	// Note: Actual timing will vary based on hardware
	// The goal is ~5 seconds, but we'll be lenient in tests
	if duration < 1*time.Second {
		t.Logf("Warning: Hash time (%v) is less than 1 second. Consider increasing parameters for production.", duration)
	}

	if duration > 30*time.Second {
		t.Errorf("Hash time (%v) is too long (>30s). Consider decreasing parameters.", duration)
	}
}

// TestDifferentPasswordLengths tests hashing with various password lengths
func TestDifferentPasswordLengths(t *testing.T) {
	passwords := []string{
		"a",                         // Very short
		"password",                  // Short
		"averageLengthPassword123!", // Medium
		strings.Repeat("veryLongPasswordWithLotsOfCharacters", 10), // Very long
	}

	for _, password := range passwords {
		t.Run(fmt.Sprintf("len_%d", len(password)), func(t *testing.T) {
			hash, err := HashPassword(password)
			if err != nil {
				t.Fatalf("Failed to hash password of length %d: %v", len(password), err)
			}

			if !VerifyPassword(password, hash) {
				t.Errorf("Failed to verify password of length %d", len(password))
			}
		})
	}
}

// TestConcurrentHashing tests that hashing works correctly when called concurrently
func TestConcurrentHashing(t *testing.T) {
	password := "testPassword123!"
	numGoroutines := 5

	results := make(chan struct {
		hash string
		err  error
	}, numGoroutines)

	// Start multiple goroutines hashing the same password
	for i := 0; i < numGoroutines; i++ {
		go func() {
			hash, err := HashPassword(password)
			results <- struct {
				hash string
				err  error
			}{hash, err}
		}()
	}

	// Collect results
	hashes := make([]string, 0, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		result := <-results
		if result.err != nil {
			t.Errorf("Goroutine %d failed: %v", i, result.err)
			continue
		}
		hashes = append(hashes, result.hash)
	}

	// Verify all hashes work
	for i, hash := range hashes {
		if !VerifyPassword(password, hash) {
			t.Errorf("Hash %d failed verification", i)
		}
	}

	// All hashes should be different (due to different salts)
	for i := 0; i < len(hashes); i++ {
		for j := i + 1; j < len(hashes); j++ {
			if hashes[i] == hashes[j] {
				t.Errorf("Hashes %d and %d are identical (should be different)", i, j)
			}
		}
	}
}

// BenchmarkHashPassword benchmarks the hashing function
func BenchmarkHashPassword(b *testing.B) {
	password := "testPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := HashPassword(password)
		if err != nil {
			b.Fatalf("Failed to hash password: %v", err)
		}
	}
}

// BenchmarkVerifyPassword benchmarks the verification function
func BenchmarkVerifyPassword(b *testing.B) {
	password := "testPassword123!"
	hash, err := HashPassword(password)
	if err != nil {
		b.Fatalf("Failed to hash password: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !VerifyPassword(password, hash) {
			b.Fatal("Password verification failed")
		}
	}
}
