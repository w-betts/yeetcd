package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// This is a simple generator that finds pipeline definitions in Go source files.
// It looks for functions that return sdk.Pipeline (by naming convention *Pipeline)
// and generates pipeline protobuf.

func main() {
	fmt.Println("Yeetcd Go Pipeline Generator")
	fmt.Println("This generator finds pipeline definitions in your code.")
	fmt.Println("")
	fmt.Println("To use this generator, add the following to any .go file in your project:")
	fmt.Println("")
	fmt.Println("  //go:generate go run github.com/yeetcd/yeetcd/sdk/generator/cmd/generate")
	fmt.Println("")
	fmt.Println("Then run: go generate")
	fmt.Println("")
	fmt.Println("The generator will find functions that return sdk.Pipeline and generate pipelines.pb")

	// For now, just print guidance. The actual implementation would:
	// 1. Parse Go source files in current directory
	// 2. Find functions returning Pipeline (naming pattern: *Pipeline)
	// 3. Generate protobuf output

	// Check if we're being run as part of go generate
	if _, err := os.Stat("pipelines.pb"); err == nil {
		// pipelines.pb already exists - user might want to regenerate
		fmt.Println("\nNote: pipelines.pb already exists. Remove it and run go generate to regenerate.")
	}

	// Try to find Go files with pipeline definitions
	files, err := filepath.Glob("*.go")
	if err != nil || len(files) == 0 {
		fmt.Println("\nNo .go files found in current directory.")
		os.Exit(0)
	}

	// Look for pipeline function definitions
	pipelineFuncs := findPipelineFunctions(files)
	if len(pipelineFuncs) > 0 {
		fmt.Printf("\nFound %d potential pipeline function(s):\n", len(pipelineFuncs))
		for _, f := range pipelineFuncs {
			fmt.Printf("  - %s\n", f)
		}
		fmt.Println("\nTo generate pipelines.pb, implement these functions to return sdk.Pipeline")
	}
}

func findPipelineFunctions(files []string) []string {
	var funcs []string

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		// Simple pattern matching for functions returning Pipeline
		// Look for: func <name>Pipeline(...) sdk.Pipeline
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			// Skip comments
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") {
				continue
			}

			// Look for function definitions that return Pipeline
			if strings.Contains(line, "func ") && strings.Contains(line, "Pipeline") && strings.Contains(line, "sdk.Pipeline") {
				// Extract function name
				funcStart := strings.Index(line, "func ")
				rest := line[funcStart+5:]
				spaceIdx := strings.Index(rest, "(")
				if spaceIdx > 0 {
					funcName := strings.TrimSpace(rest[:spaceIdx])
					// Skip methods (they have a receiver)
					if !strings.Contains(funcName, "(") {
						funcs = append(funcs, file+":"+funcName)
					}
				}
			}
		}
	}

	return funcs
}
