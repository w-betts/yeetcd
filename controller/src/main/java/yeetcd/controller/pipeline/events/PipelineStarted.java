package yeetcd.controller.pipeline.events;

import yeetcd.controller.pipeline.Pipeline;

public record PipelineStarted(Pipeline pipeline) implements PipelineEvent {
}
