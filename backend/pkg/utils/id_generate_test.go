package utils

import (
	"fmt"
	"testing"
)

func TestSnowflakeIDGeneration(t *testing.T) {
	// Initialize with node ID 1
	InitSnowflake(1)

	// Test basic ID generation
	id1 := GenerateID()
	id2 := GenerateID()

	// IDs should be different
	if id1 == id2 {
		t.Error("Generated IDs should be different")
	}

	// IDs should be greater than 0
	if id1 == 0 || id2 == 0 {
		t.Error("Generated IDs should be greater than 0")
	}

	fmt.Printf("Generated ID 1: %d\n", id1)
	fmt.Printf("Generated ID 2: %d\n", id2)
}

func TestSnowflakeStringGeneration(t *testing.T) {
	InitSnowflake(1)

	idStr := GenerateIDString()
	if idStr == "" {
		t.Error("Generated string ID should not be empty")
	}

	fmt.Printf("Generated String ID: %s\n", idStr)
}
