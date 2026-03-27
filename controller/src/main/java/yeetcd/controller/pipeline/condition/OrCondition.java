package yeetcd.controller.pipeline.condition;

import yeetcd.controller.pipeline.WorkContext;
import yeetcd.controller.pipeline.WorkResultTracker;
import yeetcd.protocol.pipeline.PipelineOuterClass;

public record OrCondition(Condition left, Condition right) implements Condition {
    @Override
    public boolean evaluate(WorkContext workContext, WorkResultTracker workResultTracker) {
        return left.evaluate(workContext, workResultTracker) || right.evaluate(workContext, workResultTracker);
    }

    public static Condition fromProtobuf(PipelineOuterClass.OrCondition orCondition) {
        return new OrCondition(Conditions.fromProtobuf(orCondition.getLeft()), Conditions.fromProtobuf(orCondition.getRight()));
    }
}
