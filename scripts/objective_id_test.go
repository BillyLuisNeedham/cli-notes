package scripts

import (
	"testing"
)

func TestGenerateObjectiveID(t *testing.T) {
	// Test that ID is generated successfully
	id, err := GenerateObjectiveID()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test that ID has correct length
	if len(id) != 8 {
		t.Errorf("Expected ID length 8, got %d", len(id))
	}

	// Test that ID is valid hex
	if !ValidateObjectiveID(id) {
		t.Errorf("Generated ID %s failed validation", id)
	}
}

func TestGenerateObjectiveID_Uniqueness(t *testing.T) {
	// Generate multiple IDs and check for uniqueness
	ids := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		id, err := GenerateObjectiveID()
		if err != nil {
			t.Fatalf("Error generating ID at iteration %d: %v", i, err)
		}

		if ids[id] {
			t.Errorf("Duplicate ID found: %s", id)
		}
		ids[id] = true
	}

	if len(ids) != iterations {
		t.Errorf("Expected %d unique IDs, got %d", iterations, len(ids))
	}
}

func TestValidateObjectiveID(t *testing.T) {
	testCases := []struct {
		name     string
		id       string
		expected bool
	}{
		{
			name:     "Valid 8-character hex",
			id:       "7a8f9b2c",
			expected: true,
		},
		{
			name:     "Valid 8-character hex uppercase",
			id:       "7A8F9B2C",
			expected: true,
		},
		{
			name:     "Invalid - too short",
			id:       "7a8f9b",
			expected: false,
		},
		{
			name:     "Invalid - too long",
			id:       "7a8f9b2c1d",
			expected: false,
		},
		{
			name:     "Invalid - non-hex characters",
			id:       "7a8f9xyz",
			expected: false,
		},
		{
			name:     "Invalid - empty string",
			id:       "",
			expected: false,
		},
		{
			name:     "Invalid - special characters",
			id:       "7a8f-b2c",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidateObjectiveID(tc.id)
			if result != tc.expected {
				t.Errorf("ValidateObjectiveID(%q) = %v, expected %v", tc.id, result, tc.expected)
			}
		})
	}
}
