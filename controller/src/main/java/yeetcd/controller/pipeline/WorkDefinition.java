package yeetcd.controller.pipeline;

import yeetcd.controller.execution.ExecutionEngine;

import java.util.concurrent.CompletableFuture;

public interface WorkDefinition {

    CompletableFuture<WorkResult> execute(WorkContext workContext, ExecutionEngine executionEngine, PipelineMetadata pipelineMetadata, Work work, WorkResultTracker workResultTracker, PipelineOutputHandler pipelineOutputHandler);
}
