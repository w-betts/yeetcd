package yeetcd.controller.pipeline;

import yeetcd.controller.execution.ExecutionEngine;
import yeetcd.controller.execution.JobDefinition;
import yeetcd.controller.execution.JobStreams;
import yeetcd.controller.pipeline.events.WorkStarted;

import java.util.List;
import java.util.concurrent.CompletableFuture;

public record ContainerisedWorkDefinition(String image, List<String> cmd) implements WorkDefinition {
    @Override
    public CompletableFuture<WorkResult> execute(WorkContext workContext, ExecutionEngine executionEngine, PipelineMetadata pipelineMetadata, Work work, WorkResultTracker workResultTracker, PipelineOutputHandler pipelineOutputHandler) {
        JobStreams jobStreams = pipelineOutputHandler.newJobStreams();
        pipelineOutputHandler.recordEvent(new WorkStarted(work, jobStreams));
        return executionEngine
                .runJob(new JobDefinition(
                        image,
                        cmd.toArray(String[]::new),
                        "/",
                        work.previousWorkStdOutAsWorkContext(workResultTracker).mergeInto(workContext).workContextMap(),
                        work.previousWorkMountInputs(workResultTracker),
                        work.outputDirectoryPaths(),
                        jobStreams
                ))
                .thenApply(jobResult -> jobResult.exitCode() == 0 ? new WorkResult(WorkStatus.SUCCESS, jobResult.outputDirectoriesParent(), jobStreams) : new WorkResult(WorkStatus.FAILURE, jobResult.outputDirectoriesParent(), jobStreams));
    }
}
