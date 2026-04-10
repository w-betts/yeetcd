package pipeline

import (
	"fmt"
)

// Arguments represents pipeline arguments
type Arguments map[string]string

// AsValidatedWorkContext validates arguments against parameters and returns WorkContext
// Arguments can contain:
// 1. Parameter values (validated against parameter definitions)
// 2. Work context overrides (keys that exist in the work context)
func (a Arguments) AsValidatedWorkContext(params Parameters, existingContext ...WorkContext) (WorkContext, error) {
	result := make(WorkContext)
	
	// Get the existing context if provided
	var workContext WorkContext
	if len(existingContext) > 0 {
		workContext = existingContext[0]
	}

	// Check for missing required parameters
	for name, param := range params {
		if param.Required {
			if _, ok := a[name]; !ok {
				return nil, fmt.Errorf("required parameter '%s' is missing", name)
			}
		}
	}

	// Validate each argument
	for name, value := range a {
		if param, ok := params[name]; ok {
			// This is a parameter - validate it
			if err := param.ValidateArgument(value); err != nil {
				return nil, fmt.Errorf("validation failed for parameter '%s': %w", name, err)
			}
			result[name] = value
		} else if workContext != nil {
			// Check if this is a work context override
			if _, exists := workContext[name]; exists {
				result[name] = value
			} else {
				return nil, fmt.Errorf("unknown parameter: '%s'", name)
			}
		} else {
			return nil, fmt.Errorf("unknown parameter: '%s'", name)
		}
	}

	// Apply default values for missing optional parameters
	for name, param := range params {
		if _, ok := result[name]; !ok && param.DefaultValue != "" {
			result[name] = param.DefaultValue
		}
	}

	return result, nil
}

// ArgumentsOf creates Arguments from key-value pairs
func ArgumentsOf(pairs ...string) Arguments {
	args := make(Arguments)
	for i := 0; i < len(pairs)-1; i += 2 {
		args[pairs[i]] = pairs[i+1]
	}
	return args
}
