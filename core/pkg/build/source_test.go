package build

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSource_SHA256_EmptyZip tests SHA256 with empty/nil zip
// Given: A Source with nil or empty Zip field
// When: SHA256 is called
// Then: It should return an empty string
func TestSource_SHA256_EmptyZip(t *testing.T) {
	s := &Source{Name: "test", Zip: nil}
	got := s.SHA256()
	assert.Empty(t, got, "SHA256() for empty zip should return empty string")
}

// TestSource_SHA256_ConsistentHash tests that SHA256 returns consistent results
// Given: A Source with specific zip content
// When: SHA256 is called multiple times
// Then: It should return the same hash each time
func TestSource_SHA256_ConsistentHash(t *testing.T) {
	s := &Source{Name: "test", Zip: []byte{0x01, 0x02, 0x03}}

	hash1 := s.SHA256()
	hash2 := s.SHA256()

	assert.Equal(t, hash1, hash2, "SHA256() should be consistent")
	// Verify it's a hex string of correct length (64 chars for SHA256)
	assert.Len(t, hash1, 64, "SHA256() should return 64 character hex string")
}

// TestSource_SHA256_DifferentContentDifferentHash tests hash uniqueness
// Given: Two Source objects with different zip content
// When: SHA256 is called on each
// Then: It should return different hashes
func TestSource_SHA256_DifferentContentDifferentHash(t *testing.T) {
	s1 := &Source{Name: "test", Zip: []byte{0x01, 0x02, 0x03}}
	s2 := &Source{Name: "test", Zip: []byte{0x01, 0x02, 0x04}}

	hash1 := s1.SHA256()
	hash2 := s2.SHA256()

	assert.NotEqual(t, hash1, hash2, "SHA256() should be different for different content")
}

// TestSource_SHA256_KnownValue tests SHA256 against known value
// Given: A Source with known content ("hello")
// When: SHA256 is called
// Then: It should return the known SHA256 hash
func TestSource_SHA256_KnownValue(t *testing.T) {
	// Known SHA256 of "hello"
	s := &Source{Name: "test", Zip: []byte("hello")}
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	got := s.SHA256()

	assert.Equal(t, want, got, "SHA256() should match known value")
}
