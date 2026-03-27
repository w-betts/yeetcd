package yeetcd.controller.pipeline;

import yeetcd.controller.pipeline.condition.*;
import yeetcd.controller.pipeline.events.*;
import yeetcd.controller.source.SourceLanguage;
import lombok.SneakyThrows;
import org.hamcrest.CoreMatchers;
import org.junit.jupiter.api.Test;

import java.nio.charset.StandardCharsets;
import java.util.*;
import java.util.stream.Collectors;

import static yeetcd.controller.execution.AbstractExecutionEngineTest.TEST_IMAGE;
import static org.hamcrest.CoreMatchers.*;
import static org.hamcrest.MatcherAssert.assertThat;

public class PipelineControllerTest {

    private static final String FAILED_TASK_OUTPUT = "failed work\n";
    private static final String SUCCESSFUL_TASK_1_OUTPUT = "successful work 1\n";
    private static final String SUCCESSFUL_TASK_2_OUTPUT = "successful work 2\n";
    private static final String SUCCESSFUL_TASK_3_OUTPUT = "successful work 3\n";

    private static final ContainerisedWorkDefinition failedContainerWorkAction = new ContainerisedWorkDefinition(
            TEST_IMAGE,
            List.of("bash", "-c", "echo -n '%s'; exit 1".formatted(FAILED_TASK_OUTPUT))
    );

    private static final ContainerisedWorkDefinition successfulContainerWorkAction1 = new ContainerisedWorkDefinition(
            TEST_IMAGE,
            List.of("bash", "-c", "echo -n '%s'".formatted(SUCCESSFUL_TASK_1_OUTPUT))
    );

    private static final ContainerisedWorkDefinition successfulContainerWorkAction2 = new ContainerisedWorkDefinition(
            TEST_IMAGE,
            List.of("bash", "-c", "echo -n '%s'".formatted(SUCCESSFUL_TASK_2_OUTPUT))
    );

    private static final ContainerisedWorkDefinition successfulContainerWorkAction3 = new ContainerisedWorkDefinition(
            TEST_IMAGE,
            List.of("bash", "-c", "echo -n '%s'".formatted(SUCCESSFUL_TASK_3_OUTPUT))
    );

    private static final ContainerisedWorkDefinition envDumpWork = new ContainerisedWorkDefinition(
            TEST_IMAGE,
            List.of("env")
    );

