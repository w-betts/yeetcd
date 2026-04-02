package yeetcd.sdk.condition;

import yeetcd.protocol.pipeline.PipelineOuterClass;
import lombok.EqualsAndHashCode;

@EqualsAndHashCode(callSuper = true)
public final class OrCondition extends Condition {
    private final Condition left;
    private final Condition right;

    private OrCondition(Condition left, Condition right) {
        this.left = left;
        this.right = right;
    }

    public Condition getLeft() {
        return left;
    }

    public Condition getRight() {
        return right;
    }

    @Override
    public PipelineOuterClass.Condition toProtobuf() {
        PipelineOuterClass.OrCondition.Builder orBuilder = PipelineOuterClass.OrCondition.newBuilder();
        left.applyTo(orBuilder, true);
        right.applyTo(orBuilder, false);
        return PipelineOuterClass.Condition.newBuilder().setOrCondition(orBuilder.build()).build();
    }

    public static Builder builder(Condition left, Condition right) {
        return new Builder(left, right);
    }

    public static class Builder {
        private final Condition left;
        private final Condition right;

        private Builder(Condition left, Condition right) {
            this.left = left;
            this.right = right;
        }

        public OrCondition build() {
            return new OrCondition(left, right);
        }
    }
}
