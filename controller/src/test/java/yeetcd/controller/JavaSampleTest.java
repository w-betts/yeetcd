package yeetcd.controller;

import yeetcd.controller.execution.DockerExecutionEngine;
import yeetcd.controller.pipeline.*;
import yeetcd.controller.pipeline.events.*;
import yeetcd.controller.source.*;
import lombok.SneakyThrows;
import org.hamcrest.Matchers;
import org.junit.jupiter.api.Test;

import java.io.File;
import java.util.*;

import static yeetcd.controller.source.SourceExtractorTest.givenLocalProjectExtraction;
import static org.hamcrest.CoreMatchers.equalTo;
import static org.hamcrest.CoreMatchers.instanceOf;
import static org.hamcrest.MatcherAssert.assertThat;
import static org.hamcrest.Matchers.greaterThan;
import static org.hamcrest.Matchers.not;

public class JavaSampleTest {

    @SneakyThrows
    @Test
    public void shouldBuildTheJavaSampleApplication() {
        // given
        SourceExtractionResult sourceExtractionResult = givenLocalProjectExtraction();
        SourceBuilder sourceBuilder = new SourceBuilder(new DockerExecutionEngine());

        // when
        List<SourceBuildResult> sourceBuildResults = sourceBuilder.build(sourceExtractionResult).get();

        // then
        assertThat(sourceBuildResults.size(), Matchers.equalTo(1));
        File outputDirectoriesParent = sourceBuildResults.get(0).outputDirectoriesParent();
        List<String> allFiles = Arrays
                .stream(Objects.requireNonNullElse(outputDirectoriesParent.listFiles(), new File[]{}))
                .flatMap(file -> Arrays.stream(Objects.requireNonNullElse(file.list(), new String[]{})))
                .toList();

        assertThat(allFiles.size(), greaterThan(0));
        assertThat(allFiles.stream().filter(it -> it.endsWith(".jar")).findAny(), not(Matchers.equalTo(Optional.empty())));
    }

    @Test
    @SneakyThrows
    public void shouldAssembleAndExecuteTheJavaSamplePipeline() {
        // given
        String sourceName = UUID.randomUUID().toString();
        byte[] projectZip = ArchiveUtils.projectZip();
        PipelineController pipelineController = new PipelineController();

        List<Pipeline> pipelines = Collections.emptyList();
        try {
            // when
            pipelines = pipelineController.assemble(new Source(sourceName, projectZip)).get();

            // then
            Optional<Pipeline> maybeSamplePipeline = pipelines.stream().filter(pipeline -> pipeline.name().equals("sample")).findFirst();
            assertThat(maybeSamplePipeline.isPresent(), equalTo(true));
            Pipeline samplePipeline = maybeSamplePipeline.get();

            // when
            TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();
            PipelineResult pipelineResult = pipelineController.execute(samplePipeline, pipelineOutputHandler).get();

            // then
            assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
            assertThat(pipelineOutputHandler.getPipelineEvents().size(), equalTo(6));

            assertThat(pipelineOutputHandler.getPipelineEvents().get(0), instanceOf(PipelineStarted.class));

            assertThat(pipelineOutputHandler.getPipelineEvents().get(1), instanceOf(WorkStarted.class));
            WorkStarted work1Started = (WorkStarted) pipelineOutputHandler.getPipelineEvents().get(1);
            assertThat(work1Started.work().workDefinition(), instanceOf(ContainerisedWorkDefinition.class));
            assertThat(pipelineOutputHandler.getPipelineEvents().get(2), instanceOf(WorkFinished.class));
            assertThat(pipelineOutputHandler.getPipelineEvents().get(3), instanceOf(WorkStarted.class));
            WorkStarted work2Started = (WorkStarted) pipelineOutputHandler.getPipelineEvents().get(3);
            assertThat(work2Started.work().workDefinition(), instanceOf(CustomWorkDefinition.class));
            assertThat(pipelineOutputHandler.getPipelineEvents().get(4), instanceOf(WorkFinished.class));

            assertThat(pipelineOutputHandler.getPipelineEvents().get(5), instanceOf(PipelineFinished.class));
        } finally {
            pipelines.forEach(pipeline -> pipelineController.getExecutionEngine().removeImage(pipeline.pipelineMetadata().builtSourceImage()));
        }
    }

