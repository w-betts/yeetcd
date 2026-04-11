package sample

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/yeetcd/yeetcd/pkg/proto/pipeline"
	"google.golang.org/protobuf/proto"
)

// TestGeneratorProducesValidProtobuf tests that the generator produces valid protobuf output.
func TestGeneratorProducesValidProtobuf(t *testing.T) {
	// Build the generator
	tmpDir := t.TempDir()
	generatorPath := filepath.Join(tmpDir, "generator")

	buildCmd := exec.Command("go", "build", "-o", generatorPath, "github.com/yeetcd/yeetcd/sdk/generator/cmd/generate")
	output, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "Failed to build generator: %s", string(output))

	// Run the generator on the sample project
	sampleDir, err := os.Getwd()
	require.NoError(t, err)

	runCmd := exec.Command(generatorPath)
	runCmd.Dir = sampleDir
	output, err = runCmd.Output()
	require.NoError(t, err, "Generator failed: %s", string(output))

	// Verify the output is valid protobuf
	var pipelines pb.Pipelines
	err = proto.Unmarshal(output, &pipelines)
	require.NoError(t, err, "Failed to unmarshal protobuf: %v", err)

	// Verify we got all 7 pipelines
	assert.Len(t, pipelines.Pipelines, 7, "Expected 7 pipelines")

	// Verify each pipeline has expected structure
	pipelineNames := make(map[string]bool)
	for _, p := range pipelines.Pipelines {
		pipelineNames[p.Name] = true
		assert.NotEmpty(t, p.Name, "Pipeline name should not be empty")
		assert.NotEmpty(t, p.FinalWork, "Pipeline should have final work")
	}

	// Verify all expected pipelines are present
	expectedPipelines := []string{
		"sample",
		"sampleCompound",
		"sampleWithWorkContext",
		"sampleWithParameters",
		"sampleWithConditions",
		"sampleWithCustomWork",
		"sampleWithCompound",
	}
	for _, name := range expectedPipelines {
		assert.True(t, pipelineNames[name], "Expected pipeline %s to be present", name)
	}
}

// TestGeneratorPipelineStructure tests that generated pipelines have correct structure.
func TestGeneratorPipelineStructure(t *testing.T) {
	// Build and run generator
	tmpDir := t.TempDir()
	generatorPath := filepath.Join(tmpDir, "generator")

	buildCmd := exec.Command("go", "build", "-o", generatorPath, "github.com/yeetcd/yeetcd/sdk/generator/cmd/generate")
	output, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "Failed to build generator: %s", string(output))

	sampleDir, err := os.Getwd()
	require.NoError(t, err)

	runCmd := exec.Command(generatorPath)
	runCmd.Dir = sampleDir
	output, err = runCmd.Output()
	require.NoError(t, err, "Generator failed: %s", string(output))

	var pipelines pb.Pipelines
	err = proto.Unmarshal(output, &pipelines)
	require.NoError(t, err, "Failed to unmarshal protobuf: %v", err)

	// Find the sample pipeline and verify its structure
	var samplePipeline *pb.Pipeline
	for _, p := range pipelines.Pipelines {
		if p.Name == "sample" {
			samplePipeline = p
			break
		}
	}
	require.NotNil(t, samplePipeline, "sample pipeline not found")

	// Verify work structure
	require.Len(t, samplePipeline.FinalWork, 1)
	work := samplePipeline.FinalWork[0]
	assert.Equal(t, "containerised-work-definition", work.Description)
	assert.NotNil(t, work.GetContainerisedWorkDefinition(), "Expected containerised work definition")
	assert.Equal(t, "maven:3.9.9-eclipse-temurin-17", work.GetContainerisedWorkDefinition().Image)

	// Find sampleWithParameters and verify parameter structure
	var paramsPipeline *pb.Pipeline
	for _, p := range pipelines.Pipelines {
		if p.Name == "sampleWithParameters" {
			paramsPipeline = p
			break
		}
	}
	require.NotNil(t, paramsPipeline, "sampleWithParameters pipeline not found")
	require.NotNil(t, paramsPipeline.Parameters, "Parameters should not be nil")
	assert.Contains(t, paramsPipeline.Parameters, "PARAMETER_NAME")

	param := paramsPipeline.Parameters["PARAMETER_NAME"]
	assert.Equal(t, pb.Parameter_STRING, param.TypeCheck)
	assert.True(t, param.Required)
	require.NotNil(t, param.DefaultValue)
	assert.Equal(t, "default", *param.DefaultValue)
	assert.Equal(t, []string{"default", "other"}, param.Choices)
}

