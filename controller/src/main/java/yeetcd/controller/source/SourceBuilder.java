package yeetcd.controller.source;

import yeetcd.controller.execution.ExecutionEngine;
import yeetcd.controller.execution.OnDiskMountInput;
import yeetcd.controller.execution.JobDefinition;
import yeetcd.controller.execution.JobStreams;
import yeetcd.controller.utils.CompletableFutureUtils;
import lombok.SneakyThrows;

import java.io.File;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

public class SourceBuilder {

    // TODO should have limited capacity
    private static final ConcurrentHashMap<String, CompletableFuture<List<SourceBuildResult>>> cache = new ConcurrentHashMap<>();
    private final ExecutionEngine executionEngine;

    public SourceBuilder(ExecutionEngine executionEngine) {
        this.executionEngine = executionEngine;
    }

    @SneakyThrows
    public CompletableFuture<List<SourceBuildResult>> build(SourceExtractionResult sourceExtractionResult) {
        return cache.computeIfAbsent(sourceExtractionResult.source().sha256(), key -> CompletableFutureUtils.zip(sourceExtractionResult
                .yeetcdDefinitions()
                .entrySet()
                .stream()
                .map(entry -> {
                    String sourceMountDir = "/var/yeetcd";
                    String workingDir = "/var/yeetcd/" + entry.getKey().replaceAll("/yeetcd.yaml", "");
                    Map<String, String> outputDirectoryPaths = entry.getValue()
                            .artifacts()
                            .stream()
                            .collect(Collectors.toMap(ArtifactDefinition::name, artifactDefinition -> workingDir + "/" + artifactDefinition.path()));
                    return executionEngine
                            .runJob(JobDefinition
                                    .builder()
                                    .image(entry.getValue().buildImage())
                                    .cmd(entry.getValue().buildCmd().split("\\s"))
                                    .workingDir(workingDir)
                                    .inputFilePaths(Map.of(
                                            sourceMountDir, new OnDiskMountInput(sourceExtractionResult.directory()),
                                            "/root/.m2", new OnDiskMountInput(new File(System.getProperty("user.home") + "/.m2"))
                                    )) // source code
                                    .jobStreams(new JobStreams(System.out, System.err))
                                    .outputDirectoryPaths(outputDirectoryPaths)  // built jars
                                    .build()
                            )
                            .thenApply(taskResult -> new SourceBuildResult(entry.getValue(), taskResult.outputDirectoriesParent()));
                })
                .toList()
        ));
    }
}
