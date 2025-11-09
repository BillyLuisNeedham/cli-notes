package scripts

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateObjectiveID creates a unique 8-character hash
func GenerateObjectiveID() (string, error) {
	bytes := make([]byte, 4) // 4 bytes = 8 hex chars
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ValidateObjectiveID checks if an ID is valid format
func ValidateObjectiveID(id string) bool {
	if len(id) != 8 {
		return false
	}
	_, err := hex.DecodeString(id)
	return err == nil
}
