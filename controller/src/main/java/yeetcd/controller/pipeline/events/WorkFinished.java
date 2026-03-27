package yeetcd.controller.pipeline.events;

import yeetcd.controller.pipeline.Work;
import yeetcd.controller.pipeline.WorkStatus;

public record WorkFinished(Work work, WorkStatus workStatus) implements PipelineEvent {
}
