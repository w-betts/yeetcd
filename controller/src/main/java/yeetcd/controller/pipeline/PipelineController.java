package yeetcd.controller.pipeline;

import yeetcd.controller.execution.*;
import yeetcd.controller.pipeline.events.PipelineFinished;
import yeetcd.controller.pipeline.events.PipelineStarted;
import yeetcd.controller.source.*;
import yeetcd.controller.utils.CompletableFutureUtils;
import yeetcd.protocol.pipeline.PipelineOuterClass;
import lombok.SneakyThrows;

import java.io.ByteArrayOutputStream;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CompletableFuture;
import java.util.stream.Collectors;

public class PipelineController {

    private final SourceBuilder sourceBuilder;
    private final SourceExtractor sourceExtractor;

    private final ExecutionEngine executionEngine;

    public PipelineController() {
        sourceExtractor = new SourceExtractor();
        executionEngine = ExecutionEngines.createForRuntime();
        sourceBuilder = new SourceBuilder(executionEngine);
    }

    public ExecutionEngine getExecutionEngine() {
        return executionEngine;
    }

    @SneakyThrows
    public CompletableFuture<List<Pipeline>> assemble(Source source) {
        SourceExtractionResult sourceExtractionResult = sourceExtractor.extract(source);
        return sourceBuilder
            .build(sourceExtractionResult)
            .thenCompose(buildResults -> CompletableFutureUtils
                .zip(buildResults
                    .stream()
                    .map(sourceBuildResult -> executionEngine
                        // build image that contains all built sources, including user and sdk generated artifacts
                        .buildImage(new BuildImageDefinition(
                            "%s_%s_builder".formatted(
                                source.name().toLowerCase(),
                                sourceBuildResult.yeetcdDefinition().name()
                            ),
                            source.sha256(),
                            sourceBuildResult.yeetcdDefinition().language().getImageBase(),
                            sourceBuildResult.outputDirectoriesParent(),
                            sourceBuildResult.yeetcdDefinition().artifacts().stream().map(ArtifactDefinition::name).toList(),
                            sourceBuildResult.yeetcdDefinition().language().getGeneratePipelineDefinitionsCmd()
                        ))
                        .thenCompose(buildImageResult -> {
                            ByteArrayOutputStream stdOut = new ByteArrayOutputStream();
                            return executionEngine
                                // Run a container that outputs the serialised pipelines definitions
                                .runJob(
                                    new JobDefinition(
                                        buildImageResult.imageId(),
                                        new String[]{},
                                        "/",
                                        Collections.emptyMap(),
                                        Collections.emptyMap(),
                                        Collections.emptyMap(),
                                        new JobStreams(stdOut, System.err)
                                    )
                                )
                                .thenApply(result -> {
                                    if (result.exitCode() == 0) {
                                        return stdOut.toByteArray();
                                    }
                                    else {
                                        throw new RuntimeException("unsuccessful pipeline assembly");
                                    }
                                })
                                .thenApply(output -> pipelinesFromProtobufOutput(
                                    output,
                                    buildImageResult.imageId(),
                                    sourceBuildResult.yeetcdDefinition().language()
                                ));
                        })
                    )
                    .toList()
                )
                .thenApply(listOfLists -> listOfLists.stream().flatMap(List::stream).collect(Collectors.toList()))
            );
    }

    public CompletableFuture<PipelineResult> execute(Pipeline pipeline, PipelineOutputHandler pipelineOutputHandler) {
        WorkResultTracker workResultTracker = new WorkResultTracker();
        pipelineOutputHandler.recordEvent(new PipelineStarted(pipeline));
        return CompletableFutureUtils
            .zip(pipeline.finalWork()
                .stream()
                .map(work -> work.execute(pipeline.workContext(), executionEngine, pipeline.pipelineMetadata(), workResultTracker, pipelineOutputHandler))
                .toList()
            )
            // They are actually all complete now, but still treat them as futures
            .thenCompose(finalWorkDone -> CompletableFutureUtils
                .zip(workResultTracker.getWorkResultMap().entrySet().stream().map(entry -> entry.getValue().thenApply(value -> Map.entry(entry.getKey(), value))).toList())
                .thenApply(entries -> {
                    PipelineResult pipelineResult = new PipelineResult(entries.stream().collect(Collectors.toMap(Map.Entry::getKey, Map.Entry::getValue)));
                    pipelineOutputHandler.recordEvent(new PipelineFinished(pipelineResult.pipelineStatus()));
                    return pipelineResult;
                })
            );
    }

    @SneakyThrows
    private static List<Pipeline> pipelinesFromProtobufOutput(byte[] output, String builtSourceImage, SourceLanguage sourceLanguage) {
        return PipelineOuterClass.Pipelines
            .parseFrom(output)
            .getPipelinesList()
            .stream()
            .map(protobuf -> Pipeline.fromProtobuf(protobuf, builtSourceImage, sourceLanguage))
            .toList();
    }
}
