package yeetcd.sdk;

import yeetcd.protocol.pipeline.PipelineOuterClass;
import lombok.EqualsAndHashCode;
import lombok.ToString;

import java.util.Arrays;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.function.Function;
import java.util.stream.Collectors;
import java.util.stream.Stream;

@EqualsAndHashCode
@ToString
public final class Pipeline {
    private final String name;
    private final Parameters parameters;
    private final WorkContext workContext;
    private final Work[] finalWork;
    private static final ConcurrentHashMap<String, Map<String, NativeWorkDefinition>> nativeWorkDefinitions = new ConcurrentHashMap<>();

    private Pipeline(String name, Parameters parameters, WorkContext workContext, Work[] finalWork) {
        this.name = name;
        this.parameters = parameters;
        this.workContext = workContext;
        this.finalWork = finalWork;
        nativeWorkDefinitions.computeIfAbsent(name, pipelineName -> nativeWorkDefinitions().collect(Collectors.toMap(
            NativeWorkDefinition::executionId,
            Function.identity(),
            (merge1, merge2) -> merge1)
        ));
    }

    public Parameters getParameters() {
        return parameters;
    }

    public WorkContext getWorkContext() {
        return workContext;
    }

    public Work[] getFinalWork() {
        return finalWork;
    }

    public Work.Builder asWorkBuilder(String description) {
        return Work.builder(description, CompoundWorkDefinition.builder(finalWork).build()).workContext(workContext);
    }

    @SuppressWarnings("unused") // This is used by generated code
    public static void runNativeWorkDefinition(String pipelineName, String taskName) {
        Map<String, NativeWorkDefinition> pipelineWorkDefinitions = nativeWorkDefinitions.get(pipelineName);
        if (pipelineWorkDefinitions == null) {
            throw new IllegalArgumentException("Pipeline %s not found in native work map".formatted(pipelineName));
        }
        NativeWorkDefinition nativeWorkDefinition = pipelineWorkDefinitions.get(taskName);
        if (nativeWorkDefinition == null) {
            throw new IllegalArgumentException("Work definition %s not found in native work map. Keys are [%s]".formatted(taskName, String.join(", ", pipelineWorkDefinitions.keySet())));
        }
        nativeWorkDefinition.run();
    }

    public static Builder builder(String name) {
        return new Builder(name);
    }

    public static class Builder {

        private final String name;
        private Parameters parameters = Parameters.empty();
        private WorkContext workContext = WorkContext.empty();
        private Work[] finalWork = new Work[]{};

        public Builder(String name) {
            this.name = name;
        }

        public Builder parameters(Parameters parameters) {
            this.parameters = parameters;
            return this;
        }

        public Builder workContext(WorkContext workContext) {
            this.workContext = workContext;
            return this;
        }

        public Builder finalWork(Work... finalWork) {
            this.finalWork = finalWork;
            return this;
        }

        public Pipeline build() {
            return new Pipeline(name, parameters, workContext, finalWork);
        }
    }

    public PipelineOuterClass.Pipeline toProtobuf() {
        return PipelineOuterClass.Pipeline
            .newBuilder()
            .setName(name)
            .putAllWorkContext(workContext.getWorkContextMap())
            .putAllParameters(parameters
                .getParametersMap()
                .entrySet().stream()
                .collect(Collectors.toMap(
                    Map.Entry::getKey,
                    entry -> entry.getValue().toProtobuf()
                ))
            )
            .addAllFinalWork(Arrays
                .stream(finalWork)
                .map(work -> work.toProtobuf(workContext))
                .toList()
            )
            .build();
    }


    private Stream<NativeWorkDefinition> nativeWorkDefinitions() {
        return Arrays
            .stream(finalWork)
            .flatMap(Work::nativeWorkDefinitions);
    }
}
