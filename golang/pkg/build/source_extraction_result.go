package build

import (
	"io"
	"os"

	"github.com/yeetcd/yeetcd/pkg/config"
)

// SourceExtractionResult holds the result of extracting source code
type SourceExtractionResult struct {
	Source            Source
	Directory         string
	YeetcdDefinitions map[string]config.YeetcdConfig
	isTempDir         bool // Track if we created a temp directory (needs cleanup)
}

// Close cleans up the temporary directory if one was created
// For directory sources, this is a no-op since we don't own the directory
func (s *SourceExtractionResult) Close() error {
	// Only remove the directory if we created it (temp directory for zip extraction)
	if s.isTempDir && s.Directory != "" {
		return os.RemoveAll(s.Directory)
	}
	return nil
}

// Ensure SourceExtractionResult implements io.Closer
var _ io.Closer = (*SourceExtractionResult)(nil)
