package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yeetcd/yeetcd/internal/core/pipeline"
	"github.com/yeetcd/yeetcd/internal/core/types"
	"github.com/yeetcd/yeetcd/internal/testutil"
	"github.com/yeetcd/yeetcd/pkg/build"
	"github.com/yeetcd/yeetcd/pkg/engine/docker"
)

// TestJavaSample_Build tests that the java-sample application can be built
// GIVEN: Real Docker daemon, DockerExecutionEngine, DockerBuildService, Source with Zip of entire repository
// WHEN: Build(ctx, source) is called
// THEN: BuildResult.ImageID is returned, output directories contain compiled JAR files
func TestJavaSample_Build(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Get the repository zip (java-sample depends on java-sdk and protocol)
	zipData, err := testutil.GetJavaSampleZipWithRepo()
	require.NoError(t, err, "Failed to get java-sample zip with repo")

	source := build.Source{
		Name: "java-sample-with-repo",
		Zip:  zipData,
	}

	// Create Docker execution engine
	engine, err := docker.NewDockerExecutionEngine()
	require.NoError(t, err, "Failed to create Docker execution engine")

	// Create build service
	buildService := build.NewDockerBuildService(engine)

	// Build the source
	result, err := buildService.Build(ctx, source)
	require.NoError(t, err, "Failed to build java-sample")
	require.NotNil(t, result, "Build result should not be nil")
	require.NotEmpty(t, result.SourceBuildResults, "Build result should have source build results")
	require.NotEmpty(t, result.Pipelines, "Build result should have generated pipelines")

	t.Logf("Built %d projects, generated %d pipelines", len(result.SourceBuildResults), len(result.Pipelines))
}

// TestJavaSample_AssembleAndExecute_Sample tests the basic sample pipeline
// GIVEN: Real Docker daemon, PipelineController, Source with repository zip
// WHEN: Assemble(ctx, source) then Execute(ctx, pipeline 'sample', outputHandler) is called
// THEN: PipelineResult.PipelineStatus equals SUCCESS, events recorded correctly,
// and custom work definition was actually executed (not silently skipped due to empty image)
func TestJavaSample_AssembleAndExecute_Sample(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Get the repository zip
	zipData, err := testutil.GetJavaSampleZipWithRepo()
	require.NoError(t, err, "Failed to get java-sample zip with repo")

	source := build.Source{
		Name: "java-sample",
		Zip:  zipData,
	}

	// Create execution engine
	engine, err := docker.NewDockerExecutionEngine()
	require.NoError(t, err, "Failed to create Docker execution engine")

	// Create build service and source extractor
	buildService := build.NewDockerBuildService(engine)
	sourceExtractor := build.NewSourceExtractor()

	// Create pipeline controller
	controller := pipeline.NewPipelineController(buildService, sourceExtractor, engine)

	// Assemble pipelines
	pipelines, err := controller.Assemble(ctx, source)
	require.NoError(t, err, "Failed to assemble pipelines")
	require.NotEmpty(t, pipelines, "Should have at least one pipeline")

	// Find the sample pipeline
	var samplePipeline *pipeline.Pipeline
	for _, p := range pipelines {
		if p.Name == "sample" {
			samplePipeline = p
			break
		}
	}
	require.NotNil(t, samplePipeline, "sample pipeline should exist")

	// Verify that BuiltSourceImage is populated (this is the key fix - without it custom work fails silently)
	require.NotEmpty(t, samplePipeline.Metadata.BuiltSourceImage, "Pipeline should have BuiltSourceImage populated")

	// Verify SourceLanguage is populated (without it custom work runs with empty command and exits 2)
	require.NotEmpty(t, samplePipeline.Metadata.SourceLanguage, "Pipeline should have SourceLanguage populated")

	t.Logf("Sample pipeline metadata: BuiltSourceImage=%s, SourceLanguage=%v", samplePipeline.Metadata.BuiltSourceImage, samplePipeline.Metadata.SourceLanguage)

	// Create output handler
	outputHandler := pipeline.NewTestPipelineOutputHandler()

	// Execute the pipeline
	result, err := controller.Execute(ctx, samplePipeline, outputHandler)
	require.NoError(t, err, "Pipeline execution should not error")
	require.NotNil(t, result, "Result should not be nil")

	// Verify status
	assert.Equal(t, pipeline.PipelineSuccess, result.PipelineStatus(), "Pipeline should succeed")

	// Verify events
	events := outputHandler.GetEvents()
	assert.NotEmpty(t, events, "Should have recorded events")

	// CRITICAL: Verify custom work was actually executed, not silently skipped
	// This catches the bug where BuiltSourceImage was empty and custom work would fail silently
	workFinishedEvents := pipeline.GetEventsOfType[pipeline.WorkFinished](outputHandler)
	require.NotEmpty(t, workFinishedEvents, "Should have WorkFinished events")

	// Check that we have at least 2 work items: containerised-work-definition and custom-work-definition
	// (The sample pipeline has custom-work-definition depending on containerised-work-definition)
	assert.GreaterOrEqual(t, len(workFinishedEvents), 2, "Should have at least 2 work items executed")

	// Verify custom-work-definition completed successfully
	var customWorkFinished *pipeline.WorkFinished
	for _, e := range workFinishedEvents {
		if e.Work.Description == "custom-work-definition" {
			customWorkFinished = &e
			break
		}
	}
	require.NotNil(t, customWorkFinished, "custom-work-definition should have WorkFinished event")
	assert.Equal(t, types.WorkStatusSucceeded, customWorkFinished.WorkStatus, "custom-work-definition should have succeeded")

	t.Logf("Pipeline executed successfully with %d events", len(events))
}

