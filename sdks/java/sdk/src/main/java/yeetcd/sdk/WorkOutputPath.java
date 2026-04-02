package yeetcd.sdk;

import yeetcd.protocol.pipeline.PipelineOuterClass;
import lombok.EqualsAndHashCode;
import lombok.ToString;

@EqualsAndHashCode
@ToString
public final class WorkOutputPath {
    private final String name;
    private final String path;

    private WorkOutputPath(String name, String path) {
        this.name = name;
        this.path = path;
    }

    public String getName() {
        return name;
    }

    public String getPath() {
        return path;
    }

    PipelineOuterClass.WorkOutputPath toProtobuf() {
        return PipelineOuterClass.WorkOutputPath
            .newBuilder()
            .setName(name)
            .setPath(path)
            .build();
    }

    public static Builder builder(String name, String path) {
        return new Builder(name, path);
    }

    public static class Builder {
        private final String name;
        private final String path;


        private Builder(String name, String path) {
            this.name = name;
            this.path = path;
        }

        public WorkOutputPath build() {
            return new WorkOutputPath(name, path);
        }

    }
}