    @Test
    @SneakyThrows
    public void shouldAssembleAndExecuteTheJavaSampleCompoundPipeline() {
        // given
        String sourceName = UUID.randomUUID().toString();
        byte[] projectZip = ArchiveUtils.projectZip();
        PipelineController pipelineController = new PipelineController();

        List<Pipeline> pipelines = Collections.emptyList();
        try {
            // when
            pipelines = pipelineController.assemble(new Source(sourceName, projectZip)).get();

            // then
            Optional<Pipeline> maybeSampleCompoundPipeline = pipelines.stream().filter(pipeline -> pipeline.name().equals("sampleCompound")).findFirst();
            assertThat(maybeSampleCompoundPipeline.isPresent(), equalTo(true));
            Pipeline sampleCompoundPipeline = maybeSampleCompoundPipeline.get();

            // when
            TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();
            PipelineResult pipelineResult = pipelineController.execute(sampleCompoundPipeline, pipelineOutputHandler).get();

            // then
            assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
            assertThat(pipelineOutputHandler.getPipelineEvents().size(), equalTo(14));

            assertThat(pipelineOutputHandler.getPipelineEvents().get(0), instanceOf(PipelineStarted.class));

            assertThat(pipelineOutputHandler.getPipelineEvents().get(1), instanceOf(WorkStarted.class));
            WorkStarted work1Started = (WorkStarted) pipelineOutputHandler.getPipelineEvents().get(1);
            assertThat(work1Started.work().workDefinition(), instanceOf(CompoundWorkDefinition.class));
            assertThat(pipelineOutputHandler.getPipelineEvents().get(2), instanceOf(WorkStarted.class));
            WorkStarted work2Started = (WorkStarted) pipelineOutputHandler.getPipelineEvents().get(2);
            assertThat(work2Started.work().workDefinition(), instanceOf(ContainerisedWorkDefinition.class));
            assertThat(pipelineOutputHandler.getPipelineEvents().get(3), instanceOf(WorkFinished.class));
            assertThat(pipelineOutputHandler.getPipelineEvents().get(4), instanceOf(WorkStarted.class));
            WorkStarted work3Started = (WorkStarted) pipelineOutputHandler.getPipelineEvents().get(4);
            assertThat(work3Started.work().workDefinition(), instanceOf(CustomWorkDefinition.class));
            assertThat(pipelineOutputHandler.getPipelineEvents().get(5), instanceOf(WorkFinished.class));
            assertThat(pipelineOutputHandler.getPipelineEvents().get(6), instanceOf(WorkFinished.class));

            assertThat(pipelineOutputHandler.getPipelineEvents().get(7), instanceOf(WorkStarted.class));
            WorkStarted workStarted4 = (WorkStarted) pipelineOutputHandler.getPipelineEvents().get(1);
            assertThat(workStarted4.work().workDefinition(), instanceOf(CompoundWorkDefinition.class));
            assertThat(pipelineOutputHandler.getPipelineEvents().get(8), instanceOf(WorkStarted.class));
            WorkStarted workStarted5 = (WorkStarted) pipelineOutputHandler.getPipelineEvents().get(8);
            assertThat(workStarted5.work().workDefinition(), instanceOf(ContainerisedWorkDefinition.class));
            assertThat(pipelineOutputHandler.getPipelineEvents().get(9), instanceOf(WorkFinished.class));
            assertThat(pipelineOutputHandler.getPipelineEvents().get(10), instanceOf(WorkStarted.class));
            WorkStarted workStarted6 = (WorkStarted) pipelineOutputHandler.getPipelineEvents().get(10);
            assertThat(workStarted6.work().workDefinition(), instanceOf(CustomWorkDefinition.class));
            assertThat(pipelineOutputHandler.getPipelineEvents().get(11), instanceOf(WorkFinished.class));
            assertThat(pipelineOutputHandler.getPipelineEvents().get(12), instanceOf(WorkFinished.class));

            assertThat(pipelineOutputHandler.getPipelineEvents().get(13), instanceOf(PipelineFinished.class));
        } finally {
            pipelines.forEach(pipeline -> pipelineController.getExecutionEngine().removeImage(pipeline.pipelineMetadata().builtSourceImage()));
        }
    }

