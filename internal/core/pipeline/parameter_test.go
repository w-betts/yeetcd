package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParameter_ValidateArgument_STRING validates STRING type
func TestParameter_ValidateArgument_STRING(t *testing.T) {
	param := &Parameter{
		TypeCheck: STRING,
	}

	err := param.ValidateArgument("any-string")
	require.NoError(t, err)
}

// TestParameter_ValidateArgument_STRING_WithChoices validates STRING with choices
func TestParameter_ValidateArgument_STRING_WithChoices(t *testing.T) {
	param := &Parameter{
		TypeCheck: STRING,
		Choices:   []string{"choice1", "choice2"},
	}

	err := param.ValidateArgument("choice1")
	require.NoError(t, err)
}

// TestParameter_ValidateArgument_STRING_WithInvalidChoice fails for invalid choice
func TestParameter_ValidateArgument_STRING_WithInvalidChoice(t *testing.T) {
	param := &Parameter{
		TypeCheck: STRING,
		Choices:   []string{"choice1", "choice2"},
	}

	err := param.ValidateArgument("invalid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "choice")
}

// TestParameter_ValidateArgument_NUMBER_Valid validates valid number
func TestParameter_ValidateArgument_NUMBER_Valid(t *testing.T) {
	param := &Parameter{
		TypeCheck: NUMBER,
	}

	err := param.ValidateArgument("123")
	require.NoError(t, err)
}

// TestParameter_ValidateArgument_NUMBER_Invalid fails for non-numeric
func TestParameter_ValidateArgument_NUMBER_Invalid(t *testing.T) {
	param := &Parameter{
		TypeCheck: NUMBER,
	}

	err := param.ValidateArgument("not-a-number")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "number")
}

// TestParameter_ValidateArgument_BOOLEAN_Valid validates valid boolean
func TestParameter_ValidateArgument_BOOLEAN_Valid(t *testing.T) {
	param := &Parameter{
		TypeCheck: BOOLEAN,
	}

	err := param.ValidateArgument("true")
	require.NoError(t, err)
}

// TestParameter_ValidateArgument_BOOLEAN_Invalid fails for non-boolean
func TestParameter_ValidateArgument_BOOLEAN_Invalid(t *testing.T) {
	param := &Parameter{
		TypeCheck: BOOLEAN,
	}

	err := param.ValidateArgument("not-a-boolean")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boolean")
}
