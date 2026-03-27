package yeetcd.test;

import yeetcd.sdk.Pipeline;
import lombok.EqualsAndHashCode;
import lombok.ToString;

import java.util.Collections;
import java.util.Map;

@EqualsAndHashCode
@ToString
public final class FakePipelineRun {

    private final Pipeline pipeline;

    private final Map<String, String> arguments;

    private FakePipelineRun(Pipeline pipeline, Map<String, String> arguments) {
        this.pipeline = pipeline;
        this.arguments = arguments;
    }

    public Pipeline getPipeline() {
        return pipeline;
    }

    public Map<String, String> getArguments() {
        return arguments;
    }

    public static Builder builder(Pipeline pipeline) {
        return new Builder(pipeline);
    }

    public static class Builder {
        private final Pipeline pipeline;

        private Map<String, String> arguments = Collections.emptyMap();

        private Builder(Pipeline pipeline) {
            this.pipeline = pipeline;
        }

        public Builder arguments(Map<String, String> arguments) {
            this.arguments = arguments;
            return this;
        }


        public FakePipelineRun build() {
            return new FakePipelineRun(pipeline, arguments);
        }
    }
}
