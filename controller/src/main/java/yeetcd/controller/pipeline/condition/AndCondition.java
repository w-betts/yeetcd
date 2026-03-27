package yeetcd.controller.pipeline.condition;

import yeetcd.controller.pipeline.WorkContext;
import yeetcd.controller.pipeline.WorkResultTracker;
import yeetcd.protocol.pipeline.PipelineOuterClass;

public record AndCondition(Condition left, Condition right) implements Condition {
    @Override
    public boolean evaluate(WorkContext workContext, WorkResultTracker workResultTracker) {
        return left.evaluate(workContext, workResultTracker) && right.evaluate(workContext, workResultTracker);
    }

    static AndCondition fromProtobuf(PipelineOuterClass.AndCondition andCondition) {
        return new AndCondition(Conditions.fromProtobuf(andCondition.getLeft()), Conditions.fromProtobuf(andCondition.getRight()));
    }
}
