package yeetcd.controller.pipeline.condition;

import yeetcd.controller.pipeline.WorkContext;
import yeetcd.controller.pipeline.WorkResultTracker;
import yeetcd.protocol.pipeline.PipelineOuterClass;

public record NotCondition(Condition condition) implements Condition {
    @Override
    public boolean evaluate(WorkContext workContext, WorkResultTracker workResultTracker) {
        return !condition.evaluate(workContext, workResultTracker);
    }

    static NotCondition fromProtobuf(PipelineOuterClass.NotCondition condition) {
        return new NotCondition(Conditions.fromProtobuf(condition.getCondition()));
    }
}
