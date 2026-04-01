package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestWorkContext_MergeInto correctly merges two contexts with source overriding destination
// GIVEN: Two WorkContext maps with overlapping keys
// WHEN: MergeInto() is called
// THEN: Result contains all keys from both maps, with source context values overriding destination for duplicate keys
func TestWorkContext_MergeInto(t *testing.T) {
	// Create destination context
	dest := WorkContext{
		"key1": "dest1",
		"key2": "dest2",
		"key3": "dest3",
	}

	// Create source context with overlapping key
	source := WorkContext{
		"key2": "source2",
		"key4": "source4",
	}

	// Merge source into destination
	result := source.MergeInto(dest)

	// Assert
	assert.Equal(t, "dest1", result["key1"])
	assert.Equal(t, "source2", result["key2"]) // Source overrides
	assert.Equal(t, "dest3", result["key3"])
	assert.Equal(t, "source4", result["key4"])
}

// TestWorkContext_FromMap creates WorkContext from a map
func TestWorkContext_FromMap(t *testing.T) {
	m := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	wc := WorkContextFromMap(m)

	assert.Equal(t, "value1", wc["key1"])
	assert.Equal(t, "value2", wc["key2"])
}

// TestWorkContext_Empty creates an empty WorkContext
func TestWorkContext_Empty(t *testing.T) {
	wc := EmptyWorkContext()

	assert.NotNil(t, wc)
	assert.Empty(t, wc)
}
