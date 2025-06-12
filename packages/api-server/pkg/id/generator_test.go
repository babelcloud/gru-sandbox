package id_test

import (
	"regexp"
	"testing"

	"github.com/babelcloud/gbox/packages/api-server/pkg/id"
	"github.com/stretchr/testify/assert"
)

func TestGenerateBoxID(t *testing.T) {
	// Test multiple IDs to ensure format and uniqueness
	generatedIDs := make(map[string]bool)
	iterations := 1000

	// UUID v4 pattern
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`

	for i := 0; i < iterations; i++ {
		boxID := id.GenerateBoxID()

		// Check format matches UUID v4 pattern
		matched, err := regexp.MatchString(pattern, boxID)
		assert.NoError(t, err)
		assert.True(t, matched, "Generated ID %s does not match UUID v4 pattern", boxID)

		// Check uniqueness
		_, exists := generatedIDs[boxID]
		assert.False(t, exists, "Generated duplicate ID: %s", boxID)
		generatedIDs[boxID] = true

		// Check length is correct (36 characters including hyphens)
		assert.Equal(t, 36, len(boxID), "UUID length should be 36 characters")

		// Check that the version bits are set correctly (version 4)
		assert.Equal(t, "4", string(boxID[14]), "UUID version should be 4")

		// Check that the variant bits are set correctly (variant 1)
		variantChar := boxID[19]
		assert.Contains(t, "89ab", string(variantChar), "UUID variant should be 1 (RFC 4122)")
	}
}
