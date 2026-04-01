package build

import (
	"testing"
)

func TestSource_SHA256_EmptyZip(t *testing.T) {
	s := &Source{Name: "test", Zip: nil}
	got := s.SHA256()
	if got != "" {
		t.Errorf("SHA256() for empty zip = %q, want empty string", got)
	}
}

func TestSource_SHA256_ConsistentHash(t *testing.T) {
	s := &Source{Name: "test", Zip: []byte{0x01, 0x02, 0x03}}

	hash1 := s.SHA256()
	hash2 := s.SHA256()

	if hash1 != hash2 {
		t.Errorf("SHA256() not consistent: %q != %q", hash1, hash2)
	}

	// Verify it's a hex string of correct length (64 chars for SHA256)
	if len(hash1) != 64 {
		t.Errorf("SHA256() length = %d, want 64", len(hash1))
	}
}

func TestSource_SHA256_DifferentContentDifferentHash(t *testing.T) {
	s1 := &Source{Name: "test", Zip: []byte{0x01, 0x02, 0x03}}
	s2 := &Source{Name: "test", Zip: []byte{0x01, 0x02, 0x04}}

	hash1 := s1.SHA256()
	hash2 := s2.SHA256()

	if hash1 == hash2 {
		t.Errorf("SHA256() should be different for different content, got same hash: %q", hash1)
	}
}

func TestSource_SHA256_KnownValue(t *testing.T) {
	// Known SHA256 of "hello"
	s := &Source{Name: "test", Zip: []byte("hello")}
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	got := s.SHA256()

	if got != want {
		t.Errorf("SHA256() = %q, want %q", got, want)
	}
}
