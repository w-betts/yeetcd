// Package sample provides example pipeline definitions for Yeetcd.
package sample

import (
	sdk "github.com/yeetcd/yeetcd/sdk/pkg/yeetcd"
)

// SamplePipeline demonstrates a basic containerised work definition.
func SamplePipeline() sdk.Pipeline {
	containerisedWork := sdk.NewWork(
		"containerised-work-definition",
		sdk.NewContainerisedWork("maven:3.9.9-eclipse-temurin-17").
			WithCommand("bash", "-c", "echo 'Hello from a containerised task'").
			Build(),
	).Build()

	return sdk.NewPipeline("sample").
		WithFinalWork(containerisedWork).
		Build()
}

// SampleCompoundPipeline demonstrates compound work definitions.
func SampleCompoundPipeline() sdk.Pipeline {
	work1 := sdk.NewWork(
		"sample-pipeline-work-1",
		sdk.NewContainerisedWork("alpine").Build(),
	).
		WithWorkContext(sdk.WorkContextOf("part", "1")).
		Build()

	work2 := sdk.NewWork(
		"sample-pipeline-work-2",
		sdk.NewContainerisedWork("alpine").Build(),
	).
		WithWorkContext(sdk.WorkContextOf("part", "2")).
		WithPreviousWork(sdk.NewPreviousWork(work1).Build()).
		Build()

	return sdk.NewPipeline("sampleCompound").
		WithFinalWork(work2).
		Build()
}

// SampleWithWorkContextPipeline demonstrates work context propagation.
func SampleWithWorkContextPipeline() sdk.Pipeline {
	pipelineWorkContext := sdk.WorkContextOf("PIPELINE_WORK_CONTEXT", "pipelineWorkContext")
	workWorkContext := sdk.WorkContextOf("WORK_WORK_CONTEXT", "workWorkContext")

	work := sdk.NewWork(
		"work-with-context",
		sdk.NewContainerisedWork("alpine").Build(),
	).
		WithWorkContext(workWorkContext).
		Build()

	return sdk.NewPipeline("sampleWithWorkContext").
		WithWorkContext(pipelineWorkContext).
		WithFinalWork(work).
		Build()
}

// SampleWithParametersPipeline demonstrates pipeline parameters.
func SampleWithParametersPipeline() sdk.Pipeline {
	param := sdk.NewParameter(sdk.TypeCheckString).
		WithRequired(true).
		WithDefaultValue("default").
		WithChoices("default", "other").
		Build()

	params := sdk.ParametersOf("PARAMETER_NAME", param)

	work := sdk.NewWork(
		"parameterized-work",
		sdk.NewContainerisedWork("alpine").Build(),
	).Build()

	return sdk.NewPipeline("sampleWithParameters").
		WithParameters(params).
		WithFinalWork(work).
		Build()
}

// SampleWithConditionsPipeline demonstrates work conditions.
func SampleWithConditionsPipeline() sdk.Pipeline {
	unconditionalWork := sdk.NewWork(
		"unconditional-work",
		sdk.NewContainerisedWork("alpine").Build(),
	).
		WithWorkContext(sdk.WorkContextOf("key", "value")).
		WithCondition(sdk.Conditions.WorkContextCondition("key", sdk.OperandEquals, "value")).
		Build()

	conditionalWork := sdk.NewWork(
		"conditional-work",
		sdk.NewContainerisedWork("alpine").Build(),
	).
		WithCondition(sdk.Conditions.WorkContextCondition("missingKey", sdk.OperandEquals, "value")).
		WithPreviousWork(sdk.NewPreviousWork(unconditionalWork).Build()).
		Build()

	return sdk.NewPipeline("sampleWithConditions").
		WithFinalWork(conditionalWork).
		Build()
}

// SampleWithCustomWorkPipeline demonstrates custom work definitions.
func SampleWithCustomWorkPipeline() sdk.Pipeline {
	work := sdk.NewWork(
		"custom-work",
		sdk.NewCustomWork(func() {
			println("Hello from a custom work definition")
		}).Build(),
	).Build()

	return sdk.NewPipeline("sampleWithCustomWork").
		WithFinalWork(work).
		Build()
}

// SampleWithCompoundPipeline demonstrates using a pipeline as compound work.
func SampleWithCompoundPipeline() sdk.Pipeline {
	innerPipeline := SamplePipeline()

	// Use the pipeline as a work (compound work)
	compoundWork := sdk.NewWork(
		"compound-work-from-pipeline",
		sdk.NewCompoundWork(innerPipeline.FinalWork...).Build(),
	).Build()

	return sdk.NewPipeline("sampleWithCompound").
		WithFinalWork(compoundWork).
		Build()
}
