package engine

// BuildImageDefinition defines how to build an image
type BuildImageDefinition struct {
	Image           string
	Tag             string
	ImageBase       ImageBase
	ArtifactDirectory string
	ArtifactNames   []string
	Cmd             string
}
