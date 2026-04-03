package build

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	pb "github.com/yeetcd/yeetcd/internal/core/proto/pipeline"
	"github.com/yeetcd/yeetcd/pkg/config"
	"github.com/yeetcd/yeetcd/pkg/engine"
	"google.golang.org/protobuf/proto"
)

// DockerBuildService implements BuildService using Docker
type DockerBuildService struct {
	engine engine.ExecutionEngine
}

// NewDockerBuildService creates a new Docker build service
func NewDockerBuildService(eng engine.ExecutionEngine) *DockerBuildService {
	return &DockerBuildService{engine: eng}
}

// Build builds the source code and returns the result
func (d *DockerBuildService) Build(ctx context.Context, source Source) (*BuildResult, error) {
	// Extract the source zip
	extractor := NewSourceExtractor()
	extractionResult, err := extractor.Extract(source)
	if err != nil {
		return nil, fmt.Errorf("failed to extract source: %w", err)
	}
	defer extractionResult.Close()

	// Build each project defined by yeetcd.yaml files
	var sourceBuildResults []SourceBuildResult
	var allPipelines []*pb.Pipeline

	for projectPath, yeetcdConfig := range extractionResult.YeetcdDefinitions {
		// Build this project
		buildResult, err := d.buildProject(ctx, extractionResult.Directory, projectPath, yeetcdConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build project %s: %w", yeetcdConfig.Name, err)
		}

		// Generate pipeline definitions by running the built container
		pipelines, imageID, err := d.generatePipelines(ctx, yeetcdConfig, buildResult)
		if err != nil {
			return nil, fmt.Errorf("failed to generate pipelines for %s: %w", yeetcdConfig.Name, err)
		}

		// Store the image ID in the source build result for later use (custom work execution)
		buildResult.ImageID = imageID
		sourceBuildResults = append(sourceBuildResults, *buildResult)

		allPipelines = append(allPipelines, pipelines...)
	}

	return &BuildResult{
		ImageID:            "", // No single image ID for multi-project builds
		Pipelines:          allPipelines,
		SourceBuildResults: sourceBuildResults,
	}, nil
}

// buildProject builds a single project defined by a yeetcd.yaml file
func (d *DockerBuildService) buildProject(ctx context.Context, extractionDir, projectPath string, yeetcdConfig config.YeetcdConfig) (*SourceBuildResult, error) {
	// Source mount directory
	sourceMountDir := "/var/yeetcd"

	// Working directory is the project root (parent of yeetcd.yaml)
	workingDir := filepath.Join(sourceMountDir, projectPath)
	if projectPath == "." {
		workingDir = sourceMountDir
	}

	// Build output directory paths from artifacts
	outputDirectoryPaths := make(map[string]string)
	for _, artifact := range yeetcdConfig.Artifacts {
		outputDirectoryPaths[artifact.Name] = filepath.Join(workingDir, artifact.Path)
	}

	// Get user's home directory for .m2 cache
	homeDir, err := getHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create job definition
	jobDef := engine.JobDefinition{
		Image:      yeetcdConfig.BuildImage,
		Cmd:        strings.Fields(yeetcdConfig.BuildCmd),
		WorkingDir: workingDir,
		InputFilePaths: map[string]engine.MountInput{
			sourceMountDir: engine.OnDiskMountInput{Dir: extractionDir},
			"/root/.m2":    engine.OnDiskMountInput{Dir: filepath.Join(homeDir, ".m2")},
		},
		OutputDirectoryPaths: outputDirectoryPaths,
		JobStreams:           engine.NewJobStreams(os.Stdout, os.Stderr),
	}

	// Run the build job
	jobResult, err := d.engine.RunJob(ctx, jobDef)
	if err != nil {
		return nil, fmt.Errorf("build job failed: %w", err)
	}

	if jobResult.ExitCode != 0 {
		return nil, fmt.Errorf("build job exited with code %d", jobResult.ExitCode)
	}

	return &SourceBuildResult{
		YeetcdConfig:            yeetcdConfig,
		OutputDirectoriesParent: jobResult.OutputDirectoriesParent,
	}, nil
}

// generatePipelines runs the built container to generate protobuf pipeline definitions
// Returns the pipelines and the image ID (which should NOT be deleted as it's needed for custom work execution)
func (d *DockerBuildService) generatePipelines(ctx context.Context, yeetcdConfig config.YeetcdConfig, buildResult *SourceBuildResult) ([]*pb.Pipeline, string, error) {
	// Get the command to generate pipeline definitions
	generateCmd := yeetcdConfig.Language.GetGeneratePipelineDefinitionsCmd()
	if generateCmd == nil {
		return nil, "", fmt.Errorf("unsupported language: %s", yeetcdConfig.Language)
	}

	// Build artifact paths for the classpath
	var artifactNames []string
	for _, artifact := range yeetcdConfig.Artifacts {
		artifactNames = append(artifactNames, artifact.Name)
	}

	// Get image base
	imageBase := yeetcdConfig.Language.GetImageBase()

	// Build the image with the compiled artifacts
	buildImageDef := engine.BuildImageDefinition{
		Image:             fmt.Sprintf("yeetcd-%s", yeetcdConfig.Name),
		Tag:               "latest",
		ImageBase:         imageBase,
		ArtifactDirectory: buildResult.OutputDirectoriesParent,
		ArtifactNames:     artifactNames,
		Cmd:               strings.Join(generateCmd, " "),
	}

	buildImageResult, err := d.engine.BuildImage(ctx, buildImageDef)
	if err != nil {
		return nil, "", fmt.Errorf("failed to build pipeline generator image: %w", err)
	}

	// Create JobStreams to capture stdout
	streams := engine.NewJobStreams(nil, os.Stderr)

	// Run the container to generate pipeline definitions
	// Pass empty command to use the CMD from the image (which is the class name)
	jobDef := engine.JobDefinition{
		Image:      buildImageResult.ImageID,
		Cmd:        []string{},
		JobStreams: streams,
	}

	jobResult, err := d.engine.RunJob(ctx, jobDef)
	if err != nil {
		return nil, "", fmt.Errorf("failed to run pipeline generator: %w", err)
	}

	if jobResult.ExitCode != 0 {
		return nil, "", fmt.Errorf("pipeline generator exited with code %d", jobResult.ExitCode)
	}

	// Parse the protobuf output from stdout
	stdout := streams.GetStdOut()

	// Debug: log stdout length and first few bytes
	fmt.Fprintf(os.Stderr, "DEBUG: stdout length: %d\n", len(stdout))
	if len(stdout) > 0 {
		fmt.Fprintf(os.Stderr, "DEBUG: stdout first 100 bytes: %v\n", stdout[:min(100, len(stdout))])
	}

	pipelines, err := d.parseProtobufPipelines(stdout)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse pipeline definitions: %w", err)
	}

	// Don't remove the image - it's needed for custom work execution
	// The image will be cleaned up separately when no longer needed

	return pipelines, buildImageResult.ImageID, nil
}

// parseProtobufPipelines parses protobuf pipeline definitions from the output
func (d *DockerBuildService) parseProtobufPipelines(stdout []byte) ([]*pb.Pipeline, error) {
	// Parse the protobuf Pipelines message
	var pbPipelines pb.Pipelines
	if err := proto.Unmarshal(stdout, &pbPipelines); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protobuf: %w", err)
	}

	return pbPipelines.GetPipelines(), nil
}

// getHomeDir returns the user's home directory
func getHomeDir() (string, error) {
	// First try the HOME environment variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// Fall back to user.Current()
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return currentUser.HomeDir, nil
}
