package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestArguments_ValidationPassesForValidStringParameter validates STRING parameter with matching choice
// GIVEN: Parameter with TypeCheck STRING and choices ['value1', 'value2']
// WHEN: AsValidatedWorkContext() is called with argument 'value1'
// THEN: Validation passes and WorkContext contains the argument
func TestArguments_ValidationPassesForValidStringParameter(t *testing.T) {
	params := Parameters{
		"param1": {
			TypeCheck: STRING,
			Choices:   []string{"value1", "value2"},
		},
	}

	args := ArgumentsOf("param1", "value1")

	ctx, err := args.AsValidatedWorkContext(params)

	require.NoError(t, err)
	assert.Equal(t, "value1", ctx["param1"])
}

// TestArguments_ValidationFailsForInvalidNumberParameter validates NUMBER parameter with non-numeric value
// GIVEN: Parameter with TypeCheck NUMBER
// WHEN: AsValidatedWorkContext() is called with argument 'not-a-number'
// THEN: Validation fails with error indicating invalid number format
func TestArguments_ValidationFailsForInvalidNumberParameter(t *testing.T) {
	params := Parameters{
		"param1": {
			TypeCheck: NUMBER,
		},
	}

	args := ArgumentsOf("param1", "not-a-number")

	_, err := args.AsValidatedWorkContext(params)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "number")
}

// TestArguments_ValidationFailsForInvalidBooleanParameter validates BOOLEAN parameter with non-boolean value
// GIVEN: Parameter with TypeCheck BOOLEAN
// WHEN: AsValidatedWorkContext() is called with argument 'not-a-boolean'
// THEN: Validation fails with error indicating invalid boolean value
func TestArguments_ValidationFailsForInvalidBooleanParameter(t *testing.T) {
	params := Parameters{
		"param1": {
			TypeCheck: BOOLEAN,
		},
	}

	args := ArgumentsOf("param1", "not-a-boolean")

	_, err := args.AsValidatedWorkContext(params)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "boolean")
}

// TestArguments_ValidationFailsWhenArgumentNotInAllowedChoices validates choices constraint
// GIVEN: Parameter with choices ['allowed1', 'allowed2']
// WHEN: AsValidatedWorkContext() is called with argument 'not-allowed'
// THEN: Validation fails with error indicating argument not in allowed choices
func TestArguments_ValidationFailsWhenArgumentNotInAllowedChoices(t *testing.T) {
	params := Parameters{
		"param1": {
			TypeCheck: STRING,
			Choices:   []string{"allowed1", "allowed2"},
		},
	}

	args := ArgumentsOf("param1", "not-allowed")

	_, err := args.AsValidatedWorkContext(params)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "choice")
}

// TestArguments_ValidationFailsWhenRequiredParameterIsMissing validates required parameter constraint
// GIVEN: Parameter with required=true
// WHEN: AsValidatedWorkContext() is called without providing that argument
// THEN: Validation fails with error indicating required parameter is missing
func TestArguments_ValidationFailsWhenRequiredParameterIsMissing(t *testing.T) {
	params := Parameters{
		"param1": {
			TypeCheck: STRING,
			Required:  true,
		},
	}

	args := Arguments{} // Empty arguments

	_, err := args.AsValidatedWorkContext(params)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}
