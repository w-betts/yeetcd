package yeetcd.controller.pipeline;

import yeetcd.controller.execution.JobStreams;
import yeetcd.controller.pipeline.events.PipelineEvent;

public interface PipelineOutputHandler {

    void recordEvent(PipelineEvent pipelineEvent);
    JobStreams newJobStreams();
}
