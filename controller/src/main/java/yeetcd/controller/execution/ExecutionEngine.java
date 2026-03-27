package yeetcd.controller.execution;

import java.util.concurrent.CompletableFuture;

public interface ExecutionEngine {

    CompletableFuture<BuildImageResult> buildImage(BuildImageDefinition buildImageDefinition);
    CompletableFuture<Void> removeImage(String image);
    CompletableFuture<JobResult> runJob(JobDefinition jobDefinition);
}
