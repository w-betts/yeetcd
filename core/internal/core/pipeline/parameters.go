package pipeline

// Parameters wraps a map of parameter definitions
type Parameters map[string]Parameter

// EmptyParameters creates empty Parameters
func EmptyParameters() Parameters {
	return make(Parameters)
}

// ParametersFromMap creates Parameters from a map
func ParametersFromMap(m map[string]Parameter) Parameters {
	return Parameters(m)
}
