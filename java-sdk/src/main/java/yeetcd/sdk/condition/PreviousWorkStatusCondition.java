package yeetcd.sdk.condition;

import yeetcd.protocol.pipeline.PipelineOuterClass;
import lombok.EqualsAndHashCode;

@EqualsAndHashCode(callSuper = true)
public final class PreviousWorkStatusCondition extends Condition {
    private final Status status;

    private PreviousWorkStatusCondition(Status status) {
        this.status = status;
    }

    public static Builder builder(Status status) {
        return new Builder(status);
    }

    public Status getStatus() {
        return status;
    }

    @Override
    PipelineOuterClass.Condition toProtobuf() {
        return PipelineOuterClass.Condition.newBuilder()
                .setPreviousWorkStatusCondition(
                        PipelineOuterClass.PreviousWorkStatusCondition.newBuilder()
                                .setStatus(status.toProtobuf())
                                .build()
                )
                .build();
    }

    public static class Builder {
        private final Status status;

        public Builder(Status status) {
            this.status = status;
        }

        public PreviousWorkStatusCondition build() {
            return new PreviousWorkStatusCondition(status);
        }
    }

    public enum Status {
        SUCCESS {
            @Override
            PipelineOuterClass.PreviousWorkStatusCondition.Status toProtobuf() {
                return PipelineOuterClass.PreviousWorkStatusCondition.Status.SUCCESS;
            }
        },
        FAILURE {
            @Override
            PipelineOuterClass.PreviousWorkStatusCondition.Status toProtobuf() {
                return PipelineOuterClass.PreviousWorkStatusCondition.Status.FAILURE;
            }
        },
        ANY {
            @Override
            PipelineOuterClass.PreviousWorkStatusCondition.Status toProtobuf() {
                return PipelineOuterClass.PreviousWorkStatusCondition.Status.ANY;
            }
        };

        abstract PipelineOuterClass.PreviousWorkStatusCondition.Status toProtobuf();
    }
}
