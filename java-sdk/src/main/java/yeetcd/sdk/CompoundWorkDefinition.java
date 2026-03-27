package yeetcd.sdk;

import yeetcd.protocol.pipeline.PipelineOuterClass;
import lombok.EqualsAndHashCode;
import lombok.ToString;

import java.util.Arrays;
import java.util.stream.Collectors;
import java.util.stream.Stream;

@EqualsAndHashCode
@ToString
public final class CompoundWorkDefinition implements WorkDefinition {
    private final Work[] finalWork;

    private CompoundWorkDefinition(Work[] finalWork) {
        this.finalWork = finalWork;
    }

    public Work[] getFinalWork() {
        return finalWork;
    }

    public static Builder builder(Work... finalWork) {
        return new Builder(finalWork);
    }

    public static class Builder {
        private final Work[] finalWork;

        public Builder(Work[] finalWork) {
            this.finalWork = finalWork;
        }

        public CompoundWorkDefinition build() {
            return new CompoundWorkDefinition(finalWork);
        }
    }

    @Override
    public void applyTo(WorkContext context, PipelineOuterClass.Work.Builder workBuilder) {
        workBuilder.setCompoundWorkDefinition(
            PipelineOuterClass.CompoundWorkDefinition
                .newBuilder()
                .addAllFinalWork(Arrays.stream(finalWork).map(work -> work.toProtobuf(context)).collect(Collectors.toList()))
                .build()
        );
    }

    @Override
    public Stream<NativeWorkDefinition> nativeWorkDefinitions() {
        return Arrays
            .stream(finalWork)
            .flatMap(work -> Stream
                .concat(
                    work.nativeWorkDefinitions(),
                    Arrays
                        .stream(work.getPreviousWork())
                        .flatMap(workWithOutputsMountPath -> workWithOutputsMountPath.getWork().nativeWorkDefinitions())
                )
            );
    }
}
