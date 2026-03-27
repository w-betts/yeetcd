package yeetcd.controller.pipeline;

import yeetcd.controller.execution.JobStreams;
import yeetcd.controller.pipeline.events.PipelineEvent;

import java.util.List;
import java.util.concurrent.CopyOnWriteArrayList;

public class TestPipelineOutputHandler implements PipelineOutputHandler {
    private final List<PipelineEvent> pipelineEvents = new CopyOnWriteArrayList<>();

    @Override
    public void recordEvent(PipelineEvent pipelineEvent) {
        pipelineEvents.add(pipelineEvent);
    }

    @Override
    public JobStreams newJobStreams() {
        return new JobStreams(System.out, System.err);
    }

    public List<PipelineEvent> getPipelineEvents() {
        return pipelineEvents;
    }

    @Override
    public String toString() {
        return "TestPipelineOutputHandler{" +
               "pipelineEvents=" + pipelineEvents +
               '}';
    }
}
