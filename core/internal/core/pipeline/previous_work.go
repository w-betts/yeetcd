package pipeline

// PreviousWork represents a dependency on previous work
type PreviousWork struct {
	Work             Work
	OutputPathsMount string
	StdOutEnvVar     string
}

// FromProtobuf converts protobuf PreviousWork to Go struct
func PreviousWorkFromProtobuf(protoPreviousWork interface{}) (*PreviousWork, error) {
	return nil, nil
}
