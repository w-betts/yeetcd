package build

import (
	"errors"
)

// SourceExtractor extracts source zips and parses yeetcd.yaml files
type SourceExtractor struct{}

// NewSourceExtractor creates a new source extractor
func NewSourceExtractor() *SourceExtractor {
	return &SourceExtractor{}
}

// Extract extracts a source zip and parses yeetcd.yaml files
func (s *SourceExtractor) Extract(source Source) (*SourceExtractionResult, error) {
	return nil, errors.New("not implemented")
}
