package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/yeetcd/yeetcd/internal/core/proto/pipeline"
)

// TestPipeline_FromProtobuf converts protobuf Pipeline message to Go Pipeline struct with all fields correctly mapped
// GIVEN: Valid protobuf Pipeline message with name, parameters, workContext, and finalWork
// WHEN: FromProtobuf() is called
// THEN: Go Pipeline struct is created with all fields correctly mapped from protobuf
func TestPipeline_FromProtobuf(t *testing.T) {
	// Create a protobuf pipeline message
	defaultValue := "default"
	protoPipeline := &pb.Pipeline{
		Name: "test-pipeline",
		Parameters: map[string]*pb.Parameter{
			"param1": {
				TypeCheck:    pb.Parameter_STRING,
				Required:     true,
				DefaultValue: &defaultValue,
			},
		},
		WorkContext: map[string]string{
			"key1": "value1",
		},
		FinalWork: []*pb.Work{
			{
				Id:          "work1",
				Description: "Test work",
				OneofTaskActions: &pb.Work_ContainerisedWorkDefinition{
					ContainerisedWorkDefinition: &pb.ContainerisedWorkDefinition{
						Image: "test-image:latest",
						Cmd:   []string{"echo", "hello"},
					},
				},
			},
		},
	}

	// Call FromProtobuf
	pipeline, err := FromProtobuf(protoPipeline)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, pipeline)
	assert.Equal(t, "test-pipeline", pipeline.Name)
	assert.NotNil(t, pipeline.Parameters)
	assert.NotNil(t, pipeline.WorkContext)
	assert.NotNil(t, pipeline.FinalWork)
}

// TestPipeline_WithArguments merges arguments into pipeline work context with correct override behavior
// GIVEN: Pipeline with existing workContext and Parameters definition
// WHEN: WithArguments() is called with Arguments containing override values
// THEN: New Pipeline is returned with arguments merged into workContext, arguments override existing context values
func TestPipeline_WithArguments(t *testing.T) {
	// Create a pipeline with existing work context
	pipeline := &Pipeline{
		Name: "test-pipeline",
		Parameters: Parameters{
			"param1": {TypeCheck: STRING, Required: true},
		},
		WorkContext: WorkContext{
			"key1": "original",
			"key2": "value2",
		},
	}

	// Create arguments with override
	args := ArgumentsOf("param1", "value1", "key1", "overridden")

	// Call WithArguments
	newPipeline, err := pipeline.WithArguments(args)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, newPipeline)
	assert.Equal(t, "overridden", newPipeline.WorkContext["key1"])
	assert.Equal(t, "value2", newPipeline.WorkContext["key2"])
}
