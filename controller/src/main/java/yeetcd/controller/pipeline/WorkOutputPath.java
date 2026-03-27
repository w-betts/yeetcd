package yeetcd.controller.pipeline;

import yeetcd.protocol.pipeline.PipelineOuterClass;

public record WorkOutputPath(String name, String path) {

    public static WorkOutputPath fromProtobuf(PipelineOuterClass.WorkOutputPath workOutputPath) {
        return new WorkOutputPath(workOutputPath.getName(), workOutputPath.getPath());
    }
}