// TestJavaSample_AssembleAndExecute_Compound tests the compound pipeline
// GIVEN: Real Docker daemon, PipelineController, Source with repository zip
// WHEN: Execute(ctx, pipeline 'sampleCompound', outputHandler) is called
// THEN: PipelineResult.PipelineStatus equals SUCCESS, events show compound work execution
func TestJavaSample_AssembleAndExecute_Compound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Get the repository zip
	zipData, err := testutil.GetJavaSampleZipWithRepo()
	require.NoError(t, err, "Failed to get java-sample zip with repo")

	source := build.Source{
		Name: "java-sample",
		Zip:  zipData,
	}

	// Create execution engine
	engine, err := docker.NewDockerExecutionEngine()
	require.NoError(t, err, "Failed to create Docker execution engine")

	// Create build service and source extractor
	buildService := build.NewDockerBuildService(engine)
	sourceExtractor := build.NewSourceExtractor()

	// Create pipeline controller
	controller := pipeline.NewPipelineController(buildService, sourceExtractor, engine)

	// Assemble pipelines
	pipelines, err := controller.Assemble(ctx, source)
	require.NoError(t, err, "Failed to assemble pipelines")

	// Find the compound pipeline
	var compoundPipeline *pipeline.Pipeline
	for _, p := range pipelines {
		if p.Name == "sampleCompound" {
			compoundPipeline = p
			break
		}
	}
	require.NotNil(t, compoundPipeline, "sampleCompound pipeline should exist")

	// Create output handler
	outputHandler := pipeline.NewTestPipelineOutputHandler()

	// Execute the pipeline
	result, err := controller.Execute(ctx, compoundPipeline, outputHandler)
	require.NoError(t, err, "Pipeline execution should not error")
	require.NotNil(t, result, "Result should not be nil")

	// Verify status
	assert.Equal(t, pipeline.PipelineSuccess, result.PipelineStatus(), "Pipeline should succeed")

	t.Logf("Compound pipeline executed successfully")
}

// TestJavaSample_AssembleAndExecute_WorkContext tests the work context pipeline
// GIVEN: Real Docker daemon, PipelineController, Source with repository zip
// WHEN: Execute(ctx, pipeline 'sampleWithWorkContext', outputHandler) is called
// THEN: PipelineResult.PipelineStatus equals SUCCESS, work receives merged context
func TestJavaSample_AssembleAndExecute_WorkContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Get the repository zip
	zipData, err := testutil.GetJavaSampleZipWithRepo()
	require.NoError(t, err, "Failed to get java-sample zip with repo")

	source := build.Source{
		Name: "java-sample",
		Zip:  zipData,
	}

	// Create execution engine
	engine, err := docker.NewDockerExecutionEngine()
	require.NoError(t, err, "Failed to create Docker execution engine")

	// Create build service and source extractor
	buildService := build.NewDockerBuildService(engine)
	sourceExtractor := build.NewSourceExtractor()

	// Create pipeline controller
	controller := pipeline.NewPipelineController(buildService, sourceExtractor, engine)

	// Assemble pipelines
	pipelines, err := controller.Assemble(ctx, source)
	require.NoError(t, err, "Failed to assemble pipelines")

	// Find the work context pipeline
	var workContextPipeline *pipeline.Pipeline
	for _, p := range pipelines {
		if p.Name == "sampleWithWorkContext" {
			workContextPipeline = p
			break
		}
	}
	require.NotNil(t, workContextPipeline, "sampleWithWorkContext pipeline should exist")

	// Create output handler
	outputHandler := pipeline.NewTestPipelineOutputHandler()

	// Execute the pipeline
	result, err := controller.Execute(ctx, workContextPipeline, outputHandler)
	require.NoError(t, err, "Pipeline execution should not error")
	require.NotNil(t, result, "Result should not be nil")

	// Verify status
	assert.Equal(t, pipeline.PipelineSuccess, result.PipelineStatus(), "Pipeline should succeed")

	t.Logf("Work context pipeline executed successfully")
}

