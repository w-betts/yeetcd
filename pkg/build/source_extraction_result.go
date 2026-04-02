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
}

// Close cleans up the temporary directory
func (s *SourceExtractionResult) Close() error {
	if s.Directory != "" {
		return os.RemoveAll(s.Directory)
	}
	return nil
}

// Ensure SourceExtractionResult implements io.Closer
var _ io.Closer = (*SourceExtractionResult)(nil)