    @Test
    @SneakyThrows
    public void shouldAssembleAndExecuteTheJavaSampleWorkContextPipeline() {
        // given
        String sourceName = UUID.randomUUID().toString();
        byte[] projectZip = ArchiveUtils.projectZip();
        PipelineController pipelineController = new PipelineController();

        List<Pipeline> pipelines = Collections.emptyList();
        try {
            // when
            pipelines = pipelineController.assemble(new Source(sourceName, projectZip)).get();

            // then
            Optional<Pipeline> maybesampleWithWorkContextPipeline = pipelines.stream().filter(pipeline -> pipeline.name().equals("sampleWithWorkContext")).findFirst();
            assertThat(maybesampleWithWorkContextPipeline.isPresent(), equalTo(true));
            Pipeline sampleWithWorkContext = maybesampleWithWorkContextPipeline.get();

            // when
            TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();
            PipelineResult pipelineResult = pipelineController.execute(sampleWithWorkContext, pipelineOutputHandler).get();

            // then
            assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
        } finally {
            pipelines.forEach(pipeline -> pipelineController.getExecutionEngine().removeImage(pipeline.pipelineMetadata().builtSourceImage()));
        }
    }

    @Test
    @SneakyThrows
    public void shouldAssembleAndExecuteTheJavaSampleWithParametersPipeline() {
        // given
        String sourceName = UUID.randomUUID().toString();
        byte[] projectZip = ArchiveUtils.projectZip();
        PipelineController pipelineController = new PipelineController();

        List<Pipeline> pipelines = Collections.emptyList();
        try {
            // when
            pipelines = pipelineController.assemble(new Source(sourceName, projectZip)).get();

            // then
            Optional<Pipeline> maybeSampleWithParametersPipeline = pipelines.stream().filter(pipeline -> pipeline.name().equals("sampleWithParameters")).findFirst();
            assertThat(maybeSampleWithParametersPipeline.isPresent(), equalTo(true));

            Pipeline sampleWithParametersAndArgumentsSupplied = maybeSampleWithParametersPipeline.get()
                    .withArguments(Arguments.of(
                            "PARAMETER_NAME",
                            "other"
                    ));

            // when
            TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();
            PipelineResult pipelineResult = pipelineController.execute(sampleWithParametersAndArgumentsSupplied, pipelineOutputHandler).get();

            // then
            assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
        } finally {
            pipelines.forEach(pipeline -> pipelineController.getExecutionEngine().removeImage(pipeline.pipelineMetadata().builtSourceImage()));
        }
    }

    @Test
    @SneakyThrows
    public void shouldAssembleAndExecuteTheJavaSampleWithWorkOutputsPipeline() {
        // given
        String sourceName = UUID.randomUUID().toString();
        byte[] projectZip = ArchiveUtils.projectZip();
        PipelineController pipelineController = new PipelineController();

        List<Pipeline> pipelines = Collections.emptyList();
        try {
            // when
            pipelines = pipelineController.assemble(new Source(sourceName, projectZip)).get();

            // then
            Optional<Pipeline> maybeSampleWithWorkOutputsPipeline = pipelines.stream().filter(pipeline -> pipeline.name().equals("sampleWithWorkOutputs")).findFirst();
            assertThat(maybeSampleWithWorkOutputsPipeline.isPresent(), equalTo(true));

            Pipeline sampleWithWorkOutputs = maybeSampleWithWorkOutputsPipeline.get();

            // when
            TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();
            PipelineResult pipelineResult = pipelineController.execute(sampleWithWorkOutputs, pipelineOutputHandler).get();

            // then
            assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
        } finally {
            pipelines.forEach(pipeline -> pipelineController.getExecutionEngine().removeImage(pipeline.pipelineMetadata().builtSourceImage()));
        }
    }

    @Test
    @SneakyThrows
    public void shouldAssembleAndExecuteTheJavaSampleWithConditionsPipeline() {
        // given
        String sourceName = UUID.randomUUID().toString();
        byte[] projectZip = ArchiveUtils.projectZip();
        PipelineController pipelineController = new PipelineController();

        List<Pipeline> pipelines = Collections.emptyList();
        try {
            // when
            pipelines = pipelineController.assemble(new Source(sourceName, projectZip)).get();

            // then
            Optional<Pipeline> maybeSampleWithConditionsPipeline = pipelines.stream().filter(pipeline -> pipeline.name().equals("sampleWithConditions")).findFirst();
            assertThat(maybeSampleWithConditionsPipeline.isPresent(), equalTo(true));

            Pipeline sampleWithConditions = maybeSampleWithConditionsPipeline.get();

            // when
            TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();
            PipelineResult pipelineResult = pipelineController.execute(sampleWithConditions, pipelineOutputHandler).get();

            // then
            assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));

