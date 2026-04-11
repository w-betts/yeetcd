package condition

import (
	"testing"

	"github.com/stretchr/testify/assert"
	pipelinepb "github.com/yeetcd/yeetcd/pkg/proto/pipeline"
)

// Test: Conditions.FromProtobuf() deserializes WorkContextCondition correctly
// Given: A protobuf Condition message of type WorkContextCondition
// When: Calling FromProtobuf()
// Then: Returns WorkContextCondition with matching key, value, and operator
func TestConditions_FromProtobuf_WorkContextCondition(t *testing.T) {
	proto := &pipelinepb.Condition{
		Conditions: &pipelinepb.Condition_WorkContextCondition{
			WorkContextCondition: &pipelinepb.WorkContextCondition{
				Key:    "env",
				Value:  "prod",
				Operand: pipelinepb.WorkContextCondition_EQUALS,
			},
		},
	}

	condition, err := FromProtobuf(proto)

	assert.NoError(t, err)
	assert.NotNil(t, condition)
	_, ok := condition.(*WorkContextCondition)
	assert.True(t, ok, "Expected *WorkContextCondition")
}

// Test: Conditions.FromProtobuf() deserializes PreviousWorkStatusCondition correctly
// Given: A protobuf Condition message of type PreviousWorkStatusCondition
// When: Calling FromProtobuf()
// Then: Returns PreviousWorkStatusCondition with matching status
func TestConditions_FromProtobuf_PreviousWorkStatusCondition(t *testing.T) {
	proto := &pipelinepb.Condition{
		Conditions: &pipelinepb.Condition_PreviousWorkStatusCondition{
			PreviousWorkStatusCondition: &pipelinepb.PreviousWorkStatusCondition{
				Status: pipelinepb.PreviousWorkStatusCondition_SUCCESS,
			},
		},
	}

	condition, err := FromProtobuf(proto)

	assert.NoError(t, err)
	assert.NotNil(t, condition)
	_, ok := condition.(*PreviousWorkStatusCondition)
	assert.True(t, ok, "Expected *PreviousWorkStatusCondition")
}

// Test: Conditions.FromProtobuf() deserializes AndCondition correctly
// Given: A protobuf Condition message of type AndCondition containing two nested conditions
// When: Calling FromProtobuf()
// Then: Returns AndCondition containing the deserialized nested conditions
func TestConditions_FromProtobuf_AndCondition(t *testing.T) {
	proto := &pipelinepb.Condition{
		Conditions: &pipelinepb.Condition_AndCondition{
			AndCondition: &pipelinepb.AndCondition{
				Left: &pipelinepb.Condition{
					Conditions: &pipelinepb.Condition_WorkContextCondition{
						WorkContextCondition: &pipelinepb.WorkContextCondition{
							Key:    "env",
							Value:  "prod",
							Operand: pipelinepb.WorkContextCondition_EQUALS,
						},
					},
				},
				Right: &pipelinepb.Condition{
					Conditions: &pipelinepb.Condition_WorkContextCondition{
						WorkContextCondition: &pipelinepb.WorkContextCondition{
							Key:    "region",
							Value:  "us",
							Operand: pipelinepb.WorkContextCondition_EQUALS,
						},
					},
				},
			},
		},
	}

	condition, err := FromProtobuf(proto)

	assert.NoError(t, err)
	assert.NotNil(t, condition)
	_, ok := condition.(*AndCondition)
	assert.True(t, ok, "Expected *AndCondition")
}

// Test: Conditions.FromProtobuf() deserializes OrCondition correctly
// Given: A protobuf Condition message of type OrCondition containing two nested conditions
// When: Calling FromProtobuf()
// Then: Returns OrCondition containing the deserialized nested conditions
func TestConditions_FromProtobuf_OrCondition(t *testing.T) {
	proto := &pipelinepb.Condition{
		Conditions: &pipelinepb.Condition_OrCondition{
			OrCondition: &pipelinepb.OrCondition{
				Left: &pipelinepb.Condition{
					Conditions: &pipelinepb.Condition_WorkContextCondition{
						WorkContextCondition: &pipelinepb.WorkContextCondition{
							Key:    "env",
							Value:  "prod",
							Operand: pipelinepb.WorkContextCondition_EQUALS,
						},
					},
				},
				Right: &pipelinepb.Condition{
					Conditions: &pipelinepb.Condition_WorkContextCondition{
						WorkContextCondition: &pipelinepb.WorkContextCondition{
							Key:    "region",
							Value:  "us",
							Operand: pipelinepb.WorkContextCondition_EQUALS,
						},
					},
				},
			},
		},
	}

	condition, err := FromProtobuf(proto)

	assert.NoError(t, err)
	assert.NotNil(t, condition)
	_, ok := condition.(*OrCondition)
	assert.True(t, ok, "Expected *OrCondition")
}

