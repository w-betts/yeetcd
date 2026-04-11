package sample

import (
	"testing"

	"github.com/stretchr/testify/assert"
	sdk "github.com/yeetcd/yeetcd/sdk/pkg/yeetcd"
)

func TestSamplePipeline(t *testing.T) {
	pipeline := samplePipeline()

	assert.Equal(t, "sample", pipeline.Name)
	assert.Len(t, pipeline.FinalWork, 1)
	assert.Equal(t, "containerised-work-definition", pipeline.FinalWork[0].Description)
}

func TestSampleCompoundPipeline(t *testing.T) {
	pipeline := sampleCompoundPipeline()

	assert.Equal(t, "sampleCompound", pipeline.Name)
	assert.Len(t, pipeline.FinalWork, 1)

	// Check that the final work has previous work
	assert.Len(t, pipeline.FinalWork[0].PreviousWork, 1)
}

func TestSampleWithWorkContextPipeline(t *testing.T) {
	pipeline := sampleWithWorkContextPipeline()

	assert.Equal(t, "sampleWithWorkContext", pipeline.Name)
	assert.Equal(t, "pipelineWorkContext", pipeline.WorkContext["PIPELINE_WORK_CONTEXT"])

	// Work should have work-level context merged with pipeline context
	assert.Equal(t, "workWorkContext", pipeline.FinalWork[0].WorkContext["WORK_WORK_CONTEXT"])
}

func TestSampleWithParametersPipeline(t *testing.T) {
	pipeline := sampleWithParametersPipeline()

	assert.Equal(t, "sampleWithParameters", pipeline.Name)
	assert.NotNil(t, pipeline.Parameters)

	param := pipeline.Parameters["PARAMETER_NAME"]
	assert.Equal(t, sdk.TypeCheckString, param.TypeCheck)
	assert.True(t, param.Required)
	assert.Equal(t, "default", param.DefaultValue)
	assert.Equal(t, []string{"default", "other"}, param.Choices)
}

func TestSampleWithConditionsPipeline(t *testing.T) {
	pipeline := sampleWithConditionsPipeline()

	assert.Equal(t, "sampleWithConditions", pipeline.Name)

	// The final work should have a condition
	assert.NotNil(t, pipeline.FinalWork[0].Condition)

	// And previous work
	assert.Len(t, pipeline.FinalWork[0].PreviousWork, 1)
}

func TestSampleWithCustomWorkPipeline(t *testing.T) {
	pipeline := sampleWithCustomWorkPipeline()

	assert.Equal(t, "sampleWithCustomWork", pipeline.Name)
	assert.Len(t, pipeline.FinalWork, 1)

	// Verify custom work can be executed
	customWork := pipeline.FinalWork[0].WorkDefinition.(*sdk.CustomWorkDefinition)
	assert.NotNil(t, customWork)

	// Should not panic when running
	customWork.Run()
}

func TestSampleWithCompoundPipeline(t *testing.T) {
	pipeline := sampleWithCompoundPipeline()

	assert.Equal(t, "sampleWithCompound", pipeline.Name)
	assert.Len(t, pipeline.FinalWork, 1)
}
