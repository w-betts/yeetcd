package yeetcd.test;

import yeetcd.sdk.*;
import yeetcd.sdk.condition.Conditions;
import yeetcd.sdk.condition.PreviousWorkStatusCondition;
import yeetcd.sdk.condition.WorkContextCondition;
import lombok.SneakyThrows;
import org.junit.jupiter.api.Test;

import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.*;
import java.util.stream.IntStream;
import java.util.stream.Stream;

import static org.hamcrest.CoreMatchers.equalTo;
import static org.hamcrest.MatcherAssert.assertThat;

public class FakePipelineRunnerTest {

    @Test
    public void shouldWorkForASimplePipeline() {
        // given
        Work work1 = Work
            .builder("work1", ContainerisedWorkDefinition.builder("test").build())
            .build();
        Work work2 = Work
            .builder("work2", ContainerisedWorkDefinition.builder("test").build())
            .previousWork(PreviousWork.builder(work1).build())
            .build();

        Pipeline pipeline = Pipeline
            .builder("test")
            .finalWork(work2)
            .build();

        // and
        FakePipelineRunner fakePipelineRunner = FakePipelineRunner
            .builder()
            .build();

        // when
        FakePipelineRunResult result = fakePipelineRunner.run(FakePipelineRun.builder(pipeline).build());

        // then
        assertThat(result, equalTo(
            FakePipelineRunResult
                .builder(List.of(
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work1)
                                .build()
                        ))
                        .build(),
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work2)
                                .build()
                        ))
                        .build()
                ))
                .status(FakePipelineStatus.SUCCESS)
                .build()
        ));
    }

    @Test
    public void shouldMatchWorkStatusFailureOverrideAndSkippingDependentWork() {
        // given
        Work work1 = Work
            .builder("work1", ContainerisedWorkDefinition.builder("test").build())
            .build();
        Work work2 = Work
            .builder("work2", ContainerisedWorkDefinition.builder("test").build())
            .previousWork(PreviousWork.builder(work1).build())
            .build();

        Pipeline pipeline = Pipeline
            .builder("test")
            .finalWork(work2)
            .build();

        // and
        FakePipelineRunner fakePipelineRunner = FakePipelineRunner
            .builder()
            .specifiedWorkResults(List.of(
                FakeWorkMatcherResult
                    .builder(
                        FakeWorkMatcher
                            .builder()
                            .work(work1)
                            .build(),
                        FakeWorkResult
                            .builder()
                            .status(FakeWorkStatus.FAILURE)
                            .build()
                    )
                    .build()
            ))
            .build();

        // when
        FakePipelineRunResult result = fakePipelineRunner.run(FakePipelineRun.builder(pipeline).build());

        // then
        assertThat(result, equalTo(
            FakePipelineRunResult
                .builder(List.of(
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work1)
                                .status(FakeWorkStatus.FAILURE)
                                .build()
                        ))
                        .build(),
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work2)
                                .status(FakeWorkStatus.SKIPPED)
                                .build()
                        ))
                        .build()
                ))
                .status(FakePipelineStatus.FAILURE)
                .build()
        ));
    }

    @Test
    public void shouldMatchWorkStatusFailureOverrideWithSpecificWorkContext() {
        // given
        Work work1 = Work
            .builder("work1", ContainerisedWorkDefinition.builder("test").build())
            .build();


        String work2WorkContextKey = "work2-key";
        String work2WorkContextValue = "work2-value";
        Work work2 = Work
            .builder("work2", ContainerisedWorkDefinition.builder("test").build())
            .workContext(WorkContext.of(work2WorkContextKey, work2WorkContextValue))
            .previousWork(PreviousWork.builder(work1).build())
            .build();

        String pipelineWorkContextKey = "pipeline-key";
        String pipelineWorkContextValue = "pipeline-value";
        String parameterKey = "parameter-key";
        String argumentValue = "argument-value";
        Pipeline pipeline = Pipeline
            .builder("test")
            .workContext(WorkContext.of(pipelineWorkContextKey, pipelineWorkContextValue))
            .parameters(Parameters.of(parameterKey, Parameter.builder(Parameter.TypeCheck.STRING).build()))
            .finalWork(work2)
            .build();

        // and
        FakePipelineRunner fakePipelineRunner = FakePipelineRunner
            .builder()
            .specifiedWorkResults(List.of(
                FakeWorkMatcherResult
                    .builder(
                        FakeWorkMatcher
                            .builder()
                            .work(work2)
                            .workContext(WorkContext.of(
                                pipelineWorkContextKey, pipelineWorkContextValue,
                                work2WorkContextKey, work2WorkContextValue,
                                parameterKey, argumentValue
                            ))
                            .build(),
                        FakeWorkResult
                            .builder()
                            .status(FakeWorkStatus.FAILURE)
                            .build()
                    )
                    .build()
            ))
            .build();

        // when
        FakePipelineRunResult result = fakePipelineRunner.run(FakePipelineRun
            .builder(pipeline)
            .arguments(Map.of(parameterKey, argumentValue))
            .build()
        );

        // then
        assertThat(result, equalTo(
            FakePipelineRunResult
                .builder(List.of(
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work1)
                                .envVars(Map.of(
                                    pipelineWorkContextKey, pipelineWorkContextValue,
                                    parameterKey, argumentValue
                                ))
                                .status(FakeWorkStatus.SUCCESS)
                                .build()
                        ))
                        .build(),
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work2)
                                .envVars(Map.of(
                                    work2WorkContextKey, work2WorkContextValue,
                                    pipelineWorkContextKey, pipelineWorkContextValue,
                                    parameterKey, argumentValue
                                ))
                                .status(FakeWorkStatus.FAILURE)
                                .build()
                        ))
                        .build()
                ))
                .status(FakePipelineStatus.FAILURE)
                .build()
        ));
    }

    @Test
    public void shouldDoAFanOut() {
        // given
        Work work1 = Work
            .builder("work1", ContainerisedWorkDefinition.builder("test").build())
            .build();

        Work work2 = Work
            .builder("work2", ContainerisedWorkDefinition.builder("test").build())
            .previousWork(PreviousWork.builder(work1).build())
            .build();

        Work work3 = Work
            .builder("work3", ContainerisedWorkDefinition.builder("test").build())
            .previousWork(PreviousWork.builder(work1).build())
            .build();

        Pipeline pipeline = Pipeline
            .builder("test")
            .finalWork(work2, work3)
            .build();

        // and
        FakePipelineRunner fakePipelineRunner = FakePipelineRunner
            .builder()
            .build();

        // when
        FakePipelineRunResult result = fakePipelineRunner.run(FakePipelineRun
            .builder(pipeline)
            .build()
        );

        // then
        assertThat(result, equalTo(
            FakePipelineRunResult
                .builder(List.of(
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work1)
                                .status(FakeWorkStatus.SUCCESS)
                                .build()
                        ))
                        .build(),
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work2)
                                .status(FakeWorkStatus.SUCCESS)
                                .build(),
                            FakeSimpleWorkExecution
                                .builder(work3)
                                .status(FakeWorkStatus.SUCCESS)
                                .build()
                        ))
                        .build()
                ))
                .status(FakePipelineStatus.SUCCESS)
                .build()
        ));
    }

    @Test
    public void shouldDoAFanIn() {
        // given
        Work work1 = Work
            .builder("work1", ContainerisedWorkDefinition.builder("test").build())
            .build();

        Work work2 = Work
            .builder("work2", ContainerisedWorkDefinition.builder("test").build())
            .build();

        Work work3 = Work
            .builder("work3", ContainerisedWorkDefinition.builder("test").build())
            .previousWork(
                PreviousWork.builder(work1).build(),
                PreviousWork.builder(work2).build()
            )
            .build();

        Pipeline pipeline = Pipeline
            .builder("test")
            .finalWork(work2, work3)
            .build();

        // and
        FakePipelineRunner fakePipelineRunner = FakePipelineRunner
            .builder()
            .build();

        // when
        FakePipelineRunResult result = fakePipelineRunner.run(FakePipelineRun
            .builder(pipeline)
            .build()
        );

        // then
        assertThat(result, equalTo(
            FakePipelineRunResult
                .builder(List.of(
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work1)
                                .status(FakeWorkStatus.SUCCESS)
                                .build(),
                            FakeSimpleWorkExecution
                                .builder(work2)
                                .status(FakeWorkStatus.SUCCESS)
                                .build()
                        ))
                        .build(),
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work3)
                                .status(FakeWorkStatus.SUCCESS)
                                .build()
                        ))
                        .build()
                ))
                .status(FakePipelineStatus.SUCCESS)
                .build()
        ));
    }

    @Test
    public void shouldTreatDifferentInstancesOfEqualWorkAsIdentical() {
        // given
        Work work1a = Work
            .builder("work1", ContainerisedWorkDefinition.builder("test").build())
            .build();
        Work work1b = Work
            .builder("work1", ContainerisedWorkDefinition.builder("test").build())
            .build();

        Work work2 = Work
            .builder("work2", ContainerisedWorkDefinition.builder("test").build())
            .previousWork(
                PreviousWork.builder(work1a).build()
            )
            .build();

        Work work3 = Work
            .builder("work3", ContainerisedWorkDefinition.builder("test").build())
            .previousWork(
                PreviousWork.builder(work1b).build()
            )
            .build();

        Pipeline pipeline = Pipeline
            .builder("test")
            .finalWork(work2, work3)
            .build();

        // and
        FakePipelineRunner fakePipelineRunner = FakePipelineRunner
            .builder()
            .build();

        // when
        FakePipelineRunResult result = fakePipelineRunner.run(FakePipelineRun
            .builder(pipeline)
            .build()
        );

        // then
        assertThat(result, equalTo(
            FakePipelineRunResult
                .builder(List.of(
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work1a)
                                .status(FakeWorkStatus.SUCCESS)
                                .build()
                        ))
                        .build(),
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work2)
                                .status(FakeWorkStatus.SUCCESS)
                                .build(),
                            FakeSimpleWorkExecution
                                .builder(work3)
                                .status(FakeWorkStatus.SUCCESS)
                                .build()
                        ))
                        .build()
                ))
                .status(FakePipelineStatus.SUCCESS)
                .build()
        ));
    }

    @Test
    public void shouldRunCompoundWork() {
        // given
        Work innerWork1 = Work
            .builder("innerWork1", ContainerisedWorkDefinition.builder("test").build())
            .build();

        Work innerWork2 = Work
            .builder("innerWork2", ContainerisedWorkDefinition.builder("test").build())
            .previousWork(PreviousWork.builder(innerWork1).build())
            .build();

        Work compoundWork = Work
            .builder("compoundWork", CompoundWorkDefinition.builder(innerWork2).build())
            .build();

        Pipeline pipeline = Pipeline
            .builder("test")
            .finalWork(compoundWork)
            .build();

        // and
        FakePipelineRunner fakePipelineRunner = FakePipelineRunner
            .builder()
            .build();

        // when
        FakePipelineRunResult result = fakePipelineRunner.run(FakePipelineRun.builder(pipeline).build());

        // then
        assertThat(result, equalTo(
            FakePipelineRunResult
                .builder(List.of(
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeCompoundWorkExecution
                                .builder(
                                    FakeWorkExecutionStage
                                        .builder(Set.of(
                                            FakeSimpleWorkExecution
                                                .builder(innerWork1)
                                                .build()
                                        ))
                                        .build(),
                                    FakeWorkExecutionStage
                                        .builder(Set.of(
                                            FakeSimpleWorkExecution
                                                .builder(innerWork2)
                                                .build()
                                        ))
                                        .build()
                                )
                                .build()
                        ))
                        .build()
                ))
                .status(FakePipelineStatus.SUCCESS)
                .build()
        ));
    }

    @Test
    public void shouldRunEqualWorkInDifferentCompoundWorkOnlyOnceIfContextIsSame() {
        // given
        Work innerWork1a = Work
            .builder("innerWork1", ContainerisedWorkDefinition.builder("test").build())
            .build();

        Work innerWork1b = Work
            .builder("innerWork1", ContainerisedWorkDefinition.builder("test").build())
            .build();

        Work compoundWork1 = Work
            .builder("compoundWork1", CompoundWorkDefinition.builder(innerWork1a).build())
            .build();

        Work compoundWork2 = Work
            .builder("compoundWork2", CompoundWorkDefinition.builder(innerWork1b).build())
            .previousWork(PreviousWork.builder(compoundWork1).build())
            .build();

        Pipeline pipeline = Pipeline
            .builder("test")
            .finalWork(compoundWork2)
            .build();

        // and
        FakePipelineRunner fakePipelineRunner = FakePipelineRunner
            .builder()
            .build();

        // when
        FakePipelineRunResult result = fakePipelineRunner.run(FakePipelineRun.builder(pipeline).build());

        // then
        assertThat(result, equalTo(
            FakePipelineRunResult
                .builder(List.of(
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeCompoundWorkExecution
                                .builder(
                                    FakeWorkExecutionStage
                                        .builder(Set.of(
                                            FakeSimpleWorkExecution
                                                .builder(innerWork1a)
                                                .build()
                                        ))
                                        .build()
                                )
                                .build()
                        ))
                        .build(),
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeCompoundWorkExecution
                                .builder()
                                .build()
                        ))
                        .build()
                ))
                .status(FakePipelineStatus.SUCCESS)
                .build()
        ));
    }

    @Test
    public void shouldRunEqualWorkInDifferentCompoundWorkOnceEachIfContextIsDifferent() {
        // given
        Work innerWork1a = Work
            .builder("innerWork1", ContainerisedWorkDefinition.builder("test").build())
            .build();

        Work innerWork1b = Work
            .builder("innerWork1", ContainerisedWorkDefinition.builder("test").build())
            .build();

        String workContextKey = "key";
        String workContextValue1 = "value1";
        Work compoundWork1 = Work
            .builder("compoundWork1", CompoundWorkDefinition.builder(innerWork1a).build())
            .workContext(WorkContext.of(workContextKey, workContextValue1))
            .build();

        String workContextValue2 = "value2";
        Work compoundWork2 = Work
            .builder("compoundWork2", CompoundWorkDefinition.builder(innerWork1b).build())
            .workContext(WorkContext.of(workContextKey, workContextValue2))
            .previousWork(PreviousWork.builder(compoundWork1).build())
            .build();

        Pipeline pipeline = Pipeline
            .builder("test")
            .finalWork(compoundWork2)
            .build();

        // and
        FakePipelineRunner fakePipelineRunner = FakePipelineRunner
            .builder()
            .build();

        // when
        FakePipelineRunResult result = fakePipelineRunner.run(FakePipelineRun.builder(pipeline).build());

        // then
        assertThat(result, equalTo(
            FakePipelineRunResult
                .builder(List.of(
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeCompoundWorkExecution
                                .builder(
                                    FakeWorkExecutionStage
                                        .builder(Set.of(
                                            FakeSimpleWorkExecution
                                                .builder(innerWork1a)
                                                .envVars(Map.of(workContextKey, workContextValue1))
                                                .build()
                                        ))
                                        .build()
                                )
                                .build()
                        ))
                        .build(),
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeCompoundWorkExecution
                                .builder(
                                    FakeWorkExecutionStage
                                        .builder(Set.of(
                                            FakeSimpleWorkExecution
                                                .builder(innerWork1b)
                                                .envVars(Map.of(workContextKey, workContextValue2))
                                                .build()
                                        ))
                                        .build()
                                )
                                .build()
                        ))
                        .build()
                ))
                .status(FakePipelineStatus.SUCCESS)
                .build()
        ));
    }

    @Test
    public void shouldIncludeOutputFilesAsInputs() {
        // given
        String outputName = "outputName";
        String outputPath = "/outputPath";
        byte[] outputValue = "outputValue".getBytes(StandardCharsets.UTF_8);
        String mountPath = "/mountPath";

        Work work1 = Work
            .builder("work1", ContainerisedWorkDefinition.builder("test").build())
            .workOutputPaths(WorkOutputPath.builder(outputName, outputPath).build())
            .build();
        Work work2 = Work
            .builder("work2", ContainerisedWorkDefinition.builder("test").build())
            .previousWork(PreviousWork.builder(work1).outputsMountPath(mountPath).build())
            .build();

        Pipeline pipeline = Pipeline
            .builder("test")
            .finalWork(work2)
            .build();


        // and
        FakePipelineRunner fakePipelineRunner = FakePipelineRunner
            .builder()
            .specifiedWorkResults(
                FakeWorkMatcherResult
                    .builder(
                        FakeWorkMatcher.builder().work(work1).build(),
                        FakeWorkResult.builder().exportedFiles(Map.of(outputPath, outputValue)).build()
                    )
                    .build()
            )
            .build();

        // when
        FakePipelineRunResult result = fakePipelineRunner.run(FakePipelineRun.builder(pipeline).build());

        // then
        assertThat(result, equalTo(
            FakePipelineRunResult
                .builder(List.of(
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work1)
                                .exportedFiles(Map.of(outputPath, outputValue))
                                .build()
                        ))
                        .build(),
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work2)
                                .inputFiles(Map.of(
                                    "%s/%s".formatted(mountPath, outputName), outputValue
                                ))
                                .build()
                        ))
                        .build()
                ))
                .status(FakePipelineStatus.SUCCESS)
                .build()
        ));
    }

    @Test
    public void shouldIncludeStdOutAsEnvVar() {
        // given
        String stdOut = "output";
        String stdOutEnvVarName = "previousWorkOutput";

        Work work1 = Work
            .builder("work1", ContainerisedWorkDefinition.builder("test").build())
            .build();
        Work work2 = Work
            .builder("work2", ContainerisedWorkDefinition.builder("test").build())
            .previousWork(PreviousWork.builder(work1).stdOutEnvVar(stdOutEnvVarName).build())
            .build();

        Pipeline pipeline = Pipeline
            .builder("test")
            .finalWork(work2)
            .build();


        // and
        FakePipelineRunner fakePipelineRunner = FakePipelineRunner
            .builder()
            .specifiedWorkResults(
                FakeWorkMatcherResult
                    .builder(
                        FakeWorkMatcher.builder().work(work1).build(),
                        FakeWorkResult.builder().stdOut(stdOut).build()
                    )
                    .build()
            )
            .build();

        // when
        FakePipelineRunResult result = fakePipelineRunner.run(FakePipelineRun.builder(pipeline).build());

        // then
        assertThat(result, equalTo(
            FakePipelineRunResult
                .builder(List.of(
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work1)
                                .stdOut(stdOut)
                                .build()
                        ))
                        .build(),
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work2)
                                .envVars(Map.of(stdOutEnvVarName, stdOut))
                                .build()
                        ))
                        .build()
                ))
                .status(FakePipelineStatus.SUCCESS)
                .build()
        ));
    }

    @Test
    public void shouldNotRunWorkIfConditionIsNotMet() {
        // given

        Work work1 = Work
            .builder("work1", ContainerisedWorkDefinition.builder("test").build())
            .build();
        Work work2 = Work
            .builder("work2", ContainerisedWorkDefinition.builder("test").build())
            .condition(Conditions.workContextCondition("missingKey", WorkContextCondition.Operand.EQUALS, "value"))
            .previousWork(PreviousWork.builder(work1).build())
            .build();

        Pipeline pipeline = Pipeline
            .builder("test")
            .finalWork(work2)
            .build();


        // and
        FakePipelineRunner fakePipelineRunner = FakePipelineRunner
            .builder()
            .build();

        // when
        FakePipelineRunResult result = fakePipelineRunner.run(FakePipelineRun.builder(pipeline).build());

        // then
        assertThat(result, equalTo(
            FakePipelineRunResult
                .builder(List.of(
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work1)
                                .build()
                        ))
                        .build(),
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work2)
                                .status(FakeWorkStatus.SKIPPED)
                                .build()
                        ))
                        .build()
                ))
                .status(FakePipelineStatus.SUCCESS)
                .build()
        ));
    }

    @Test
    public void shouldRunWorkIfComplexConditionIsMetDespitePreviousFailure() {
        // given
        Work work1 = Work
            .builder("work1", ContainerisedWorkDefinition.builder("test").build())
            .build();
        String workContextKey = "key";
        String workContextValue = "value";
        Work work2 = Work
            .builder("work2", ContainerisedWorkDefinition.builder("test").build())
            .condition(
                Conditions.and(
                    Conditions.not(
                        Conditions.or(
                            Conditions.workContextCondition("missing1", WorkContextCondition.Operand.EQUALS, "value"),
                            Conditions.workContextCondition("missing2", WorkContextCondition.Operand.EQUALS, "value")
                        )
                    ),
                    Conditions.and(
                        Conditions.previousWorkStatusCondition(PreviousWorkStatusCondition.Status.ANY),
                        Conditions.workContextCondition(workContextKey, WorkContextCondition.Operand.EQUALS, workContextValue)
                    )
                ))
            .previousWork(PreviousWork.builder(work1).build())
            .build();

        Pipeline pipeline = Pipeline
            .builder("test")
            .parameters(Parameters.of(workContextKey, Parameter.builder(Parameter.TypeCheck.STRING).build()))
            .finalWork(work2)
            .build();


        // and
        FakePipelineRunner fakePipelineRunner = FakePipelineRunner
            .builder()
            .specifiedWorkResults(FakeWorkMatcherResult
                .builder(
                    FakeWorkMatcher.builder().work(work1).build(),
                    FakeWorkResult.builder().status(FakeWorkStatus.FAILURE).build()
                )
                .build()
            )
            .build();

        // when
        FakePipelineRunResult result = fakePipelineRunner.run(FakePipelineRun
            .builder(pipeline)
            .arguments(Map.of(workContextKey, workContextValue))
            .build()
        );

        // then
        assertThat(result, equalTo(
            FakePipelineRunResult
                .builder(List.of(
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work1)
                                .status(FakeWorkStatus.FAILURE)
                                .envVars(Map.of(
                                    workContextKey, workContextValue
                                ))
                                .build()
                        ))
                        .build(),
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work2)
                                .status(FakeWorkStatus.SUCCESS)
                                .envVars(Map.of(
                                    workContextKey, workContextValue
                                ))
                                .build()
                        ))
                        .build()
                ))
                .status(FakePipelineStatus.FAILURE)
                .build()
        ));
    }

    @Test
    public void shouldDynamicallyGenerateWork() {
        // given
        int workCount = 2;
        Work[] predefinedWorks = IntStream.range(0, workCount).mapToObj(index -> Work.builder("dynamicWork%d".formatted(index), ContainerisedWorkDefinition.builder("test").build()).build()).toArray(Work[]::new);

        Work dynamicWorkGenerator = Work
            .builder("dynamicWorkGenerator", new DynamicWorkGeneratingWorkDefinition() {
                @Override
                public Work createWork() {
                    return Work.builder("dynamicCompoundWork", CompoundWorkDefinition.builder(predefinedWorks).build()).build();
                }

                @Override
                protected Stream<CustomWorkDefinition> dynamicCustomWorkDefinitions() {
                    return Stream.of();
                }
            })
            .build();

        Pipeline pipeline = Pipeline
            .builder("test")
            .finalWork(dynamicWorkGenerator)
            .build();


        // and
        FakePipelineRunner fakePipelineRunner = FakePipelineRunner
            .builder()
            .build();

        // when
        FakePipelineRunResult result = fakePipelineRunner.run(FakePipelineRun
            .builder(pipeline)
            .build()
        );

        // then
        assertThat(result, equalTo(
            FakePipelineRunResult
                .builder(List.of(
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeCompoundWorkExecution
                                .builder(
                                    FakeWorkExecutionStage
                                        .builder(Set.of(
                                            FakeSimpleWorkExecution
                                                .builder(predefinedWorks[0])
                                                .build(),
                                            FakeSimpleWorkExecution
                                                .builder(predefinedWorks[1])
                                                .build()
                                        ))
                                        .build()
                                )
                                .build()
                        ))
                        .build()
                ))
                .status(FakePipelineStatus.SUCCESS)
                .build()
        ));
    }

    @SneakyThrows
    @Test
    public void shouldDynamicallyGenerateWorkBasedOnPreviousWorkAndWorkContextInputs() {
        // given
        Path yeetcdTestDir = Files.createTempDirectory("yeetcd_test");
        yeetcdTestDir.toFile().deleteOnExit();

        String workContextKey = "workContextKey";
        String workContextValue = "workContextValue";

        String stdOutKey = "stdOutKey";
        String stdOutValue = "stdOutValue";

        String outputName = "outputName";
        String outputPath = "/outputPath";
        byte[] outputValue = "outputValue".getBytes(StandardCharsets.UTF_8);
        String mountPath = "%s/mountPath".formatted(yeetcdTestDir.toString());

        Work workConditionalOnWorkContext = Work.builder("workConditionalOnWorkContext", ContainerisedWorkDefinition.builder("test").build()).build();
        Work workConditionalOnStdOutInput = Work.builder("workConditionalOnStdOutInput", ContainerisedWorkDefinition.builder("test").build()).build();
        Work workConditionalOnInputFile = Work.builder("workConditionalOnInputFile", ContainerisedWorkDefinition.builder("test").build()).build();

        Work work1 = Work
            .builder("work1", ContainerisedWorkDefinition.builder("test").build())
            .workOutputPaths(WorkOutputPath.builder(outputName, outputPath).build())
            .build();

        Work dynamicWorkGenerator = Work
            .builder("dynamicWorkGenerator", new DynamicWorkGeneratingWorkDefinition() {
                @Override
                @SneakyThrows
                public Work createWork() {
                    List<Work> work = new LinkedList<>();
                    if (Objects.equals(workContextValue(workContextKey), workContextValue)) {
                        work.add(workConditionalOnWorkContext);
                    }
                    if (Objects.equals(workContextValue(stdOutKey), (stdOutValue))) {
                        work.add(workConditionalOnStdOutInput);
                    }
                    if (Arrays.equals(Files.readAllBytes(Path.of(mountPath, outputName)), outputValue)) {
                        work.add(workConditionalOnInputFile);
                    }
                    return Work.builder("dynamicCompoundWork", CompoundWorkDefinition.builder(work.toArray(Work[]::new)).build()).build();
                }

                @Override
                protected Stream<CustomWorkDefinition> dynamicCustomWorkDefinitions() {
                    return Stream.of();
                }
            })
            .workContext(WorkContext.of(workContextKey, workContextValue))
            .previousWork(PreviousWork.builder(work1).stdOutEnvVar(stdOutKey).outputsMountPath(mountPath).build())
            .build();

        Pipeline pipeline = Pipeline
            .builder("test")
            .finalWork(dynamicWorkGenerator)
            .build();


        // and
        FakePipelineRunner fakePipelineRunner = FakePipelineRunner
            .builder()
            .specifiedWorkResults(
                FakeWorkMatcherResult
                    .builder(
                        FakeWorkMatcher.builder().work(work1).build(),
                        FakeWorkResult.builder()
                            .exportedFiles(Map.of(outputPath, outputValue))
                            .stdOut(stdOutValue)
                            .build()
                    )
                    .build()
            )
            .build();

        // when
        FakePipelineRunResult result = fakePipelineRunner.run(FakePipelineRun
            .builder(pipeline)
            .build()
        );

        // then
        assertThat(result, equalTo(
            FakePipelineRunResult
                .builder(List.of(
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeSimpleWorkExecution
                                .builder(work1)
                                .exportedFiles(Map.of(outputPath, outputValue))
                                .stdOut(stdOutValue)
                                .build()
                        ))
                        .build(),
                    FakeWorkExecutionStage
                        .builder(Set.of(
                            FakeCompoundWorkExecution
                                .builder(
                                    FakeWorkExecutionStage
                                        .builder(Set.of(
                                            FakeSimpleWorkExecution
                                                .builder(workConditionalOnWorkContext)
                                                .build(),
                                            FakeSimpleWorkExecution
                                                .builder(workConditionalOnStdOutInput)
                                                .build(),
                                            FakeSimpleWorkExecution
                                                .builder(workConditionalOnInputFile)
                                                .build()
                                        ))
                                        .build()
                                )
                                .build()
                        ))
                        .build()
                ))
                .status(FakePipelineStatus.SUCCESS)
                .build()
        ));
    }
}
