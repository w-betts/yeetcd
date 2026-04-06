package yeetcd.sdk;

import yeetcd.protocol.pipeline.PipelineOuterClass;

public abstract class CustomWorkDefinition extends NativeWorkDefinition {
    @Override
    public final void applyTo(WorkContext containingContext, PipelineOuterClass.Work.Builder workBuilder) {
        workBuilder.setCustomWorkDefinition(
                PipelineOuterClass.CustomWorkDefinition
                        .newBuilder()
                        .setExecutionId(executionId())
                        .build()
        );
    }

    public final String getExecutionId() {
        return executionId();
    }
}