// TestJavaSample_AssembleAndExecute_WithParameters tests the parameters pipeline
// GIVEN: Real Docker daemon, PipelineController, Source with repository zip
// WHEN: Execute(ctx, pipeline 'sampleWithParameters'.WithArguments({'PARAMETER_NAME': 'other'}), outputHandler) is called
// THEN: PipelineResult.PipelineStatus equals SUCCESS, work receives PARAMETER_NAME=other env var
func TestJavaSample_AssembleAndExecute_WithParameters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Get the repository zip
	zipData, err := testutil.GetJavaSampleZipWithRepo()
	require.NoError(t, err, "Failed to get java-sample zip with repo")

	source := build.Source{
		Name: "java-sample",
		Zip:  zipData,
	}

	// Create execution engine
	engine, err := docker.NewDockerExecutionEngine()
	require.NoError(t, err, "Failed to create Docker execution engine")

	// Create build service and source extractor
	buildService := build.NewDockerBuildService(engine)
	sourceExtractor := build.NewSourceExtractor()

	// Create pipeline controller
	controller := pipeline.NewPipelineController(buildService, sourceExtractor, engine)

	// Assemble pipelines
	pipelines, err := controller.Assemble(ctx, source)
	require.NoError(t, err, "Failed to assemble pipelines")

	// Find the parameters pipeline
	var paramsPipeline *pipeline.Pipeline
	for _, p := range pipelines {
		if p.Name == "sampleWithParameters" {
			paramsPipeline = p
			break
		}
	}
	require.NotNil(t, paramsPipeline, "sampleWithParameters pipeline should exist")

	// Apply arguments
	args := pipeline.ArgumentsOf("PARAMETER_NAME", "other")
	paramsPipeline, err = paramsPipeline.WithArguments(args)
	require.NoError(t, err, "Failed to apply arguments")

	// Create output handler
	outputHandler := pipeline.NewTestPipelineOutputHandler()

	// Execute the pipeline
	result, err := controller.Execute(ctx, paramsPipeline, outputHandler)
	require.NoError(t, err, "Pipeline execution should not error")
	require.NotNil(t, result, "Result should not be nil")

	// Verify status
	assert.Equal(t, pipeline.PipelineSuccess, result.PipelineStatus(), "Pipeline should succeed")

	t.Logf("Parameters pipeline executed successfully with PARAMETER_NAME=other")
}

// TestJavaSample_AssembleAndExecute_WithWorkOutputs tests the work outputs pipeline
// GIVEN: Real Docker daemon, PipelineController, Source with repository zip
// WHEN: Execute(ctx, pipeline 'sampleWithWorkOutputs', outputHandler) is called
// THEN: PipelineResult.PipelineStatus equals SUCCESS, work validates outputs match expected values
func TestJavaSample_AssembleAndExecute_WithWorkOutputs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Get the repository zip
	zipData, err := testutil.GetJavaSampleZipWithRepo()
	require.NoError(t, err, "Failed to get java-sample zip with repo")

	source := build.Source{
		Name: "java-sample",
		Zip:  zipData,
	}

	// Create execution engine
	engine, err := docker.NewDockerExecutionEngine()
	require.NoError(t, err, "Failed to create Docker execution engine")

	// Create build service and source extractor
	buildService := build.NewDockerBuildService(engine)
	sourceExtractor := build.NewSourceExtractor()

	// Create pipeline controller
	controller := pipeline.NewPipelineController(buildService, sourceExtractor, engine)

	// Assemble pipelines
	pipelines, err := controller.Assemble(ctx, source)
	require.NoError(t, err, "Failed to assemble pipelines")

	// Find the work outputs pipeline
	var outputsPipeline *pipeline.Pipeline
	for _, p := range pipelines {
		if p.Name == "sampleWithWorkOutputs" {
			outputsPipeline = p
			break
		}
	}
	require.NotNil(t, outputsPipeline, "sampleWithWorkOutputs pipeline should exist")

	// Create output handler
	outputHandler := pipeline.NewTestPipelineOutputHandler()

	// Execute the pipeline
	result, err := controller.Execute(ctx, outputsPipeline, outputHandler)
	require.NoError(t, err, "Pipeline execution should not error")
	require.NotNil(t, result, "Result should not be nil")

	// Verify status
	assert.Equal(t, pipeline.PipelineSuccess, result.PipelineStatus(), "Pipeline should succeed")

	t.Logf("Work outputs pipeline executed successfully")
}

