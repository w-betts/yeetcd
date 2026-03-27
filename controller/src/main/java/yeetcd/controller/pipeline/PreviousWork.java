package yeetcd.controller.pipeline;

import yeetcd.protocol.pipeline.PipelineOuterClass;

public record PreviousWork(Work work, String outputPathsMount, String stdOutEnvVar) {

    public static PreviousWork fromProtobuf(PipelineOuterClass.PreviousWork previousWork) {
        return new PreviousWork(
                Work.fromProtobuf(previousWork.getWork()),
                previousWork.getOutputPathsMount(),
                previousWork.getStdOutEnvVar()
        );
    }
}
