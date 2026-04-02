package yeetcd.sdk;

import yeetcd.protocol.pipeline.PipelineOuterClass;
import lombok.EqualsAndHashCode;
import lombok.ToString;

@EqualsAndHashCode
@ToString
public final class PreviousWork {
    private final Work work;

    private final String outputPathsMount;

    private final String stdOutEnvVar;

    private PreviousWork(Work work, String outputPathsMount, String stdOutEnvVar) {
        this.work = work;
        this.outputPathsMount = outputPathsMount;
        this.stdOutEnvVar = stdOutEnvVar;
    }

    public Work getWork() {
        return work;
    }

    public String getOutputPathsMount() {
        return outputPathsMount;
    }

    public String getStdOutEnvVar() {
        return stdOutEnvVar;
    }

    public static Builder builder(Work work) {
        return new Builder(work);
    }
    public static class Builder {

        private final Work work;
        private String outputsMountPath;

        private String stdOutEnvVar;

        public Builder(Work work) {
            this.work = work;
        }

        public Builder outputsMountPath(String outputsMountPath) {
            this.outputsMountPath = outputsMountPath;
            return this;
        }

        public Builder stdOutEnvVar(String stdOutEnvVar) {
            this.stdOutEnvVar = stdOutEnvVar;
            return this;
        }
        public PreviousWork build() {
            return new PreviousWork(work, outputsMountPath, stdOutEnvVar);
        }

    }


    PipelineOuterClass.PreviousWork toProtobuf(WorkContext containingContext) {
        PipelineOuterClass.PreviousWork.Builder builder = PipelineOuterClass.PreviousWork.newBuilder().setWork(work.toProtobuf(containingContext));
        if (outputPathsMount != null) {
            builder = builder.setOutputPathsMount(outputPathsMount);
        }
        if (stdOutEnvVar != null) {
            builder = builder.setStdOutEnvVar(stdOutEnvVar);
        }
        return builder.build();
    }

}