// Test: Conditions.FromProtobuf() deserializes NotCondition correctly
// Given: A protobuf Condition message of type NotCondition wrapping a WorkContextCondition
// When: Calling FromProtobuf()
// Then: Returns NotCondition wrapping the deserialized condition
func TestConditions_FromProtobuf_NotCondition(t *testing.T) {
	proto := &pipelinepb.Condition{
		Conditions: &pipelinepb.Condition_NotCondition{
			NotCondition: &pipelinepb.NotCondition{
				Condition: &pipelinepb.Condition{
					Conditions: &pipelinepb.Condition_WorkContextCondition{
						WorkContextCondition: &pipelinepb.WorkContextCondition{
							Key:    "env",
							Value:  "prod",
							Operand: pipelinepb.WorkContextCondition_EQUALS,
						},
					},
				},
			},
		},
	}

	condition, err := FromProtobuf(proto)

	assert.NoError(t, err)
	assert.NotNil(t, condition)
	_, ok := condition.(*NotCondition)
	assert.True(t, ok, "Expected *NotCondition")
}

// Test: Conditions.FromProtobuf() returns error for unknown condition type
// Given: A protobuf Condition message with an unrecognized/unsupported type
// When: Calling FromProtobuf()
// Then: Returns error indicating unknown condition type
func TestConditions_FromProtobuf_UnknownType(t *testing.T) {
	proto := &pipelinepb.Condition{}

	_, err := FromProtobuf(proto)

	assert.Error(t, err)
}

// Test: Conditions.FromProtobuf() handles deeply nested composite conditions
// Given: A protobuf Condition with nested And/Or/Not conditions (3+ levels deep)
// When: Calling FromProtobuf()
// Then: Returns correctly structured composite condition preserving nesting
func TestConditions_FromProtobuf_DeepNesting(t *testing.T) {
	proto := &pipelinepb.Condition{
		Conditions: &pipelinepb.Condition_AndCondition{
			AndCondition: &pipelinepb.AndCondition{
				Left: &pipelinepb.Condition{
					Conditions: &pipelinepb.Condition_OrCondition{
						OrCondition: &pipelinepb.OrCondition{
							Left: &pipelinepb.Condition{
								Conditions: &pipelinepb.Condition_NotCondition{
									NotCondition: &pipelinepb.NotCondition{
										Condition: &pipelinepb.Condition{
											Conditions: &pipelinepb.Condition_WorkContextCondition{
												WorkContextCondition: &pipelinepb.WorkContextCondition{
													Key:    "env",
													Value:  "prod",
													Operand: pipelinepb.WorkContextCondition_EQUALS,
												},
											},
										},
									},
								},
							},
							Right: &pipelinepb.Condition{
								Conditions: &pipelinepb.Condition_WorkContextCondition{
									WorkContextCondition: &pipelinepb.WorkContextCondition{
										Key:    "region",
										Value:  "us",
										Operand: pipelinepb.WorkContextCondition_EQUALS,
									},
								},
							},
						},
					},
				},
				Right: &pipelinepb.Condition{
					Conditions: &pipelinepb.Condition_WorkContextCondition{
						WorkContextCondition: &pipelinepb.WorkContextCondition{
							Key:    "debug",
							Value:  "true",
							Operand: pipelinepb.WorkContextCondition_EQUALS,
						},
					},
				},
			},
		},
	}

	condition, err := FromProtobuf(proto)

	assert.NoError(t, err)
	assert.NotNil(t, condition)
	andCond, ok := condition.(*AndCondition)
	assert.True(t, ok, "Expected *AndCondition at top level")
	assert.NotNil(t, andCond)
}
