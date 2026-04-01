package pipeline

// WorkOutputPath represents an output path for work
type WorkOutputPath struct {
	Name string
	Path string
}

// FromProtobuf converts protobuf WorkOutputPath to Go struct
func WorkOutputPathFromProtobuf(protoWorkOutputPath interface{}) (*WorkOutputPath, error) {
	return nil, nil
}
