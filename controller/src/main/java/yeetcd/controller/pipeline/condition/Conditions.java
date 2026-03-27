package yeetcd.controller.pipeline.condition;

import yeetcd.protocol.pipeline.PipelineOuterClass;

public class Conditions {

    public static final Condition PREVIOUS_WORK_SUCCESS = new PreviousWorkStatusCondition(PreviousWorkStatusCondition.Status.SUCCESS);

    public static Condition fromProtobuf(PipelineOuterClass.Condition condition) {
        if (condition.hasAndCondition()) {
            return AndCondition.fromProtobuf(condition.getAndCondition());
        }
        if (condition.hasOrCondition()) {
            return OrCondition.fromProtobuf(condition.getOrCondition());
        }
        if (condition.hasNotCondition()) {
            return NotCondition.fromProtobuf(condition.getNotCondition());
        }
        if (condition.hasWorkContextCondition()) {
            return WorkContextCondition.fromProtobuf(condition.getWorkContextCondition());
        }
        if (condition.hasPreviousWorkStatusCondition()) {
            return PreviousWorkStatusCondition.fromProtobuf(condition.getPreviousWorkStatusCondition());
        }
        throw new IllegalArgumentException("Unrecognised condition");
    }

}
