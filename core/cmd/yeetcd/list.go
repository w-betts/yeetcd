package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/yeetcd/yeetcd/internal/core/pipeline"
	"github.com/yeetcd/yeetcd/pkg/build"
	"github.com/yeetcd/yeetcd/pkg/engine/docker"
)

var (
	listSourcePath string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available pipelines from source code",
	Long: `List all available pipelines from a source zip file.

This command:
  1. Reads the source zip file
  2. Builds the source in a container
  3. Generates pipeline definitions
  4. Lists all available pipelines and their parameters

Example:
  yeetcd list --source ./project.zip`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listPipelines()
	},
}

func init() {
	listCmd.Flags().StringVar(&listSourcePath, "source", "", "Path to source zip file (required)")
	listCmd.MarkFlagRequired("source")
}

// listPipelines lists all available pipelines from the source
func listPipelines() error {
	ctx := context.Background()

	// Step 1: Read zip file
	slog.Info("reading source zip", "path", listSourcePath)
	zipData, err := os.ReadFile(listSourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Step 2: Create Source
	source := build.Source{
		Name: listSourcePath,
		Zip:  zipData,
	}

	// Step 3: Create Docker execution engine
	slog.Info("initializing Docker execution engine")
	executionEngine, err := docker.NewDockerExecutionEngine()
	if err != nil {
		return fmt.Errorf("failed to create Docker execution engine: %w", err)
	}

	// Step 4: Create build service
	buildService := build.NewDockerBuildService(executionEngine)

	// Step 5: Create source extractor
	sourceExtractor := build.NewSourceExtractor()

	// Step 6: Create PipelineController
	controller := pipeline.NewPipelineController(buildService, sourceExtractor, executionEngine)

	// Step 7: Assemble pipelines from source
	slog.Info("assembling pipelines from source")
	pipelines, err := controller.Assemble(ctx, source)
	if err != nil {
		return fmt.Errorf("failed to assemble pipelines: %w", err)
	}

	// Step 8: Print pipeline names and parameters
	if len(pipelines) == 0 {
		fmt.Println("No pipelines found in source")
		return nil
	}

	fmt.Printf("\nFound %d pipeline(s):\n\n", len(pipelines))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tPARAMETERS")
	fmt.Fprintln(w, "----\t----------")

	for _, p := range pipelines {
		paramNames := make([]string, 0, len(p.Parameters))
		for name := range p.Parameters {
			paramNames = append(paramNames, name)
		}
		if len(paramNames) == 0 {
			paramNames = []string{"(none)"}
		}
		fmt.Fprintf(w, "%s\t%s\n", p.Name, joinParams(paramNames))
	}

	w.Flush()
	fmt.Println()

	return nil
}

// joinParams joins parameter names with commas
func joinParams(params []string) string {
	if len(params) == 0 {
		return ""
	}
	result := params[0]
	for i := 1; i < len(params); i++ {
		result += ", " + params[i]
	}
	return result
}
