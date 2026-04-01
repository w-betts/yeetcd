package engine

import (
	"fmt"
	"strings"
)

// ImageBase represents base image types
type ImageBase int

const (
	JAVA ImageBase = iota
)

// BaseImage returns the base image name
func (i ImageBase) BaseImage() string {
	switch i {
	case JAVA:
		return "maven:3.9.9-eclipse-temurin-17"
	default:
		return ""
	}
}

// EntryPoint returns the entry point command for the image.
// For JAVA, it builds a classpath from the artifact parent directory and artifact names.
func (i ImageBase) EntryPoint(artifactParentDirectoryPath string, artifactDefinitionNames []string) []string {
	switch i {
	case JAVA:
		// Build classpath like: /artifacts/classes:/artifacts/classes/*:/artifacts/dependencies:/artifacts/dependencies/*
		var classPathParts []string
		for _, name := range artifactDefinitionNames {
			path := fmt.Sprintf("%s/%s", artifactParentDirectoryPath, name)
			classPathParts = append(classPathParts, path, fmt.Sprintf("%s/*", path))
		}
		classPath := strings.Join(classPathParts, ":")
		return []string{"java", "-cp", classPath}
	default:
		return nil
	}
}
