package build

import (
	"archive/zip"
	"bytes"
	"io"
	"io/fs"
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

// Extract extracts a source zip or uses a source directory and parses yeetcd.yaml files
func (s *SourceExtractor) Extract(source Source) (*SourceExtractionResult, error) {
	// Handle directory source
	if source.Directory != "" {
		return s.extractFromDirectory(source)
	}

	// Handle zip source
	return s.extractFromZip(source)
}

// extractFromDirectory handles extraction from a directory source
func (s *SourceExtractor) extractFromDirectory(source Source) (*SourceExtractionResult, error) {
	// Verify directory exists
	info, err := os.Stat(source.Directory)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fs.ErrInvalid
	}

	// Map to store yeetcd.yaml configs keyed by their parent directory path
	yeetcdDefinitions := make(map[string]config.YeetcdConfig)

	// Walk the directory and find all yeetcd.yaml files
	err = filepath.WalkDir(source.Directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check if this is a yeetcd.yaml file
		if filepath.Base(path) == "yeetcd.yaml" {
			// Read the file
			contents, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Parse the config
			yeetcdConfig, err := config.LoadFromBytes(contents)
			if err != nil {
				return err
			}

			// Get the relative path from the source directory
			relPath, err := filepath.Rel(source.Directory, filepath.Dir(path))
			if err != nil {
				return err
			}

			// Use "." for root directory
			if relPath == "." {
				relPath = ""
			}

			yeetcdDefinitions[relPath] = *yeetcdConfig
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &SourceExtractionResult{
		Source:            source,
		Directory:         source.Directory,
		YeetcdDefinitions: yeetcdDefinitions,
		isTempDir:         false, // User-provided directory, don't clean up
	}, nil
}

// extractFromZip handles extraction from a zip source
func (s *SourceExtractor) extractFromZip(source Source) (*SourceExtractionResult, error) {
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
		isTempDir:         true, // We created this temp directory, clean it up
	}, nil
}
