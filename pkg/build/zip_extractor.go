package build

import (
	"errors"
)

// FileHandler handles specific files during zip extraction
type FileHandler struct {
	ShouldHandle func(name string) bool
	Handle       func(parent string, contents []byte) error
}

// HandledFile represents a file that was handled
type HandledFile struct {
	Parent   string
	Contents []byte
}

// ZipExtractor extracts zip files
type ZipExtractor struct{}

// NewZipExtractor creates a new zip extractor
func NewZipExtractor() *ZipExtractor {
	return &ZipExtractor{}
}

// Extract extracts zip data to a destination directory
func (z *ZipExtractor) Extract(zipData []byte, destDir string, handlers ...FileHandler) error {
	return errors.New("not implemented")
}
