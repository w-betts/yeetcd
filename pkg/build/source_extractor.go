package build

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/yeetcd/yeetcd/pkg/config"
)

// SourceExtractor extracts source zips and parses yeetcd.yaml files
type SourceExtractor struct{}

// NewSourceExtractor creates a new source extractor
func NewSourceExtractor() *SourceExtractor {
	return &SourceExtractor{}
}

// Extract extracts a source zip and parses yeetcd.yaml files
func (s *SourceExtractor) Extract(source Source) (*SourceExtractionResult, error) {
	// Create temp directory for extraction
	destDir, err := os.MkdirTemp("", "yeetcd-extraction-")
	if err != nil {
		return nil, err
	}

	// Open the zip reader
	reader, err := zip.NewReader(bytes.NewReader(source.Zip), int64(len(source.Zip)))
	if err != nil {
		os.RemoveAll(destDir)
		return nil, err
	}

	// Map to store yeetcd.yaml configs keyed by their parent directory path
	yeetcdDefinitions := make(map[string]config.YeetcdConfig)

	// Extract all files and parse yeetcd.yaml files
	for _, file := range reader.File {
		// Skip directories
		if file.FileInfo().IsDir() {
			continue
		}

		// Create the file path
		filePath := filepath.Join(destDir, file.Name)

		// Create parent directories if needed
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			os.RemoveAll(destDir)
			return nil, err
		}

		// Open the file from the zip
		rc, err := file.Open()
		if err != nil {
			os.RemoveAll(destDir)
			return nil, err
		}

		// Read file contents
		contents, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			os.RemoveAll(destDir)
			return nil, err
		}

		// Write the file to disk
		if err := os.WriteFile(filePath, contents, 0644); err != nil {
			os.RemoveAll(destDir)
			return nil, err
		}

		// If this is a yeetcd.yaml file, parse it
		if filepath.Base(file.Name) == "yeetcd.yaml" {
			yeetcdConfig, err := config.LoadFromBytes(contents)
			if err != nil {
				os.RemoveAll(destDir)
				return nil, err
			}

			// Key is the parent directory path (relative to extraction root)
			parentDir := filepath.Dir(file.Name)
			yeetcdDefinitions[parentDir] = *yeetcdConfig
		}
	}

	return &SourceExtractionResult{
		Source:            source,
		Directory:         destDir,
		YeetcdDefinitions: yeetcdDefinitions,
	}, nil
}
