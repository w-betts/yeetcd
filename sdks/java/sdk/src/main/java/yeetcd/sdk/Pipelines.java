package yeetcd.sdk;

import yeetcd.protocol.pipeline.PipelineOuterClass;
import lombok.EqualsAndHashCode;
import lombok.ToString;

import java.util.Arrays;

@EqualsAndHashCode
@ToString
final class Pipelines {
    private final Pipeline[] pipelines;

    Pipelines(Pipeline... pipelines) {
        this.pipelines = pipelines;
    }

    PipelineOuterClass.Pipelines toProtobuf() {
        return PipelineOuterClass.Pipelines
                .newBuilder()
                .addAllPipelines(Arrays
                        .stream(pipelines)
                        .map(Pipeline::toProtobuf)
                        .toList()
                )
                .build();
    }
}
