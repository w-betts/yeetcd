package build

import (
	"crypto/sha256"
	"encoding/hex"
)

// Source represents source code, which can be either a zip archive or a directory
type Source struct {
	Name      string
	Zip       []byte // Zip data (if source is a zip archive)
	Directory string // Directory path (if source is a directory)
	SkipBuild bool   // Skip the build step (use pre-compiled classes)
}

// SHA256 returns hex-encoded SHA256 hash of zip contents
// Returns empty string if source is a directory (directories don't have a single hash)
func (s *Source) SHA256() string {
	if len(s.Zip) == 0 {
		return ""
	}
	hash := sha256.Sum256(s.Zip)
	return hex.EncodeToString(hash[:])
}
