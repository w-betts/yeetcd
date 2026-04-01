package engine

// ImageBase represents base image types
type ImageBase int

const (
	JAVA ImageBase = iota
)

// BaseImage returns the base image name
func (i ImageBase) BaseImage() string {
	return ""
}

// EntryPoint returns the entry point for the image
func (i ImageBase) EntryPoint() string {
	return ""
}