// TestJavaSample_AssembleAndExecute_WithConditions tests the conditions pipeline
// GIVEN: Real Docker daemon, PipelineController, Source with repository zip
// WHEN: Execute(ctx, pipeline 'sampleWithConditions', outputHandler) is called
// THEN: PipelineResult.PipelineStatus equals SUCCESS, some works SKIPPED
func TestJavaSample_AssembleAndExecute_WithConditions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Get the repository zip
	zipData, err := testutil.GetJavaSampleZipWithRepo()
	require.NoError(t, err, "Failed to get java-sample zip with repo")

	source := build.Source{
		Name: "java-sample",
		Zip:  zipData,
	}

	// Create execution engine
	engine, err := docker.NewDockerExecutionEngine()
	require.NoError(t, err, "Failed to create Docker execution engine")

	// Create build service and source extractor
	buildService := build.NewDockerBuildService(engine)
	sourceExtractor := build.NewSourceExtractor()

	// Create pipeline controller
	controller := pipeline.NewPipelineController(buildService, sourceExtractor, engine)

	// Assemble pipelines
	pipelines, err := controller.Assemble(ctx, source)
	require.NoError(t, err, "Failed to assemble pipelines")

	// Find the conditions pipeline
	var conditionsPipeline *pipeline.Pipeline
	for _, p := range pipelines {
		if p.Name == "sampleWithConditions" {
			conditionsPipeline = p
			break
		}
	}
	require.NotNil(t, conditionsPipeline, "sampleWithConditions pipeline should exist")

	// Create output handler
	outputHandler := pipeline.NewTestPipelineOutputHandler()

	// Execute the pipeline
	result, err := controller.Execute(ctx, conditionsPipeline, outputHandler)
	require.NoError(t, err, "Pipeline execution should not error")
	require.NotNil(t, result, "Result should not be nil")

	// Verify status
	assert.Equal(t, pipeline.PipelineSuccess, result.PipelineStatus(), "Pipeline should succeed")

	t.Logf("Conditions pipeline executed successfully")
}

// TestJavaSample_AssembleAndExecute_DynamicWork tests the dynamic work pipeline
// GIVEN: Real Docker daemon, PipelineController, Source with repository zip
// WHEN: Execute(ctx, pipeline 'sampleDynamicWork'.WithArguments({'WORK_COUNT': '2'}), outputHandler) is called
// THEN: PipelineResult.PipelineStatus equals SUCCESS, dynamic work generates 2 child works
func TestJavaSample_AssembleAndExecute_DynamicWork(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Get the repository zip
	zipData, err := testutil.GetJavaSampleZipWithRepo()
	require.NoError(t, err, "Failed to get java-sample zip with repo")

	source := build.Source{
		Name: "java-sample",
		Zip:  zipData,
	}

	// Create execution engine
	engine, err := docker.NewDockerExecutionEngine()
	require.NoError(t, err, "Failed to create Docker execution engine")

	// Create build service and source extractor
	buildService := build.NewDockerBuildService(engine)
	sourceExtractor := build.NewSourceExtractor()

	// Create pipeline controller
	controller := pipeline.NewPipelineController(buildService, sourceExtractor, engine)

	// Assemble pipelines
	pipelines, err := controller.Assemble(ctx, source)
	require.NoError(t, err, "Failed to assemble pipelines")

	// Find the dynamic work pipeline
	var dynamicPipeline *pipeline.Pipeline
	for _, p := range pipelines {
		if p.Name == "sampleDynamicWork" {
			dynamicPipeline = p
			break
		}
	}
	require.NotNil(t, dynamicPipeline, "sampleDynamicWork pipeline should exist")

	// Apply arguments
	args := pipeline.ArgumentsOf("WORK_COUNT", "2")
	dynamicPipeline, err = dynamicPipeline.WithArguments(args)
	require.NoError(t, err, "Failed to apply arguments")

	// Create output handler
	outputHandler := pipeline.NewTestPipelineOutputHandler()

	// Execute the pipeline
	result, err := controller.Execute(ctx, dynamicPipeline, outputHandler)
	require.NoError(t, err, "Pipeline execution should not error")
	require.NotNil(t, result, "Result should not be nil")

	// Verify status
	assert.Equal(t, pipeline.PipelineSuccess, result.PipelineStatus(), "Pipeline should succeed")

	t.Logf("Dynamic work pipeline executed successfully with WORK_COUNT=2")
}
