package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yeetcd/yeetcd/pkg/engine"
)

// CreateDockerfile generates a Dockerfile for building images.
// It creates a Dockerfile in the context directory and returns the path and a cleanup function.
// Note: We don't use ENTRYPOINT - instead we rely on the Cmd from JobDefinition.
// This avoids the issue where Docker combines ENTRYPOINT + CMD incorrectly.
func CreateDockerfile(ctx context.Context, def engine.BuildImageDefinition, contextDir string) (string, func(), error) {
	// Create Dockerfile path
	dockerfile := filepath.Join(contextDir, "Dockerfile")

	// Get base image
	baseImage := def.ImageBase.BaseImage()
	if baseImage == "" {
		return "", nil, fmt.Errorf("unknown image base: %d", def.ImageBase)
	}

	// For source images (custom work), don't set ENTRYPOINT - use CMD only
	// The full command (java -cp ... MainClass args) comes from JobDefinition.Cmd
	// For non-source images (containerised work), the image provides its own entrypoint
	//
	// Build Dockerfile content - just add artifacts, no ENTRYPOINT/CMD
	// The image base provides the runtime, and JobDefinition.Cmd provides the command
	content := fmt.Sprintf(`FROM %s
ADD / /artifacts
`, baseImage)

	// Write Dockerfile
	if err := os.WriteFile(dockerfile, []byte(content), 0644); err != nil {
		return "", nil, fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	// Return cleanup function
	cleanup := func() {
		if err := os.Remove(dockerfile); err != nil {
			// Log error but don't panic - cleanup is best effort
		}
	}

	return dockerfile, cleanup, nil
}

// formatJSONArray formats a string slice as a JSON array string
func formatJSONArray(items []string) string {
	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = fmt.Sprintf(`"%s"`, item)
	}
	return fmt.Sprintf("[%s]", strings.Join(quoted, ", "))
}
