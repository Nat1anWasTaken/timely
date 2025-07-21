package utils

import (
	"log"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	snowflakeNode *snowflake.Node
	once          sync.Once
)

// InitSnowflake initializes the snowflake node with the given node ID
func InitSnowflake(nodeID int64) {
	once.Do(func() {
		node, err := snowflake.NewNode(nodeID)
		if err != nil {
			log.Fatalf("Failed to create snowflake node: %v", err)
		}
		snowflakeNode = node
	})
}

// GenerateID generates a new snowflake ID
func GenerateID() uint64 {
	if snowflakeNode == nil {
		// Default to node 1 if not initialized
		InitSnowflake(1)
	}
	return uint64(snowflakeNode.Generate().Int64())
}

// GenerateIDString generates a new snowflake ID as string
func GenerateIDString() string {
	if snowflakeNode == nil {
		InitSnowflake(1)
	}
	return snowflakeNode.Generate().String()
}
