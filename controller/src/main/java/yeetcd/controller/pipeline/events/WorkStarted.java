package yeetcd.controller.pipeline.events;

import yeetcd.controller.execution.JobStreams;
import yeetcd.controller.pipeline.Work;

public record WorkStarted(Work work, JobStreams jobStreams) implements PipelineEvent {
}
