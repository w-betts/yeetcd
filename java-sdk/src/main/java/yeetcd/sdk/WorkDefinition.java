package yeetcd.sdk;

import yeetcd.protocol.pipeline.PipelineOuterClass;

import java.util.stream.Stream;

public interface WorkDefinition {
    void applyTo(WorkContext workContext, PipelineOuterClass.Work.Builder taskBuilder);
    Stream<NativeWorkDefinition> nativeWorkDefinitions();
}
