package yeetcd.controller.pipeline.condition;

import yeetcd.controller.pipeline.WorkContext;
import yeetcd.controller.pipeline.WorkResultTracker;

public interface Condition {

    boolean evaluate(WorkContext workContext, WorkResultTracker workResultTracker);
}
