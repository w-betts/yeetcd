package yeetcd.controller.pipeline;

import yeetcd.controller.execution.ExecutionEngine;
import yeetcd.controller.pipeline.events.WorkStarted;
import yeetcd.controller.utils.CompletableFutureUtils;

import java.util.List;
import java.util.concurrent.CompletableFuture;
import java.util.stream.Collectors;

public record CompoundWorkDefinition(List<Work> finalWork) implements WorkDefinition {
    @Override
    public CompletableFuture<WorkResult> execute(WorkContext workContext, ExecutionEngine executionEngine, PipelineMetadata pipelineMetadata, Work work, WorkResultTracker workResultTracker, PipelineOutputHandler pipelineOutputHandler) {
        pipelineOutputHandler.recordEvent(new WorkStarted(work, pipelineOutputHandler.newJobStreams()));
        return CompletableFutureUtils
                .zip(finalWork.stream().map(workItem -> workItem.execute(workContext, executionEngine, pipelineMetadata, workResultTracker, pipelineOutputHandler)).collect(Collectors.toList()))
                .thenApply(workResults -> workResults.stream().allMatch(workResult -> workResult.workStatus() == WorkStatus.SUCCESS) ? WorkStatus.SUCCESS : WorkStatus.FAILURE)
                .thenApply(compoundResult -> new WorkResult(compoundResult, null, null));
    }
}
