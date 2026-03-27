package yeetcd.controller.pipeline;

import yeetcd.controller.execution.ExecutionEngine;
import yeetcd.controller.execution.MountInput;
import yeetcd.controller.pipeline.condition.Condition;
import yeetcd.controller.pipeline.condition.Conditions;
import yeetcd.controller.pipeline.events.WorkFinished;
import yeetcd.controller.utils.CompletableFutureUtils;
import yeetcd.protocol.pipeline.PipelineOuterClass;
import lombok.SneakyThrows;
import org.apache.commons.lang3.StringUtils;

import java.nio.charset.StandardCharsets;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.concurrent.CompletableFuture;
import java.util.stream.Collectors;

public record Work(String id, String description, WorkContext workContext, WorkDefinition workDefinition, Condition condition, List<WorkOutputPath> outputPaths, List<PreviousWork> previousWork) {

    CompletableFuture<WorkResult> execute(WorkContext containingContext, ExecutionEngine executionEngine, PipelineMetadata pipelineMetadata, WorkResultTracker workResultTracker, PipelineOutputHandler pipelineOutputHandler) {
        return CompletableFutureUtils
            .zip(previousWork
                .stream()
                .map(previousWork -> previousWork.work().execute(containingContext, executionEngine, pipelineMetadata, workResultTracker, pipelineOutputHandler))
                .toList()
            )
            .thenCompose(predecessorResults -> workResultTracker.getOrExecute(this, () -> {
                if (condition.evaluate(workContext, workResultTracker)) {
                    return workDefinition
                        .execute(
                            workContext.mergeInto(containingContext),
                            executionEngine,
                            pipelineMetadata,
                            this,
                            workResultTracker,
                            pipelineOutputHandler
                        )
                        .thenCompose(workResult -> {
                            if (workResult.workStatus() != WorkStatus.SUCCESS || !(workDefinition instanceof DynamicWorkGeneratingWorkDefinition)) {
                                return CompletableFuture.completedFuture(workResult);
                            }
                            else  {
                                Work work = Work.fromProtobufBytes(workResult.jobStreams().getStdOut());
                                return work.execute(containingContext, executionEngine, pipelineMetadata, workResultTracker, pipelineOutputHandler);
                            }
                        })
                        .thenApply(workResult -> {
                            pipelineOutputHandler.recordEvent(new WorkFinished(this, workResult.workStatus()));
                            return workResult;
                        });
                } else {
                    pipelineOutputHandler.recordEvent(new WorkFinished(this, WorkStatus.SKIPPED));
                    return CompletableFuture.completedFuture(new WorkResult(WorkStatus.SKIPPED, null, null));
                }
            }));
    }

    @SneakyThrows
    private static Work fromProtobufBytes(byte[] bytes) {
        return fromProtobuf(PipelineOuterClass.Work.parseFrom(bytes));
    }

    public static Work fromProtobuf(PipelineOuterClass.Work workProtobuf) {
        return new Work(
            workProtobuf.getId(),
            workProtobuf.getDescription(),
            new WorkContext(workProtobuf.getWorkContextMap()),
            WorkDefinitions.fromWorkProtobuf(workProtobuf),
            Conditions.fromProtobuf(workProtobuf.getCondition()),
            workProtobuf.getOutputPathsList().stream().map(WorkOutputPath::fromProtobuf).toList(),
            workProtobuf.getPreviousWorkList().stream().map(PreviousWork::fromProtobuf).toList()
        );
    }

    WorkContext previousWorkStdOutAsWorkContext(WorkResultTracker workResultTracker) {
        Map<String, String> previousWorkStdOuts = new HashMap<>();
        previousWork.stream()
            .filter(previousWork -> StringUtils.isNotBlank(previousWork.stdOutEnvVar()))
            .forEach(previousWork -> previousWorkStdOuts.put(
                previousWork.stdOutEnvVar(),
                new String(workResultTracker.stdOut(previousWork), StandardCharsets.UTF_8)
            ));
        return WorkContext.fromMap(previousWorkStdOuts);
    }

    Map<String, MountInput> previousWorkMountInputs(WorkResultTracker workResultTracker) {
        return previousWork.stream()
            .filter(previousWork -> StringUtils.isNotBlank(previousWork.outputPathsMount()))
            .collect(Collectors.toMap(PreviousWork::outputPathsMount, workResultTracker::outputDirectoriesMountInput));
    }

    Map<String, String> outputDirectoryPaths() {
        return outputPaths.stream().collect(Collectors.toMap(WorkOutputPath::name, WorkOutputPath::path));
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        Work work = (Work) o;
        return Objects.equals(id, work.id);
    }

    @Override
    public int hashCode() {
        return Objects.hash(id);
    }
}
