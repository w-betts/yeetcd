package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yeetcd/yeetcd/internal/cli"
	"github.com/yeetcd/yeetcd/internal/core/pipeline"
	"github.com/yeetcd/yeetcd/pkg/build"
	"github.com/yeetcd/yeetcd/pkg/engine"
	"github.com/yeetcd/yeetcd/pkg/engine/docker"
	"github.com/yeetcd/yeetcd/pkg/engine/mock"
)

var (
	runSourcePath           string
	runPipelineName         string
	runArguments            []string
	runMockExecutionAddress string
	runClasspath            string
	runSkipBuild            bool
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute a pipeline from source code",
	Long: `Execute a pipeline from source code.

This command:
  1. Reads the source (zip file or directory)
  2. Builds the source in a container
  3. Generates pipeline definitions
  4. Executes the specified pipeline with the Docker executor
  5. Reports progress and results

Example:
  yeetcd run --source ./project.zip --pipeline sample
  yeetcd run --source ./project-dir --pipeline sample
  yeetcd run --source ./project.zip --pipeline sample --argument KEY1=value1 --argument KEY2=value2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPipeline()
	},
}

func init() {
	runCmd.Flags().StringVar(&runSourcePath, "source", "", "Path to source zip file or directory (required)")
	runCmd.Flags().StringVar(&runPipelineName, "pipeline", "", "Name of the pipeline to execute (required)")
	runCmd.Flags().StringArrayVar(&runArguments, "argument", []string{}, "Pipeline arguments in KEY=VALUE format (can be repeated)")
	runCmd.Flags().StringVar(&runMockExecutionAddress, "mock-execution-engine-address", "", "Address of mock execution engine (e.g., localhost:50051)")
	runCmd.Flags().StringVar(&runClasspath, "classpath", "", "Classpath to use for running pipeline generator and custom work")
	runCmd.Flags().BoolVar(&runSkipBuild, "skip-build", false, "Skip the build step (use pre-compiled classes)")

	runCmd.MarkFlagRequired("source")
	runCmd.MarkFlagRequired("pipeline")
}

// runPipeline executes the pipeline with the given source and arguments
func runPipeline() error {
	ctx := context.Background()

	// Step 1: Create Source from path (auto-detect directory vs zip)
	source, err := createSourceFromPath(runSourcePath)
	if err != nil {
		return err
	}

	// Step 2: Create execution engine (Docker or Mock)
	var executionEngine engine.ExecutionEngine
	if runMockExecutionAddress != "" {
		slog.Info("initializing mock execution engine", "address", runMockExecutionAddress)
		executionEngine, err = mock.NewMockExecutionEngine(runMockExecutionAddress)
		if err != nil {
			return fmt.Errorf("failed to create mock execution engine: %w", err)
		}
	} else {
		slog.Info("initializing Docker execution engine")
		executionEngine, err = docker.NewDockerExecutionEngine()
		if err != nil {
			return fmt.Errorf("failed to create Docker execution engine: %w", err)
		}
	}

	// Step 3: Create build service
	buildService := build.NewDockerBuildService(executionEngine)

	// Step 4: Create source extractor
	sourceExtractor := build.NewSourceExtractor()

	// Step 5: Create PipelineController
	controller := pipeline.NewPipelineController(buildService, sourceExtractor, executionEngine)

	// Step 6: Assemble pipelines from source
	slog.Info("assembling pipelines from source")
	pipelines, err := controller.Assemble(ctx, source)
	if err != nil {
		return fmt.Errorf("failed to assemble pipelines: %w", err)
	}

	// Step 7: Find the requested pipeline
	var targetPipeline *pipeline.Pipeline
	for _, p := range pipelines {
		if p.Name == runPipelineName {
			targetPipeline = p
			break
		}
	}
	if targetPipeline == nil {
		availablePipelines := make([]string, len(pipelines))
		for i, p := range pipelines {
			availablePipelines[i] = p.Name
		}
		return fmt.Errorf("pipeline '%s' not found. Available pipelines: %s",
			runPipelineName, strings.Join(availablePipelines, ", "))
	}

	// Step 8: Parse and apply arguments
	if len(runArguments) > 0 {
		args := parseArguments(runArguments)
		var err error
		targetPipeline, err = targetPipeline.WithArguments(args)
		if err != nil {
			return fmt.Errorf("failed to apply arguments: %w", err)
		}
	}

	// Step 9: Create CLI output handler
	outputHandler := cli.NewOutputHandler()

	// Step 10: Execute the pipeline
	slog.Info("executing pipeline", "name", targetPipeline.Name)
	result, err := controller.Execute(ctx, targetPipeline, outputHandler)
	if err != nil {
		return fmt.Errorf("pipeline execution failed: %w", err)
	}

	// Step 11: Print result and exit with appropriate code
	status := result.PipelineStatus()
	slog.Info("pipeline completed", "status", status)

	if status != pipeline.PipelineSuccess {
		os.Exit(1)
	}

	return nil
}

// createSourceFromPath creates a Source from a path, auto-detecting if it's a directory or zip file
func createSourceFromPath(path string) (build.Source, error) {
	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return build.Source{}, fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		return build.Source{}, fmt.Errorf("failed to access source path: %w", err)
	}

	// Handle directory
	if info.IsDir() {
		slog.Info("using source directory", "path", absPath)
		return build.Source{
			Name:      filepath.Base(absPath),
			Directory: absPath,
		}, nil
	}

	// Handle zip file
	slog.Info("reading source zip", "path", absPath)
	zipData, err := os.ReadFile(absPath)
	if err != nil {
		return build.Source{}, fmt.Errorf("failed to read source file: %w", err)
	}

	return build.Source{
		Name: filepath.Base(absPath),
		Zip:  zipData,
	}, nil
}

// parseArguments converts CLI arguments to pipeline.Arguments
func parseArguments(args []string) pipeline.Arguments {
	argMap := make(pipeline.Arguments)
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			argMap[parts[0]] = parts[1]
		}
	}
	return argMap
}
