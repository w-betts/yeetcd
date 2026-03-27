package yeetcd.sdk.condition;

import yeetcd.protocol.pipeline.PipelineOuterClass;
import lombok.EqualsAndHashCode;

@EqualsAndHashCode(callSuper = true)
public final class NotCondition extends Condition {
    private final Condition condition;

    private NotCondition(Condition condition) {
        this.condition = condition;
    }

    public Condition getCondition() {
        return condition;
    }

    @Override
    PipelineOuterClass.Condition toProtobuf() {
        PipelineOuterClass.NotCondition.Builder builder = PipelineOuterClass.NotCondition.newBuilder();
        condition.applyTo(builder);
        return builder.getCondition();
    }

    public static Builder builder(Condition condition) {
        return new Builder(condition);
    }

    public static class Builder {
        private final Condition condition;

        public Builder(Condition condition) {
            this.condition = condition;
        }

        public NotCondition build() {
            return new NotCondition(condition);
        }
    }
}
