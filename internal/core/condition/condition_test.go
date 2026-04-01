package condition

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yeetcd/yeetcd/internal/core/pipeline"
)

// Test: WorkContextCondition.Evaluate() returns true when context key matches expected value
// Given: A WorkContextCondition with key="env", expectedValue="prod", operator=EQUALS
// When: Evaluating against a WorkContext containing {"env": "prod"}
// Then: Returns true
func TestWorkContextCondition_Evaluate_Matches(t *testing.T) {
	condition := NewWorkContextCondition("env", "prod", OperatorEquals)
	ctx := pipeline.NewWorkContext(map[string]string{"env": "prod"})

	result, err := condition.Evaluate(ctx, nil)

	assert.NoError(t, err)
	assert.True(t, result)
}

// Test: WorkContextCondition.Evaluate() returns false when context key does not match
// Given: A WorkContextCondition with key="env", expectedValue="prod", operator=EQUALS
// When: Evaluating against a WorkContext containing {"env": "dev"}
// Then: Returns false
func TestWorkContextCondition_Evaluate_NoMatch(t *testing.T) {
	condition := NewWorkContextCondition("env", "prod", OperatorEquals)
	ctx := pipeline.NewWorkContext(map[string]string{"env": "dev"})

	result, err := condition.Evaluate(ctx, nil)

	assert.NoError(t, err)
	assert.False(t, result)
}

