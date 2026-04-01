package build

import (
	"crypto/sha256"
	"encoding/hex"
)

// Source represents source code
type Source struct {
	Name string
	Zip  []byte
}

// SHA256 returns hex-encoded SHA256 hash of zip contents
func (s *Source) SHA256() string {
	if len(s.Zip) == 0 {
		return ""
	}
	hash := sha256.Sum256(s.Zip)
	return hex.EncodeToString(hash[:])
}
