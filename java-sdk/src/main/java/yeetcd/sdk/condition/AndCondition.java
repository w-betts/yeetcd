package yeetcd.sdk.condition;

import yeetcd.protocol.pipeline.PipelineOuterClass;
import lombok.EqualsAndHashCode;

@EqualsAndHashCode(callSuper = true)
public final class AndCondition extends Condition {
    private final Condition left;
    private final Condition right;

    private AndCondition(Condition left, Condition right) {
        this.left = left;
        this.right = right;
    }

    public Condition getLeft() {
        return left;
    }

    public Condition getRight() {
        return right;
    }

    public static Builder builder(Condition left, Condition right) {
        return new Builder(left, right);
    }

    @Override
    public PipelineOuterClass.Condition toProtobuf() {
        PipelineOuterClass.AndCondition.Builder andBuilder = PipelineOuterClass.AndCondition.newBuilder();
        left.applyTo(andBuilder, true);
        right.applyTo(andBuilder, false);
        return PipelineOuterClass.Condition.newBuilder().setAndCondition(andBuilder.build()).build();
    }

    public static class Builder {
        private final Condition left;
        private final Condition right;

        private Builder(Condition left, Condition right) {
            this.left = left;
            this.right = right;
        }

        public AndCondition build() {
            return new AndCondition(left, right);
        }
    }
}
