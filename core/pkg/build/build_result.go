package build

import pb "github.com/yeetcd/yeetcd/pkg/proto/pipeline"

// BuildResult is the result of building source
type BuildResult struct {
	ImageID            string
	Pipelines          []*pb.Pipeline // protobuf Pipeline messages
	SourceBuildResults []SourceBuildResult
}
