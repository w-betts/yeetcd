package build

import (
	"errors"
	"io"
)

// SourceExtractionResult holds the result of extracting source code
type SourceExtractionResult struct {
	Source            Source
	Directory         string
	YeetcdDefinitions map[string]interface{} // Will be map[string]config.YeetcdConfig
}

// Close cleans up the temporary directory
func (s *SourceExtractionResult) Close() error {
	return errors.New("not implemented")
}

// Ensure SourceExtractionResult implements io.Closer
var _ io.Closer = (*SourceExtractionResult)(nil)
