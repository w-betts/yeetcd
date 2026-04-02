package yeetcd.sdk.condition;

import yeetcd.protocol.pipeline.PipelineOuterClass;
import lombok.EqualsAndHashCode;

@EqualsAndHashCode(callSuper = true)
public final class WorkContextCondition extends Condition {
    private final String key;

    private final Operand operand;

    private final String value;
    private WorkContextCondition(String key, Operand operand, String value) {
        this.key = key;
        this.operand = operand;
        this.value = value;
    }

    public String getKey() {
        return key;
    }

    public Operand getOperand() {
        return operand;
    }

    public String getValue() {
        return value;
    }

    public static Builder builder(String key, Operand operand, String value) {
        return new Builder(key, operand, value);
    }

    @Override
    PipelineOuterClass.Condition toProtobuf() {
        return PipelineOuterClass.Condition.newBuilder()
            .setWorkContextCondition(PipelineOuterClass.WorkContextCondition.newBuilder()
                .setKey(key)
                .setOperand(operand.toProtobuf())
                .setValue(value)
                .build()
            )
            .build();
    }

    public static class Builder {
        private final String key;
        private final Operand operand;
        private final String value;

        public Builder(String key, Operand operand, String value) {
            this.key = key;
            this.operand = operand;
            this.value = value;
        }

        public WorkContextCondition build() {
            return new WorkContextCondition(key, operand, value);
        }
    }

    public enum Operand {
        EQUALS {
            @Override
            PipelineOuterClass.WorkContextCondition.Operand toProtobuf() {
                return PipelineOuterClass.WorkContextCondition.Operand.EQUALS;
            }
        };

        abstract PipelineOuterClass.WorkContextCondition.Operand toProtobuf();

    }
}
