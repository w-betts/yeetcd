package yeetcd.sdk;

import yeetcd.protocol.pipeline.PipelineOuterClass;
import lombok.SneakyThrows;

import java.util.stream.Stream;

public abstract class DynamicWorkGeneratingWorkDefinition extends NativeWorkDefinition {

    @Override
    public final void applyTo(WorkContext containingContext, PipelineOuterClass.Work.Builder workBuilder) {
        workBuilder.setDynamicWorkGeneratingWorkDefinition(
            PipelineOuterClass.DynamicWorkGeneratingWorkDefinition
                .newBuilder()
                .setExecutionId(executionId())
                .build()
        );
    }

    @Override
    @SneakyThrows
    public void run() {
        createWork().toProtobuf(WorkContext.empty()).writeTo(System.out);
        System.out.flush();
    }

    @Override
    public Stream<NativeWorkDefinition> nativeWorkDefinitions() {
        return Stream.concat(super.nativeWorkDefinitions(), dynamicCustomWorkDefinitions());
    }

    public abstract Work createWork();

    protected abstract Stream<CustomWorkDefinition> dynamicCustomWorkDefinitions();
}
