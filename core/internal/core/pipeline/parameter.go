package pipeline

import (
	"fmt"
	"strconv"
	"strings"
)

// TypeCheck represents parameter type checking
type TypeCheck int

const (
	STRING TypeCheck = iota
	NUMBER
	BOOLEAN
)

// Parameter represents a pipeline parameter
type Parameter struct {
	TypeCheck    TypeCheck
	Required     bool
	DefaultValue string
	Choices      []string
}

// ValidateArgument validates an argument against this parameter
func (p *Parameter) ValidateArgument(value string) error {
	// Check choices constraint first
	if len(p.Choices) > 0 {
		found := false
		for _, choice := range p.Choices {
			if choice == value {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("value '%s' is not in allowed choices: %v", value, p.Choices)
		}
	}

	// Validate based on type
	switch p.TypeCheck {
	case STRING:
		// Any string is valid
		return nil
	case NUMBER:
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("invalid number format: '%s'", value)
		}
		return nil
	case BOOLEAN:
		lower := strings.ToLower(value)
		if lower != "true" && lower != "false" {
			return fmt.Errorf("invalid boolean value: '%s' (must be 'true' or 'false')", value)
		}
		return nil
	default:
		return fmt.Errorf("unknown type check: %v", p.TypeCheck)
	}
}

// FromProtobuf converts protobuf Parameter to Go struct
func ParameterFromProtobuf(protoParameter interface{}) (*Parameter, error) {
	return nil, nil
}