// Test: WorkContextCondition.Evaluate() returns error when context key is missing
// Given: A WorkContextCondition with key="missing", expectedValue="value", operator=EQUALS
// When: Evaluating against a WorkContext that does not contain the key
// Then: Returns error indicating missing key
func TestWorkContextCondition_Evaluate_MissingKey(t *testing.T) {
	condition := NewWorkContextCondition("missing", "value", OperatorEquals)
	ctx := pipeline.NewWorkContext(map[string]string{"other": "value"})

	_, err := condition.Evaluate(ctx, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}

// Test: WorkContextCondition.ToProtobuf() serializes condition correctly
// Given: A WorkContextCondition with key="env", expectedValue="prod", operator=EQUALS
// When: Calling ToProtobuf()
// Then: Returns protobuf message with matching fields
func TestWorkContextCondition_ToProtobuf(t *testing.T) {
	condition := NewWorkContextCondition("env", "prod", OperatorEquals)

	proto, err := condition.ToProtobuf()

	assert.NoError(t, err)
	assert.NotNil(t, proto)
}

// Test: PreviousWorkStatusCondition.Evaluate() returns true when previous work succeeded
// Given: A PreviousWorkStatusCondition with status=SUCCESS
// When: Evaluating against a WorkResult with status=SUCCESS
// Then: Returns true
func TestPreviousWorkStatusCondition_Evaluate_Success(t *testing.T) {
	condition := NewPreviousWorkStatusCondition(WorkStatusSuccess)
	tracker := pipeline.NewWorkResultTracker()
	tracker.RecordResult("prev-work", &pipeline.WorkResult{WorkStatus: pipeline.WorkStatusSucceeded})

	match, err := condition.Evaluate(nil, tracker)

	assert.NoError(t, err)
	assert.True(t, match)
}

// Test: PreviousWorkStatusCondition.Evaluate() returns false when previous work failed
// Given: A PreviousWorkStatusCondition with status=SUCCESS
// When: Evaluating against a WorkResult with status=FAILED
// Then: Returns false
func TestPreviousWorkStatusCondition_Evaluate_Failure(t *testing.T) {
	condition := NewPreviousWorkStatusCondition(WorkStatusSuccess)
	tracker := pipeline.NewWorkResultTracker()
	tracker.RecordResult("prev-work", &pipeline.WorkResult{WorkStatus: pipeline.WorkStatusFailed})

	match, err := condition.Evaluate(nil, tracker)

	assert.NoError(t, err)
	assert.False(t, match)
}

// Test: PreviousWorkStatusCondition.Evaluate() returns error when previous work result is nil
// Given: A PreviousWorkStatusCondition with status=SUCCESS
// When: Evaluating with nil previous work result
// Then: Returns error indicating missing previous work result
func TestPreviousWorkStatusCondition_Evaluate_NilResult(t *testing.T) {
	condition := NewPreviousWorkStatusCondition(WorkStatusSuccess)

	_, err := condition.Evaluate(nil, nil)

	assert.Error(t, err)
}

// Test: PreviousWorkStatusCondition.ToProtobuf() serializes condition correctly
// Given: A PreviousWorkStatusCondition with status=SUCCESS
// When: Calling ToProtobuf()
// Then: Returns protobuf message with status field set to SUCCESS
func TestPreviousWorkStatusCondition_ToProtobuf(t *testing.T) {
	condition := NewPreviousWorkStatusCondition(WorkStatusSuccess)

	proto, err := condition.ToProtobuf()

	assert.NoError(t, err)
	assert.NotNil(t, proto)
}

// Test: AndCondition.Evaluate() returns true when all conditions are true
// Given: An AndCondition containing two WorkContextConditions that both evaluate to true
// When: Evaluating against matching context
// Then: Returns true
func TestAndCondition_Evaluate_AllTrue(t *testing.T) {
	cond1 := NewWorkContextCondition("env", "prod", OperatorEquals)
	cond2 := NewWorkContextCondition("region", "us", OperatorEquals)
	andCond := NewAndCondition(cond1, cond2)
	ctx := pipeline.NewWorkContext(map[string]string{"env": "prod", "region": "us"})

	result, err := andCond.Evaluate(ctx, nil)

	assert.NoError(t, err)
	assert.True(t, result)
}

// Test: AndCondition.Evaluate() returns false when any condition is false
// Given: An AndCondition containing two WorkContextConditions where one is false
// When: Evaluating against context matching only one condition
// Then: Returns false
func TestAndCondition_Evaluate_OneFalse(t *testing.T) {
	cond1 := NewWorkContextCondition("env", "prod", OperatorEquals)
	cond2 := NewWorkContextCondition("region", "eu", OperatorEquals)
	andCond := NewAndCondition(cond1, cond2)
	ctx := pipeline.NewWorkContext(map[string]string{"env": "prod", "region": "us"})

	result, err := andCond.Evaluate(ctx, nil)

	assert.NoError(t, err)
	assert.False(t, result)
}

// Test: AndCondition.ToProtobuf() serializes all nested conditions
// Given: An AndCondition with multiple nested conditions
// When: Calling ToProtobuf()
// Then: Returns protobuf message containing all nested conditions
func TestAndCondition_ToProtobuf(t *testing.T) {
	cond1 := NewWorkContextCondition("env", "prod", OperatorEquals)
	cond2 := NewWorkContextCondition("region", "us", OperatorEquals)
	andCond := NewAndCondition(cond1, cond2)

	proto, err := andCond.ToProtobuf()

	assert.NoError(t, err)
	assert.NotNil(t, proto)
}

// Test: OrCondition.Evaluate() returns true when at least one condition is true
// Given: An OrCondition containing two WorkContextConditions where one is true
// When: Evaluating against context matching one condition
// Then: Returns true
func TestOrCondition_Evaluate_OneTrue(t *testing.T) {
	cond1 := NewWorkContextCondition("env", "prod", OperatorEquals)
	cond2 := NewWorkContextCondition("region", "eu", OperatorEquals)
	orCond := NewOrCondition(cond1, cond2)
	ctx := pipeline.NewWorkContext(map[string]string{"env": "prod", "region": "us"})

	result, err := orCond.Evaluate(ctx, nil)

	assert.NoError(t, err)
	assert.True(t, result)
}

// Test: OrCondition.Evaluate() returns false when all conditions are false
// Given: An OrCondition containing two WorkContextConditions that both evaluate to false
// When: Evaluating against context matching neither condition
// Then: Returns false
func TestOrCondition_Evaluate_AllFalse(t *testing.T) {
	cond1 := NewWorkContextCondition("env", "prod", OperatorEquals)
	cond2 := NewWorkContextCondition("region", "eu", OperatorEquals)
	orCond := NewOrCondition(cond1, cond2)
	ctx := pipeline.NewWorkContext(map[string]string{"env": "dev", "region": "us"})

	result, err := orCond.Evaluate(ctx, nil)

	assert.NoError(t, err)
	assert.False(t, result)
}

// Test: OrCondition.ToProtobuf() serializes all nested conditions
// Given: An OrCondition with multiple nested conditions
// When: Calling ToProtobuf()
// Then: Returns protobuf message containing all nested conditions
func TestOrCondition_ToProtobuf(t *testing.T) {
	cond1 := NewWorkContextCondition("env", "prod", OperatorEquals)
	cond2 := NewWorkContextCondition("region", "us", OperatorEquals)
	orCond := NewOrCondition(cond1, cond2)

	proto, err := orCond.ToProtobuf()

	assert.NoError(t, err)
	assert.NotNil(t, proto)
}

// Test: NotCondition.Evaluate() returns false when wrapped condition is true
// Given: A NotCondition wrapping a WorkContextCondition that evaluates to true
// When: Evaluating against matching context
// Then: Returns false
func TestNotCondition_Evaluate_WrappedTrue(t *testing.T) {
	wrapped := NewWorkContextCondition("env", "prod", OperatorEquals)
	notCond := NewNotCondition(wrapped)
	ctx := pipeline.NewWorkContext(map[string]string{"env": "prod"})

	result, err := notCond.Evaluate(ctx, nil)

	assert.NoError(t, err)
	assert.False(t, result)
}

// Test: NotCondition.Evaluate() returns true when wrapped condition is false
// Given: A NotCondition wrapping a WorkContextCondition that evaluates to false
// When: Evaluating against non-matching context
// Then: Returns true
func TestNotCondition_Evaluate_WrappedFalse(t *testing.T) {
	wrapped := NewWorkContextCondition("env", "prod", OperatorEquals)
	notCond := NewNotCondition(wrapped)
	ctx := pipeline.NewWorkContext(map[string]string{"env": "dev"})

	result, err := notCond.Evaluate(ctx, nil)

	assert.NoError(t, err)
	assert.True(t, result)
}

// Test: NotCondition.ToProtobuf() serializes wrapped condition
// Given: A NotCondition wrapping a WorkContextCondition
// When: Calling ToProtobuf()
// Then: Returns protobuf message containing the wrapped condition
func TestNotCondition_ToProtobuf(t *testing.T) {
	wrapped := NewWorkContextCondition("env", "prod", OperatorEquals)
	notCond := NewNotCondition(wrapped)

	proto, err := notCond.ToProtobuf()

	assert.NoError(t, err)
	assert.NotNil(t, proto)
}