            assertThat(pipelineOutputHandler.getPipelineEvents().size(), equalTo(6));

            assertThat(pipelineOutputHandler.getPipelineEvents().get(0), instanceOf(PipelineStarted.class));

            assertThat(pipelineOutputHandler.getPipelineEvents().get(1), instanceOf(WorkStarted.class));
            WorkStarted work1Started = (WorkStarted) pipelineOutputHandler.getPipelineEvents().get(1);
            assertThat(work1Started.work().workDefinition(), instanceOf(CustomWorkDefinition.class));
            assertThat(pipelineOutputHandler.getPipelineEvents().get(2), instanceOf(WorkFinished.class));
            WorkFinished work1Finished = (WorkFinished) pipelineOutputHandler.getPipelineEvents().get(2);
            assertThat(work1Finished.workStatus(), equalTo(WorkStatus.SUCCESS));

            assertThat(pipelineOutputHandler.getPipelineEvents().get(3), instanceOf(WorkFinished.class));
            WorkFinished work2Finished = (WorkFinished) pipelineOutputHandler.getPipelineEvents().get(3);
            assertThat(work2Finished.workStatus(), equalTo(WorkStatus.SKIPPED));

            assertThat(pipelineOutputHandler.getPipelineEvents().get(4), instanceOf(WorkFinished.class));
            WorkFinished work3Finished = (WorkFinished) pipelineOutputHandler.getPipelineEvents().get(4);
            assertThat(work3Finished.workStatus(), equalTo(WorkStatus.SKIPPED));

            assertThat(pipelineOutputHandler.getPipelineEvents().get(5), instanceOf(PipelineFinished.class));
        } finally {
            pipelines.forEach(pipeline -> pipelineController.getExecutionEngine().removeImage(pipeline.pipelineMetadata().builtSourceImage()));
        }
    }

    @Test
    @SneakyThrows
    public void shouldAssembleAndExecuteTheJavaSampleDynamicWorkPipeline() {
        // given
        String sourceName = UUID.randomUUID().toString();
        byte[] projectZip = ArchiveUtils.projectZip();
        PipelineController pipelineController = new PipelineController();

        List<Pipeline> pipelines = Collections.emptyList();
        try {
            // when
            pipelines = pipelineController.assemble(new Source(sourceName, projectZip)).get();

            // then
            Optional<Pipeline> maybeSampleDynamicWorkPipeline = pipelines.stream().filter(pipeline -> pipeline.name().equals("sampleDynamicWork")).findFirst();
            assertThat(maybeSampleDynamicWorkPipeline.isPresent(), equalTo(true));

            Pipeline sampleDynamicWorkPipeline = maybeSampleDynamicWorkPipeline.get();

            // when
            Pipeline sampleDynamicWorkWithWorkCountOf1 = sampleDynamicWorkPipeline.withArguments(Arguments.of("WORK_COUNT", "1"));
            TestPipelineOutputHandler pipelineOutputHandler1 = new TestPipelineOutputHandler();
            PipelineResult pipelineResult = pipelineController.execute(sampleDynamicWorkWithWorkCountOf1, pipelineOutputHandler1).get();

            // then
            assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
            assertThat(pipelineOutputHandler1.getPipelineEvents().size(), equalTo(8));

            // when
            Pipeline sampleDynamicWorkWithWorkCountOf2 = sampleDynamicWorkPipeline.withArguments(Arguments.of("WORK_COUNT", "2"));
            TestPipelineOutputHandler pipelineOutputHandler2 = new TestPipelineOutputHandler();
            PipelineResult pipelineResult2 = pipelineController.execute(sampleDynamicWorkWithWorkCountOf2, pipelineOutputHandler2).get();

            // then
            assertThat(pipelineResult2.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
            assertThat(pipelineOutputHandler2.getPipelineEvents().size(), equalTo(10));
        } finally {
            pipelines.forEach(pipeline -> pipelineController.getExecutionEngine().removeImage(pipeline.pipelineMetadata().builtSourceImage()));
        }
    }
}
