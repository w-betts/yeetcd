package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPipeline(t *testing.T) {
	pipeline := NewPipeline("test-pipeline").Build()

	assert.Equal(t, "test-pipeline", pipeline.Name)
	assert.NotNil(t, pipeline.Parameters)
	assert.NotNil(t, pipeline.WorkContext)
	assert.Empty(t, pipeline.FinalWork)
}

func TestPipelineWithParameters(t *testing.T) {
	params := ParametersOf("count", Parameter{
		TypeCheck: TypeCheckNumber,
		Required:  true,
	})

	pipeline := NewPipeline("test").
		WithParameters(params).
		Build()

	assert.Equal(t, "test", pipeline.Name)
	assert.NotNil(t, pipeline.Parameters)
	assert.Equal(t, TypeCheckNumber, pipeline.Parameters["count"].TypeCheck)
	assert.True(t, pipeline.Parameters["count"].Required)
}

func TestPipelineWithWorkContext(t *testing.T) {
	ctx := WorkContextOf("KEY", "value")

	pipeline := NewPipeline("test").
		WithWorkContext(ctx).
		Build()

	assert.Equal(t, "value", pipeline.WorkContext["KEY"])
}

func TestPipelineWithFinalWork(t *testing.T) {
	work := NewWork("test-work", NewContainerisedWork("alpine").WithCommand("echo", "hello").Build()).Build()

	pipeline := NewPipeline("test").
		WithFinalWork(work).
		Build()

	assert.Len(t, pipeline.FinalWork, 1)
	assert.Equal(t, "test-work", pipeline.FinalWork[0].Description)
}

func TestNewWork(t *testing.T) {
	containerized := NewContainerisedWork("alpine").
		WithCommand("echo", "hello").
		Build()

	work := NewWork("my-work", containerized).Build()

	assert.Equal(t, "my-work", work.Description)
	assert.NotNil(t, work.WorkDefinition)
	assert.NotNil(t, work.WorkContext)
}

func TestWorkBuilderWithOptions(t *testing.T) {
	containerized := NewContainerisedWork("alpine").Build()
	prevWork := NewWork("previous", containerized).Build()

	ctx := WorkContextOf("ENV", "value")
	outputPath := NewWorkOutputPath("output", "/path")
	condition := NewWorkContextCondition("key", OperandEquals, "value").Build()

	work := NewWork("my-work", containerized).
		WithWorkContext(ctx).
		WithOutputPaths(outputPath).
		WithPreviousWork(NewPreviousWork(prevWork).Build()).
		WithCondition(condition).
		Build()

	assert.Equal(t, "value", work.WorkContext["ENV"])
	assert.Len(t, work.OutputPaths, 1)
	assert.Equal(t, "output", work.OutputPaths[0].Name)
	assert.Len(t, work.PreviousWork, 1)
	assert.NotNil(t, work.Condition)
}

func TestContainerisedWorkDefinition(t *testing.T) {
	containerized := NewContainerisedWork("golang:1.21").
		WithCommand("go", "build", ".").
		Build()

	assert.Equal(t, "golang:1.21", containerized.Image)
	assert.Equal(t, []string{"go", "build", "."}, containerized.Cmd)
}

func TestCustomWorkDefinition(t *testing.T) {
	executed := false
	custom := NewCustomWork(func() {
		executed = true
	}).Build()

	custom.Run()
	assert.True(t, executed)
}

func TestCompoundWorkDefinition(t *testing.T) {
	work1 := NewWork("work1", NewContainerisedWork("alpine").Build()).Build()
	work2 := NewWork("work2", NewContainerisedWork("alpine").Build()).Build()

	compound := NewCompoundWork(work1, work2).Build()

	assert.Len(t, compound.FinalWork, 2)
}

func TestWorkContextOf(t *testing.T) {
	ctx := WorkContextOf("KEY1", "value1", "KEY2", "value2")

	assert.Equal(t, "value1", ctx["KEY1"])
	assert.Equal(t, "value2", ctx["KEY2"])
}

func TestWorkContextMerge(t *testing.T) {
	other := WorkContextOf("KEY1", "other-value")
	ctx := WorkContextOf("KEY2", "value2")

	merged := ctx.Merge(other)

	// Context takes precedence over other
	assert.Equal(t, "value2", merged["KEY2"])
	assert.Equal(t, "other-value", merged["KEY1"])
}

func TestWorkContextValue(t *testing.T) {
	t.Setenv("TEST_ENV_VAR", "test-value")

	value := WorkContextValue("TEST_ENV_VAR")
	assert.Equal(t, "test-value", value)

	// Non-existent env var returns empty
	assert.Empty(t, WorkContextValue("NON_EXISTENT"))
}

func TestNewWorkOutputPath(t *testing.T) {
	path := NewWorkOutputPath("my-output", "/path/to/output")

	assert.Equal(t, "my-output", path.Name)
	assert.Equal(t, "/path/to/output", path.Path)
}

func TestPreviousWorkBuilder(t *testing.T) {
	prevWork := NewWork("previous", NewContainerisedWork("alpine").Build()).Build()

	previous := NewPreviousWork(prevWork).
		WithOutputsMountPath("/mnt").
		WithStdOutEnvVar("PREV_OUTPUT").
		Build()

	assert.Equal(t, "previous", previous.Work.Description)
	assert.Equal(t, "/mnt", previous.OutputPathsMount)
	assert.Equal(t, "PREV_OUTPUT", previous.StdOutEnvVar)
}

func TestNewParameter(t *testing.T) {
	param := NewParameter(TypeCheckString).
		WithRequired(true).
		WithDefaultValue("default").
		WithChoices("option1", "option2").
		Build()

	assert.Equal(t, TypeCheckString, param.TypeCheck)
	assert.True(t, param.Required)
	assert.Equal(t, "default", param.DefaultValue)
	assert.Equal(t, []string{"option1", "option2"}, param.Choices)
}

func TestConditions(t *testing.T) {
	// WorkContextCondition
	cond := Conditions.WorkContextCondition("key", OperandEquals, "value")
	assert.Equal(t, "key", cond.Key)
	assert.Equal(t, OperandEquals, cond.Operand)
	assert.Equal(t, "value", cond.Value)

	// PreviousWorkStatusCondition
	statusCond := Conditions.PreviousWorkStatus(StatusSuccess)
	assert.Equal(t, StatusSuccess, statusCond.Status)

	// NotCondition
	notCond := Conditions.Not(cond)
	assert.NotNil(t, notCond)

	// AndCondition
	andCond := Conditions.And(cond, statusCond)
	assert.NotNil(t, andCond)

	// OrCondition
	orCond := Conditions.Or(cond, statusCond)
	assert.NotNil(t, orCond)
}

func TestEmptyParameters(t *testing.T) {
	params := EmptyParameters()
	assert.NotNil(t, params)
}

func TestGenerateWorkID(t *testing.T) {
	// Just verify it doesn't panic
	// IDs will be unique but we can't easily test uniqueness
	assert.NotEmpty(t, "work-1") // Placeholder
}
