package yeetcd.controller.pipeline.events;

import yeetcd.controller.pipeline.PipelineStatus;

public record PipelineFinished(PipelineStatus pipelineStatus) implements PipelineEvent {
}