    @SneakyThrows
    @Test
    public void shouldPassThePipelineWhenAllWorkIsSuccessful() {
        // given
        PipelineController pipelineController = new PipelineController();

        Work work1 = new Work(UUID.randomUUID().toString(), "work1", WorkContext.empty(), successfulContainerWorkAction1, Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), Collections.emptyList());
        Work work2 = new Work(UUID.randomUUID().toString(), "work2", WorkContext.empty(), successfulContainerWorkAction2, Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), List.of(new PreviousWork(work1, null, null)));
        Pipeline pipeline = new Pipeline(
                UUID.randomUUID().toString(),
                List.of(work2),
                new PipelineMetadata("", "", SourceLanguage.JAVA)
        );
        TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();

        // when
        PipelineResult pipelineResult = pipelineController.execute(pipeline, pipelineOutputHandler).get();

        // then
        assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work1).workStatus(), equalTo(WorkStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work2).workStatus(), equalTo(WorkStatus.SUCCESS));

        assertThat(pipelineOutputHandler.getPipelineEvents().size(), equalTo(6));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(0), equalTo(new PipelineStarted(pipeline)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(1), SUCCESSFUL_TASK_1_OUTPUT);
        assertThat(pipelineOutputHandler.getPipelineEvents().get(2), equalTo(new WorkFinished(work1, WorkStatus.SUCCESS)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(3), SUCCESSFUL_TASK_2_OUTPUT);
        assertThat(pipelineOutputHandler.getPipelineEvents().get(4), equalTo(new WorkFinished(work2, WorkStatus.SUCCESS)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(5), equalTo(new PipelineFinished(PipelineStatus.SUCCESS)));
    }

    @SneakyThrows
    @Test
    public void shouldFailThePipelineWhenOneWorkFails() {
        // given
        PipelineController pipelineController = new PipelineController();

        Work work1 = new Work(UUID.randomUUID().toString(), "work1", WorkContext.empty(), successfulContainerWorkAction1, Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), Collections.emptyList());
        Work work2 = new Work(UUID.randomUUID().toString(), "work2", WorkContext.empty(), failedContainerWorkAction, Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), List.of(new PreviousWork(work1, null, null)));
        Pipeline pipeline = new Pipeline(
                UUID.randomUUID().toString(),
                List.of(work2),
                new PipelineMetadata("", "", SourceLanguage.JAVA)
        );
        TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();

        // when
        PipelineResult pipelineResult = pipelineController.execute(pipeline, pipelineOutputHandler).get();

        // then
        assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.FAILURE));
        assertThat(pipelineResult.workResults().get(work1).workStatus(), equalTo(WorkStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work2).workStatus(), equalTo(WorkStatus.FAILURE));

        assertThat(pipelineOutputHandler.getPipelineEvents().size(), equalTo(6));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(0), equalTo(new PipelineStarted(pipeline)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(1), SUCCESSFUL_TASK_1_OUTPUT);
        assertThat(pipelineOutputHandler.getPipelineEvents().get(2), equalTo(new WorkFinished(work1, WorkStatus.SUCCESS)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(3), FAILED_TASK_OUTPUT);
        assertThat(pipelineOutputHandler.getPipelineEvents().get(4), equalTo(new WorkFinished(work2, WorkStatus.FAILURE)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(5), equalTo(new PipelineFinished(PipelineStatus.FAILURE)));
    }

    @SneakyThrows
    @Test
    public void shouldSkipWorksWhichDependOnFailedWork() {
        // given
        PipelineController pipelineController = new PipelineController();

        Work work1 = new Work(UUID.randomUUID().toString(), "work1", WorkContext.empty(), failedContainerWorkAction, Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), Collections.emptyList());
        Work work2 = new Work(UUID.randomUUID().toString(), "work2", WorkContext.empty(), successfulContainerWorkAction1, Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), List.of(new PreviousWork(work1, null, null)));
        Pipeline pipeline = new Pipeline(
                UUID.randomUUID().toString(),
                List.of(work2),
                new PipelineMetadata("", "", SourceLanguage.JAVA)
        );
        TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();

        // when
        PipelineResult pipelineResult = pipelineController.execute(pipeline, pipelineOutputHandler).get();

        // then
        assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.FAILURE));
        assertThat(pipelineResult.workResults().get(work1).workStatus(), equalTo(WorkStatus.FAILURE));
        assertThat(pipelineResult.workResults().get(work2).workStatus(), equalTo(WorkStatus.SKIPPED));

        assertThat(pipelineOutputHandler.getPipelineEvents().size(), equalTo(5));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(0), equalTo(new PipelineStarted(pipeline)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(1), FAILED_TASK_OUTPUT);
        assertThat(pipelineOutputHandler.getPipelineEvents().get(2), equalTo(new WorkFinished(work1, WorkStatus.FAILURE)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(3), equalTo(new WorkFinished(work2, WorkStatus.SKIPPED)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(4), equalTo(new PipelineFinished(PipelineStatus.FAILURE)));
    }

    @SneakyThrows
    @Test
    public void shouldRunExplicitlyListedAndDependencyWorkOnceOnly() {
        // given
        PipelineController pipelineController = new PipelineController();

        Work work1 = new Work(UUID.randomUUID().toString(), "work1", WorkContext.empty(), successfulContainerWorkAction1, Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), Collections.emptyList());
        Work work2 = new Work(UUID.randomUUID().toString(), "work2", WorkContext.empty(), successfulContainerWorkAction2, Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), List.of(new PreviousWork(work1, null, null)));
        Pipeline pipeline = new Pipeline(
                UUID.randomUUID().toString(),
                List.of(work1, work2),
                new PipelineMetadata("", "", SourceLanguage.JAVA)
        );
        TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();

        // when
        PipelineResult pipelineResult = pipelineController.execute(pipeline, pipelineOutputHandler).get();

        // then
        assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work1).workStatus(), equalTo(WorkStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work2).workStatus(), equalTo(WorkStatus.SUCCESS));

        assertThat(pipelineOutputHandler.getPipelineEvents().size(), equalTo(6));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(0), equalTo(new PipelineStarted(pipeline)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(1), SUCCESSFUL_TASK_1_OUTPUT);
        assertThat(pipelineOutputHandler.getPipelineEvents().get(2), equalTo(new WorkFinished(work1, WorkStatus.SUCCESS)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(3), SUCCESSFUL_TASK_2_OUTPUT);
        assertThat(pipelineOutputHandler.getPipelineEvents().get(4), equalTo(new WorkFinished(work2, WorkStatus.SUCCESS)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(5), equalTo(new PipelineFinished(PipelineStatus.SUCCESS)));
    }

    @SneakyThrows
    @Test
    public void shouldRunMultipleDependenciesOfAWork() {
        // given
        PipelineController pipelineController = new PipelineController();

        Work work1 = new Work(UUID.randomUUID().toString(), "work1", WorkContext.empty(), successfulContainerWorkAction1, Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), Collections.emptyList());
        Work work2 = new Work(UUID.randomUUID().toString(), "work2", WorkContext.empty(), successfulContainerWorkAction2, Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), Collections.emptyList());
        Work work3 = new Work(UUID.randomUUID().toString(), "work3", WorkContext.empty(), successfulContainerWorkAction3, Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), List.of(new PreviousWork(work1, null, null), new PreviousWork(work2, null, null)));
        Pipeline pipeline = new Pipeline(
                UUID.randomUUID().toString(),
                List.of(work2, work3),
                new PipelineMetadata("", "", SourceLanguage.JAVA)
        );
        TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();

        // when
        PipelineResult pipelineResult = pipelineController.execute(pipeline, pipelineOutputHandler).get();

        // then
        assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work1).workStatus(), equalTo(WorkStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work2).workStatus(), equalTo(WorkStatus.SUCCESS));

        assertThat(pipelineOutputHandler.getPipelineEvents().size(), equalTo(8));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(0), equalTo(new PipelineStarted(pipeline)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(1), SUCCESSFUL_TASK_2_OUTPUT, SUCCESSFUL_TASK_3_OUTPUT);
        // It's awkward to assert here as we don't know if the start/finish will come first
        assertThat(pipelineOutputHandler.getPipelineEvents().get(2), anyOf(instanceOf(WorkStarted.class), instanceOf(WorkFinished.class)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(3), anyOf(instanceOf(WorkStarted.class), instanceOf(WorkFinished.class)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(4), anyOf(
                equalTo(new WorkFinished(work1, WorkStatus.SUCCESS)),
                equalTo(new WorkFinished(work2, WorkStatus.SUCCESS))
        ));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(5), SUCCESSFUL_TASK_3_OUTPUT);
        assertThat(pipelineOutputHandler.getPipelineEvents().get(6), equalTo(new WorkFinished(work3, WorkStatus.SUCCESS)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(7), equalTo(new PipelineFinished(PipelineStatus.SUCCESS)));
    }

    @SneakyThrows
    @Test
    public void shouldRunADependencyOfMultipleWorksOnceOnly() {
        // given
        PipelineController pipelineController = new PipelineController();

        Work work1 = new Work(UUID.randomUUID().toString(), "work1", WorkContext.empty(), successfulContainerWorkAction1, Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), Collections.emptyList());
        Work work2 = new Work(UUID.randomUUID().toString(), "work2", WorkContext.empty(), successfulContainerWorkAction2, Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), List.of(new PreviousWork(work1, null, null)));
        Work work3 = new Work(UUID.randomUUID().toString(), "work3", WorkContext.empty(), successfulContainerWorkAction3, Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), List.of(new PreviousWork(work1, null, null)));
        Pipeline pipeline = new Pipeline(
                UUID.randomUUID().toString(),
                List.of(work2, work3),
                new PipelineMetadata("", "", SourceLanguage.JAVA)
        );
        TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();

        // when
        PipelineResult pipelineResult = pipelineController.execute(pipeline, pipelineOutputHandler).get();

        // then
        assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work1).workStatus(), equalTo(WorkStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work2).workStatus(), equalTo(WorkStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work3).workStatus(), equalTo(WorkStatus.SUCCESS));

        assertThat(pipelineOutputHandler.getPipelineEvents().size(), equalTo(8));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(0), equalTo(new PipelineStarted(pipeline)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(1), SUCCESSFUL_TASK_1_OUTPUT);
        assertThat(pipelineOutputHandler.getPipelineEvents().get(2), equalTo(new WorkFinished(work1, WorkStatus.SUCCESS)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(3), SUCCESSFUL_TASK_2_OUTPUT, SUCCESSFUL_TASK_3_OUTPUT);
        // It's awkward to assert here as we don't know if the start/finish will come first
        assertThat(pipelineOutputHandler.getPipelineEvents().get(4), anyOf(instanceOf(WorkStarted.class), instanceOf(WorkFinished.class)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(5), anyOf(instanceOf(WorkStarted.class), instanceOf(WorkFinished.class)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(6), anyOf(
                equalTo(new WorkFinished(work2, WorkStatus.SUCCESS)),
                equalTo(new WorkFinished(work3, WorkStatus.SUCCESS))
        ));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(7), equalTo(new PipelineFinished(PipelineStatus.SUCCESS)));
    }

    @SneakyThrows
    @Test
    public void shouldRunCompoundWork() {
        // given
        PipelineController pipelineController = new PipelineController();

        Work work1 = new Work(UUID.randomUUID().toString(), "work1", WorkContext.empty(), successfulContainerWorkAction1, Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), Collections.emptyList());
        Work work2 = new Work(UUID.randomUUID().toString(), "work2", WorkContext.empty(), successfulContainerWorkAction2, Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), List.of(new PreviousWork(work1, null, null)));

        Work compoundWork = new Work(UUID.randomUUID().toString(), "work2", WorkContext.empty(), new CompoundWorkDefinition(List.of(work2)), Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), Collections.emptyList());
        Pipeline pipeline = new Pipeline(
                UUID.randomUUID().toString(),
                List.of(compoundWork),
                new PipelineMetadata("", "", SourceLanguage.JAVA)
        );
        TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();

        // when
        PipelineResult pipelineResult = pipelineController.execute(pipeline, pipelineOutputHandler).get();

        // then
        assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work1).workStatus(), equalTo(WorkStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work2).workStatus(), equalTo(WorkStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(compoundWork).workStatus(), equalTo(WorkStatus.SUCCESS));

        assertThat(pipelineOutputHandler.getPipelineEvents().size(), equalTo(8));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(0), equalTo(new PipelineStarted(pipeline)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(1), "");
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(2), SUCCESSFUL_TASK_1_OUTPUT);
        assertThat(pipelineOutputHandler.getPipelineEvents().get(3), equalTo(new WorkFinished(work1, WorkStatus.SUCCESS)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(4), SUCCESSFUL_TASK_2_OUTPUT);
        assertThat(pipelineOutputHandler.getPipelineEvents().get(5), equalTo(new WorkFinished(work2, WorkStatus.SUCCESS)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(6), equalTo(new WorkFinished(compoundWork, WorkStatus.SUCCESS)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(7), equalTo(new PipelineFinished(PipelineStatus.SUCCESS)));
    }

    @SneakyThrows
    @Test
    public void shouldRunIdenticalWorkDefinedInDifferentCompoundWorkOnlyOnce() {
        // given
        PipelineController pipelineController = new PipelineController();

        Work work1 = new Work(UUID.randomUUID().toString(), "work1", WorkContext.empty(), successfulContainerWorkAction1, Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), Collections.emptyList());

        Work compoundWork1 = new Work(UUID.randomUUID().toString(), "compoundWork1", WorkContext.empty(), new CompoundWorkDefinition(List.of(work1)), Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), Collections.emptyList());
        Work compoundWork2 = new Work(UUID.randomUUID().toString(), "compoundWork2", WorkContext.empty(), new CompoundWorkDefinition(List.of(work1)), Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), List.of(new PreviousWork(compoundWork1, null, null)));
        Pipeline pipeline = new Pipeline(
                UUID.randomUUID().toString(),
                List.of(compoundWork2),
                new PipelineMetadata("", "", SourceLanguage.JAVA)
        );
        TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();

        // when
        PipelineResult pipelineResult = pipelineController.execute(pipeline, pipelineOutputHandler).get();

        // then
        assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work1).workStatus(), equalTo(WorkStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(compoundWork1).workStatus(), equalTo(WorkStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(compoundWork2).workStatus(), equalTo(WorkStatus.SUCCESS));

        assertThat(pipelineOutputHandler.getPipelineEvents().size(), equalTo(8));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(0), equalTo(new PipelineStarted(pipeline)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(1), "");
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(2), SUCCESSFUL_TASK_1_OUTPUT);
        assertThat(pipelineOutputHandler.getPipelineEvents().get(3), equalTo(new WorkFinished(work1, WorkStatus.SUCCESS)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(4), equalTo(new WorkFinished(compoundWork1, WorkStatus.SUCCESS)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(5), "");
        assertThat(pipelineOutputHandler.getPipelineEvents().get(6), equalTo(new WorkFinished(compoundWork2, WorkStatus.SUCCESS)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(7), equalTo(new PipelineFinished(PipelineStatus.SUCCESS)));
    }

    @SneakyThrows
    @Test
    public void shouldIncludeAndOverrideContainingWorkContexts() {
        // given
        PipelineController pipelineController = new PipelineController();

        String innerWorkOverride = "innerWorkOverride";
        String innerWorkOverrideValue = UUID.randomUUID().toString();
        WorkContext innerWorkContext = WorkContext.fromMap(Map.of(
                innerWorkOverride, innerWorkOverrideValue
        ));
        String compoundWorkOverride = "compoundWorkOverride";
        String compoundWorkOverrideValue = UUID.randomUUID().toString();
        WorkContext compoundWorkContext = WorkContext.fromMap(Map.of(
                compoundWorkOverride, compoundWorkOverrideValue
        ));

        Work innerWork = new Work(UUID.randomUUID().toString(), "work1", innerWorkContext, envDumpWork, Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), Collections.emptyList());
        Work compoundWork = new Work(UUID.randomUUID().toString(), "work2", compoundWorkContext, new CompoundWorkDefinition(List.of(innerWork)), Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), Collections.emptyList());

        String pipelineContextKey = "pipelineContextKey";
        String pipelineContextValue = UUID.randomUUID().toString();

        Pipeline pipeline = new Pipeline(
                UUID.randomUUID().toString(),
                Parameters.empty(),
                WorkContext.fromMap(Map.of(
                        pipelineContextKey, pipelineContextValue,
                        innerWorkOverride, "willBeOverridden",
                        compoundWorkOverride, "willBeOverridden"
                )),
                List.of(compoundWork),
                new PipelineMetadata("", "", SourceLanguage.JAVA)
        );
        TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();

        // when
        PipelineResult pipelineResult = pipelineController.execute(pipeline, pipelineOutputHandler).get();

        // then
        assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(innerWork).workStatus(), equalTo(WorkStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(compoundWork).workStatus(), equalTo(WorkStatus.SUCCESS));

        assertThat(pipelineOutputHandler.getPipelineEvents().size(), equalTo(6));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(0), equalTo(new PipelineStarted(pipeline)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(1), "");
        assertWorkOutputContainsString(pipelineOutputHandler.getPipelineEvents().get(2), "%s=%s".formatted(innerWorkOverride, innerWorkOverrideValue));
        assertWorkOutputContainsString(pipelineOutputHandler.getPipelineEvents().get(2), "%s=%s".formatted(compoundWorkOverride, compoundWorkOverrideValue));
        assertWorkOutputContainsString(pipelineOutputHandler.getPipelineEvents().get(2), "%s=%s".formatted(pipelineContextKey, pipelineContextValue));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(3), equalTo(new WorkFinished(innerWork, WorkStatus.SUCCESS)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(4), equalTo(new WorkFinished(compoundWork, WorkStatus.SUCCESS)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(5), equalTo(new PipelineFinished(PipelineStatus.SUCCESS)));
    }

    @SneakyThrows
    @Test
    public void shouldIncludeArgumentsInPipelineWorkContext() {
        // given
        PipelineController pipelineController = new PipelineController();

        Work work = new Work(UUID.randomUUID().toString(), "work1", WorkContext.empty(), envDumpWork, Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), Collections.emptyList());

        String pipelineContextThatWillBeOverridden = "pipelineContextThatWillBeOverridden";
        String pipelineContextThatWillBeOverriddenValue = UUID.randomUUID().toString();
        String pipelineArgumentValueForOverriddenContext = UUID.randomUUID().toString();

        String pipelineContextThatWillNotBeOverridden = "pipelineContextThatWillNotBeOverridden";
        String pipelineContextThatWillNotBeOverriddenValue = UUID.randomUUID().toString();

        String pipelineArgumentThatIsNotInContext = "pipelineArgumentThatIsNotInContext";
        String pipelineArgumentThatIsNotInContextValue = UUID.randomUUID().toString();

        Pipeline pipeline = new Pipeline(
                UUID.randomUUID().toString(),
                Parameters.fromMap(Map.of(
                        pipelineContextThatWillBeOverridden, new Parameter(Parameter.TypeCheck.STRING, false, null, Collections.emptyList()),
                        pipelineArgumentThatIsNotInContext, new Parameter(Parameter.TypeCheck.STRING, false, null, Collections.emptyList())
                )),
                WorkContext.fromMap(Map.of(
                        pipelineContextThatWillBeOverridden, pipelineContextThatWillBeOverriddenValue,
                        pipelineContextThatWillNotBeOverridden, pipelineContextThatWillNotBeOverriddenValue
                )),
                List.of(work),
                new PipelineMetadata("", "", SourceLanguage.JAVA)
        );
        TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();
        Pipeline pipelineWithArguments = pipeline.withArguments(Arguments.of(
                pipelineContextThatWillBeOverridden, pipelineArgumentValueForOverriddenContext,
                pipelineArgumentThatIsNotInContext, pipelineArgumentThatIsNotInContextValue
        ));

        // when
        PipelineResult pipelineResult = pipelineController.execute(pipelineWithArguments, pipelineOutputHandler).get();

        // then
        assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work).workStatus(), equalTo(WorkStatus.SUCCESS));

        assertThat(pipelineOutputHandler.getPipelineEvents().size(), equalTo(4));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(0), equalTo(new PipelineStarted(pipelineWithArguments)));
        assertWorkOutputContainsString(pipelineOutputHandler.getPipelineEvents().get(1), "%s=%s".formatted(pipelineContextThatWillBeOverridden, pipelineArgumentValueForOverriddenContext));
        assertWorkOutputContainsString(pipelineOutputHandler.getPipelineEvents().get(1), "%s=%s".formatted(pipelineContextThatWillNotBeOverridden, pipelineContextThatWillNotBeOverriddenValue));
        assertWorkOutputContainsString(pipelineOutputHandler.getPipelineEvents().get(1), "%s=%s".formatted(pipelineArgumentThatIsNotInContext, pipelineArgumentThatIsNotInContextValue));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(2), equalTo(new WorkFinished(work, WorkStatus.SUCCESS)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(3), equalTo(new PipelineFinished(PipelineStatus.SUCCESS)));
    }

    @SneakyThrows
    @Test
    public void shouldMountOutputFileFromPreviousWork() {
        // given
        PipelineController pipelineController = new PipelineController();

        String output = UUID.randomUUID().toString();
        String outputName = "output-name";
        String outputPath = "/var/output";
        Work work1 = new Work(UUID.randomUUID().toString(), "work1", WorkContext.empty(), new ContainerisedWorkDefinition(
                TEST_IMAGE,
                List.of("bash", "-c", "echo -n '%s' > %s".formatted(output, outputPath))
        ), Conditions.PREVIOUS_WORK_SUCCESS, List.of(new WorkOutputPath(outputName, outputPath)), Collections.emptyList());
        String outputsMountPath = "/work1";
        Work work2 = new Work(UUID.randomUUID().toString(), "work2", WorkContext.empty(), new ContainerisedWorkDefinition(
                TEST_IMAGE,
                List.of("bash", "-c", "cat %s/%s".formatted(outputsMountPath, outputName))
        ), Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), List.of(new PreviousWork(work1, outputsMountPath, null)));

        Pipeline pipeline = new Pipeline(
                UUID.randomUUID().toString(),
                List.of(work2),
                new PipelineMetadata("", "", SourceLanguage.JAVA)
        );
        TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();

        // when
        PipelineResult pipelineResult = pipelineController.execute(pipeline, pipelineOutputHandler).get();

        // then
        assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work1).workStatus(), equalTo(WorkStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work2).workStatus(), equalTo(WorkStatus.SUCCESS));

        assertThat(pipelineOutputHandler.getPipelineEvents().size(), equalTo(6));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(0), equalTo(new PipelineStarted(pipeline)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(1), "");
        assertThat(pipelineOutputHandler.getPipelineEvents().get(2), equalTo(new WorkFinished(work1, WorkStatus.SUCCESS)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(3), output);
        assertThat(pipelineOutputHandler.getPipelineEvents().get(4), equalTo(new WorkFinished(work2, WorkStatus.SUCCESS)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(5), equalTo(new PipelineFinished(PipelineStatus.SUCCESS)));
    }

    @SneakyThrows
    @Test
    public void shouldMountOutputDirectoryFromPreviousWork() {
        // given
        PipelineController pipelineController = new PipelineController();

        String output = UUID.randomUUID().toString();
        String outputName = "output-name";
        String outputDirectory = "/var/output";
        String fileInOutputDirectory = "file_name";
        Work work1 = new Work(UUID.randomUUID().toString(), "work1", WorkContext.empty(), new ContainerisedWorkDefinition(
                TEST_IMAGE,
                List.of("bash", "-c", "mkdir -p %s; echo -n '%s' > %s/%s".formatted(outputDirectory, output, outputDirectory, fileInOutputDirectory))
        ), Conditions.PREVIOUS_WORK_SUCCESS, List.of(new WorkOutputPath(outputName, outputDirectory)), Collections.emptyList());
        String outputsMountPath = "/work1";
        Work work2 = new Work(UUID.randomUUID().toString(), "work2", WorkContext.empty(), new ContainerisedWorkDefinition(
                TEST_IMAGE,
                List.of("bash", "-c", "cat %s/%s/%s".formatted(outputsMountPath, outputName, fileInOutputDirectory))
        ), Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), List.of(new PreviousWork(work1, outputsMountPath, null)));

        Pipeline pipeline = new Pipeline(
                UUID.randomUUID().toString(),
                List.of(work2),
                new PipelineMetadata("", "", SourceLanguage.JAVA)
        );
        TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();

        // when
        PipelineResult pipelineResult = pipelineController.execute(pipeline, pipelineOutputHandler).get();

        // then
        assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work1).workStatus(), equalTo(WorkStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work2).workStatus(), equalTo(WorkStatus.SUCCESS));

        assertThat(pipelineOutputHandler.getPipelineEvents().size(), equalTo(6));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(0), equalTo(new PipelineStarted(pipeline)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(1), "");
        assertThat(pipelineOutputHandler.getPipelineEvents().get(2), equalTo(new WorkFinished(work1, WorkStatus.SUCCESS)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(3), output);
        assertThat(pipelineOutputHandler.getPipelineEvents().get(4), equalTo(new WorkFinished(work2, WorkStatus.SUCCESS)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(5), equalTo(new PipelineFinished(PipelineStatus.SUCCESS)));
    }

    @SneakyThrows
    @Test
    public void shouldSetStdOutEnvVarFromPreviousWork() {
        // given
        PipelineController pipelineController = new PipelineController();

        String stdOut = UUID.randomUUID().toString();
        Work work1 = new Work(UUID.randomUUID().toString(), "work1", WorkContext.empty(), new ContainerisedWorkDefinition(
                TEST_IMAGE,
                List.of("bash", "-c", "echo -n '%s'".formatted(stdOut))
        ), Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), Collections.emptyList());
        String stdOutEnvVarName = "STD_OUT_ENV_VAR";
        Work work2 = new Work(UUID.randomUUID().toString(), "work2", WorkContext.empty(), new ContainerisedWorkDefinition(
                TEST_IMAGE,
                List.of("bash", "-c", "echo -n \"${%s}\"".formatted(stdOutEnvVarName))
        ), Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), List.of(new PreviousWork(work1, null, stdOutEnvVarName)));

        Pipeline pipeline = new Pipeline(
                UUID.randomUUID().toString(),
                List.of(work2),
                new PipelineMetadata("", "", SourceLanguage.JAVA)
        );
        TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();

        // when
        PipelineResult pipelineResult = pipelineController.execute(pipeline, pipelineOutputHandler).get();

        // then
        assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work1).workStatus(), equalTo(WorkStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work2).workStatus(), equalTo(WorkStatus.SUCCESS));

        assertThat(pipelineOutputHandler.getPipelineEvents().size(), equalTo(6));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(0), equalTo(new PipelineStarted(pipeline)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(2), equalTo(new WorkFinished(work1, WorkStatus.SUCCESS)));
        assertWorkOutputOneOf(pipelineOutputHandler.getPipelineEvents().get(3), stdOut);
        assertThat(pipelineOutputHandler.getPipelineEvents().get(4), equalTo(new WorkFinished(work2, WorkStatus.SUCCESS)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(5), equalTo(new PipelineFinished(PipelineStatus.SUCCESS)));
    }

    @SneakyThrows
    @Test
    public void shouldRunWorkIfComplexConditionIsMet() {
        // given
        PipelineController pipelineController = new PipelineController();

        String workContextKey = UUID.randomUUID().toString();
        String workContextValue = UUID.randomUUID().toString();
        Condition complexCondition = new AndCondition(
            new WorkContextCondition(workContextKey, WorkContextCondition.Operand.EQUALS, workContextValue),
            new NotCondition(
                new OrCondition(
                    new WorkContextCondition("missing1", WorkContextCondition.Operand.EQUALS, ""),
                    new WorkContextCondition("missing2", WorkContextCondition.Operand.EQUALS, "")
                )
            )
        );
        Work work1 = new Work(UUID.randomUUID().toString(), "work1", WorkContext.fromMap(Map.of(workContextKey, workContextValue)), successfulContainerWorkAction1,
            complexCondition, Collections.emptyList(), Collections.emptyList());

        Pipeline pipeline = new Pipeline(
                UUID.randomUUID().toString(),
                List.of(work1),
                new PipelineMetadata("", "", SourceLanguage.JAVA)
        );
        TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();

        // when
        PipelineResult pipelineResult = pipelineController.execute(pipeline, pipelineOutputHandler).get();

        // then
        assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work1).workStatus(), equalTo(WorkStatus.SUCCESS));

        assertThat(pipelineOutputHandler.getPipelineEvents().size(), equalTo(4));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(0), equalTo(new PipelineStarted(pipeline)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(2), equalTo(new WorkFinished(work1, WorkStatus.SUCCESS)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(3), equalTo(new PipelineFinished(PipelineStatus.SUCCESS)));
    }

    @SneakyThrows
    @Test
    public void shouldNotRunWorkIfConditionNotMet() {
        // given
        PipelineController pipelineController = new PipelineController();

        String workContextKey = UUID.randomUUID().toString();
        String workContextValue = UUID.randomUUID().toString();
        Work work1 = new Work(UUID.randomUUID().toString(), "work1", WorkContext.fromMap(Map.of(workContextKey, workContextValue)), successfulContainerWorkAction1,
            new NotCondition(new WorkContextCondition(workContextKey, WorkContextCondition.Operand.EQUALS, workContextValue)), Collections.emptyList(), Collections.emptyList());

        Pipeline pipeline = new Pipeline(
                UUID.randomUUID().toString(),
                List.of(work1),
                new PipelineMetadata("", "", SourceLanguage.JAVA)
        );
        TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();

        // when
        PipelineResult pipelineResult = pipelineController.execute(pipeline, pipelineOutputHandler).get();

        // then
        assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.SUCCESS));
        assertThat(pipelineResult.workResults().get(work1).workStatus(), equalTo(WorkStatus.SKIPPED));

        assertThat(pipelineOutputHandler.getPipelineEvents().size(), equalTo(3));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(0), equalTo(new PipelineStarted(pipeline)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(1), equalTo(new WorkFinished(work1, WorkStatus.SKIPPED)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(2), equalTo(new PipelineFinished(PipelineStatus.SUCCESS)));
    }

    @SneakyThrows
    @Test
    public void shouldRunWorkAfterFailureIfConditionIsMet() {
        // given
        PipelineController pipelineController = new PipelineController();

        String workContextKey = UUID.randomUUID().toString();
        String workContextValue = UUID.randomUUID().toString();
        Work work1 = new Work(UUID.randomUUID().toString(), "work1", WorkContext.fromMap(Map.of(workContextKey, workContextValue)), failedContainerWorkAction,
            Conditions.PREVIOUS_WORK_SUCCESS, Collections.emptyList(), Collections.emptyList());
        Work work2 = new Work(UUID.randomUUID().toString(), "work2", WorkContext.fromMap(Map.of(workContextKey, workContextValue)), successfulContainerWorkAction1,
            new PreviousWorkStatusCondition(PreviousWorkStatusCondition.Status.FAILURE), Collections.emptyList(), List.of(new PreviousWork(work1, null, null)));

        Pipeline pipeline = new Pipeline(
                UUID.randomUUID().toString(),
                List.of(work2),
                new PipelineMetadata("", "", SourceLanguage.JAVA)
        );
        TestPipelineOutputHandler pipelineOutputHandler = new TestPipelineOutputHandler();

        // when
        PipelineResult pipelineResult = pipelineController.execute(pipeline, pipelineOutputHandler).get();

        // then
        assertThat(pipelineResult.pipelineStatus(), equalTo(PipelineStatus.FAILURE));
        assertThat(pipelineResult.workResults().get(work1).workStatus(), equalTo(WorkStatus.FAILURE));
        assertThat(pipelineResult.workResults().get(work2).workStatus(), equalTo(WorkStatus.SUCCESS));

        assertThat(pipelineOutputHandler.getPipelineEvents().size(), equalTo(6));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(0), equalTo(new PipelineStarted(pipeline)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(2), equalTo(new WorkFinished(work1, WorkStatus.FAILURE)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(4), equalTo(new WorkFinished(work2, WorkStatus.SUCCESS)));
        assertThat(pipelineOutputHandler.getPipelineEvents().get(5), equalTo(new PipelineFinished(PipelineStatus.FAILURE)));
    }

    private static void assertWorkOutputOneOf(PipelineEvent pipelineEvent, String... workOutputs) {
        assertThat(pipelineEvent, instanceOf(WorkStarted.class));
        WorkStarted workStarted = (WorkStarted) pipelineEvent;
        byte[] workStdOut = workStarted.jobStreams().getStdOut();
        assertThat(new String(workStdOut, StandardCharsets.UTF_8), anyOf(Arrays.stream(workOutputs).map(CoreMatchers::equalTo).collect(Collectors.toList())));
    }

    private static void assertWorkOutputContainsString(PipelineEvent pipelineEvent, String subString) {
        assertThat(pipelineEvent, instanceOf(WorkStarted.class));
        WorkStarted workStarted = (WorkStarted) pipelineEvent;
        byte[] workStdOut = workStarted.jobStreams().getStdOut();
        assertThat(new String(workStdOut, StandardCharsets.UTF_8), containsString(subString));
    }
}
