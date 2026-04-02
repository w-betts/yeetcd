package yeetcd.sdk;

import yeetcd.protocol.pipeline.PipelineOuterClass;
import lombok.EqualsAndHashCode;
import lombok.ToString;

import java.util.Arrays;
import java.util.stream.Stream;

@EqualsAndHashCode
@ToString
public final class ContainerisedWorkDefinition implements WorkDefinition {
    private final String image;
    private final String[] command;

    private ContainerisedWorkDefinition(String image, String... command) {
        this.image = image;
        this.command = command;
    }

    public static Builder builder(String image) {
        return new Builder(image);
    }

    public static class Builder {
        private final String image;
        private String[] command = new String[]{};

        public Builder(String image) {
            this.image = image;
        }

        public Builder command(String... command) {
            this.command = command;
            return this;
        }

        public ContainerisedWorkDefinition build() {
            return new ContainerisedWorkDefinition(image, command);
        }
    }

    @Override
    public void applyTo(WorkContext containingContext, PipelineOuterClass.Work.Builder workBuilder) {
        workBuilder.setContainerisedWorkDefinition(
            PipelineOuterClass.ContainerisedWorkDefinition
                .newBuilder()
                .setImage(image)
                .addAllCmd(Arrays.asList(command))
                .build()
        );
    }

    @Override
    public Stream<NativeWorkDefinition> nativeWorkDefinitions() {
        return Stream.empty();
    }
}
