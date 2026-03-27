package yeetcd.sdk.condition;

import yeetcd.protocol.pipeline.PipelineOuterClass;
import lombok.EqualsAndHashCode;

@EqualsAndHashCode
public abstract class Condition {

    public void applyTo(PipelineOuterClass.Work.Builder builder) {
        builder.setCondition(toProtobuf());
    }

    public void applyTo(PipelineOuterClass.AndCondition.Builder builder, boolean isLeft) {
        if (isLeft) {
            builder.setLeft(toProtobuf());
        } else {
            builder.setRight(toProtobuf());
        }
    }

    public void applyTo(PipelineOuterClass.OrCondition.Builder builder, boolean isLeft) {
        if (isLeft) {
            builder.setLeft(toProtobuf());
        } else {
            builder.setRight(toProtobuf());
        }
    }

    public void applyTo(PipelineOuterClass.NotCondition.Builder builder) {
        builder.setCondition(toProtobuf());
    }

    abstract PipelineOuterClass.Condition toProtobuf();

}