// TestGeneratorWorkContext tests that work context is properly serialized.
func TestGeneratorWorkContext(t *testing.T) {
	// Build and run generator
	tmpDir := t.TempDir()
	generatorPath := filepath.Join(tmpDir, "generator")

	buildCmd := exec.Command("go", "build", "-o", generatorPath, "github.com/yeetcd/yeetcd/sdk/generator/cmd/generate")
	output, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "Failed to build generator: %s", string(output))

	sampleDir, err := os.Getwd()
	require.NoError(t, err)

	runCmd := exec.Command(generatorPath)
	runCmd.Dir = sampleDir
	output, err = runCmd.Output()
	require.NoError(t, err, "Generator failed: %s", string(output))

	var pipelines pb.Pipelines
	err = proto.Unmarshal(output, &pipelines)
	require.NoError(t, err, "Failed to unmarshal protobuf: %v", err)

	// Find sampleWithWorkContext pipeline
	var ctxPipeline *pb.Pipeline
	for _, p := range pipelines.Pipelines {
		if p.Name == "sampleWithWorkContext" {
			ctxPipeline = p
			break
		}
	}
	require.NotNil(t, ctxPipeline, "sampleWithWorkContext pipeline not found")

	// Verify pipeline-level work context
	assert.Equal(t, "pipelineWorkContext", ctxPipeline.WorkContext["PIPELINE_WORK_CONTEXT"])

	// Verify work-level work context
	require.Len(t, ctxPipeline.FinalWork, 1)
	work := ctxPipeline.FinalWork[0]
	assert.Equal(t, "workWorkContext", work.WorkContext["WORK_WORK_CONTEXT"])
}

// TestGeneratorConditions tests that conditions are properly serialized.
func TestGeneratorConditions(t *testing.T) {
	// Build and run generator
	tmpDir := t.TempDir()
	generatorPath := filepath.Join(tmpDir, "generator")

	buildCmd := exec.Command("go", "build", "-o", generatorPath, "github.com/yeetcd/yeetcd/sdk/generator/cmd/generate")
	output, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "Failed to build generator: %s", string(output))

	sampleDir, err := os.Getwd()
	require.NoError(t, err)

	runCmd := exec.Command(generatorPath)
	runCmd.Dir = sampleDir
	output, err = runCmd.Output()
	require.NoError(t, err, "Generator failed: %s", string(output))

	var pipelines pb.Pipelines
	err = proto.Unmarshal(output, &pipelines)
	require.NoError(t, err, "Failed to unmarshal protobuf: %v", err)

	// Find sampleWithConditions pipeline
	var condPipeline *pb.Pipeline
	for _, p := range pipelines.Pipelines {
		if p.Name == "sampleWithConditions" {
			condPipeline = p
			break
		}
	}
	require.NotNil(t, condPipeline, "sampleWithConditions pipeline not found")

	// Verify the conditional work has a condition
	require.Len(t, condPipeline.FinalWork, 1)
	work := condPipeline.FinalWork[0]
	require.NotNil(t, work.Condition, "Work should have a condition")

	// Verify it's a WorkContextCondition
	wcc := work.Condition.GetWorkContextCondition()
	require.NotNil(t, wcc, "Expected WorkContextCondition")
	assert.Equal(t, "missingKey", wcc.Key)
	assert.Equal(t, pb.WorkContextCondition_EQUALS, wcc.Operand)
	assert.Equal(t, "value", wcc.Value)
}

// TestGeneratorPreviousWork tests that previous work dependencies are properly serialized.
func TestGeneratorPreviousWork(t *testing.T) {
	// Build and run generator
	tmpDir := t.TempDir()
	generatorPath := filepath.Join(tmpDir, "generator")

	buildCmd := exec.Command("go", "build", "-o", generatorPath, "github.com/yeetcd/yeetcd/sdk/generator/cmd/generate")
	output, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "Failed to build generator: %s", string(output))

	sampleDir, err := os.Getwd()
	require.NoError(t, err)

	runCmd := exec.Command(generatorPath)
	runCmd.Dir = sampleDir
	output, err = runCmd.Output()
	require.NoError(t, err, "Generator failed: %s", string(output))

	var pipelines pb.Pipelines
	err = proto.Unmarshal(output, &pipelines)
	require.NoError(t, err, "Failed to unmarshal protobuf: %v", err)

	// Find sampleCompound pipeline
	var compoundPipeline *pb.Pipeline
	for _, p := range pipelines.Pipelines {
		if p.Name == "sampleCompound" {
			compoundPipeline = p
			break
		}
	}
	require.NotNil(t, compoundPipeline, "sampleCompound pipeline not found")

	// Verify the final work has previous work
	require.Len(t, compoundPipeline.FinalWork, 1)
	work := compoundPipeline.FinalWork[0]
	require.Len(t, work.PreviousWork, 1, "Work should have 1 previous work dependency")

	prevWork := work.PreviousWork[0]
	assert.NotNil(t, prevWork.Work, "Previous work should have a Work reference")
	assert.Equal(t, "sample-pipeline-work-1", prevWork.Work.Description)
}
