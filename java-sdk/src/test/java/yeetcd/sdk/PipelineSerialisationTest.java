package yeetcd.sdk;

import yeetcd.protocol.pipeline.PipelineOuterClass;
import yeetcd.sdk.condition.PreviousWorkStatusCondition;
import yeetcd.sdk.condition.WorkContextCondition;
import org.junit.jupiter.api.Test;

import java.util.Collections;
import java.util.List;
import java.util.UUID;

import static yeetcd.sdk.condition.Conditions.*;
import static org.hamcrest.CoreMatchers.*;
import static org.hamcrest.MatcherAssert.assertThat;

public class PipelineSerialisationTest {

    @Test
    public void shouldSerialiseEmptyPipeline() {
        // given
        Pipeline pipeline = Pipeline.builder("name").build();

        // when
        PipelineOuterClass.Pipeline pipelineProtobuf = pipeline.toProtobuf();

        // then
        assertThat(pipelineProtobuf.getFinalWorkList().size(), equalTo(0));
    }

    @Test
    public void shouldSerialisePipelineWithParametersWithNoDefaultOrChoices() {
        // given
        String parameterName = "parameter1";
        Parameter parameter = Parameter.builder(Parameter.TypeCheck.STRING).required(true).build();
        Pipeline pipeline = Pipeline
            .builder("name")
            .parameters(Parameters.of(parameterName, parameter))
            .build();

        // when
        PipelineOuterClass.Pipeline pipelineProtobuf = pipeline.toProtobuf();

        // then
        assertThat(pipelineProtobuf.getParametersCount(), equalTo(1));
        PipelineOuterClass.Parameter parameter1 = pipelineProtobuf.getParametersMap().get(parameterName);
        assertThat(parameter1, notNullValue());
        assertThat(parameter1.getTypeCheck(), equalTo(PipelineOuterClass.Parameter.TYPE_CHECK.STRING));
        assertThat(parameter1.getRequired(), equalTo(true));
        assertThat(parameter1.hasDefaultValue(), equalTo(false));
        assertThat(parameter1.getChoicesList().stream().toList(), equalTo(Collections.emptyList()));
    }

    @Test
    public void shouldSerialisePipelineWithParameters() {
        // given
        String parameterName = "parameter1";
        String defaultValue = "value1";
        List<String> choices = List.of("value1", "value2");
        Parameter parameter = Parameter.builder(Parameter.TypeCheck.STRING).required(true).defaultValue(defaultValue).choices(choices).build();
        Pipeline pipeline = Pipeline
            .builder("name")
            .parameters(Parameters.of(parameterName, parameter))
            .build();

        // when
        PipelineOuterClass.Pipeline pipelineProtobuf = pipeline.toProtobuf();

        // then
        assertThat(pipelineProtobuf.getParametersCount(), equalTo(1));
        PipelineOuterClass.Parameter parameter1 = pipelineProtobuf.getParametersMap().get(parameterName);
        assertThat(parameter1, notNullValue());
        assertThat(parameter1.getTypeCheck(), equalTo(PipelineOuterClass.Parameter.TYPE_CHECK.STRING));
        assertThat(parameter1.getRequired(), equalTo(true));
        assertThat(parameter1.getDefaultValue(), equalTo(defaultValue));
        assertThat(parameter1.getChoicesList().stream().toList(), equalTo(choices));
    }

    @Test
    public void shouldSerialisePipelineWithWorkContext() {
        // given
        String workContextKey = "context1";
        String workContextValue = "value1";
        Pipeline pipeline = Pipeline
            .builder("name")
            .workContext(WorkContext.of(workContextKey, workContextValue))
            .build();

        // when
        PipelineOuterClass.Pipeline pipelineProtobuf = pipeline.toProtobuf();

        // then
        assertThat(pipelineProtobuf.getWorkContextCount(), equalTo(1));
        String workContext1 = pipelineProtobuf.getWorkContextMap().get(workContextKey);
        assertThat(workContext1, notNullValue());
        assertThat(workContext1, equalTo(workContextValue));
    }

    @Test
    public void shouldSerialiseWithOutputPaths() {
        // given
        String work1OutputName = "name";
        String work1OutputPath = "/var/yeetcd";
        Work work1 = Work
            .builder("test1", new CustomWorkDefinition() {
                @Override
                public void run() {
                    System.out.println("test1");
                }
            })
            .workOutputPaths(WorkOutputPath.builder(work1OutputName, work1OutputPath).build())
            .build();
        String work1MountPath = "/var/work1/outputs";
        Work work2 = Work
            .builder("test2", new CustomWorkDefinition() {
                @Override
                public void run() {
                    System.out.println("test2");
                }
            })
            .previousWork(PreviousWork.builder(work1).outputsMountPath(work1MountPath).build())
            .build();

        Pipeline pipeline = Pipeline
            .builder("name")
            .finalWork(work2)
            .build();


        // when
        PipelineOuterClass.Pipeline pipelineProtobuf = pipeline.toProtobuf();

        // then
        assertThat(pipelineProtobuf.getFinalWorkList().size(), equalTo(1));
        PipelineOuterClass.Work previousWork = pipelineProtobuf.getFinalWorkList().get(0);
        assertThat(previousWork.getPreviousWorkList().size(), equalTo(1));
        assertThat(previousWork.getPreviousWork(0).getWork().getOutputPaths(0).getName(), equalTo(work1OutputName));
        assertThat(previousWork.getPreviousWork(0).getWork().getOutputPaths(0).getPath(), equalTo(work1OutputPath));
        assertThat(previousWork.getPreviousWork(0).getOutputPathsMount(), equalTo(work1MountPath));
    }

    @Test
    public void shouldSerialiseWithStdOutEnvVar() {
        // given
        Work work1 = Work
            .builder("test1", new CustomWorkDefinition() {
                @Override
                public void run() {
                    System.out.println("test1");
                }
            })
            .build();
        String stdOutEnvVar = "ENV_VAR_NAME";
        Work work2 = Work
            .builder("test2", new CustomWorkDefinition() {
                @Override
                public void run() {
                    System.out.println("test2");
                }
            })
            .previousWork(PreviousWork.builder(work1).stdOutEnvVar(stdOutEnvVar).build())
            .build();

        Pipeline pipeline = Pipeline
            .builder("name")
            .finalWork(work2)
            .build();

        // when
        PipelineOuterClass.Pipeline pipelineProtobuf = pipeline.toProtobuf();

        // then
        assertThat(pipelineProtobuf.getFinalWorkList().size(), equalTo(1));
        PipelineOuterClass.Work previousWork = pipelineProtobuf.getFinalWorkList().get(0);
        assertThat(previousWork.getPreviousWorkList().size(), equalTo(1));
        assertThat(previousWork.getPreviousWork(0).getStdOutEnvVar(), equalTo(stdOutEnvVar));
    }

    @Test
    public void shouldSerialiseWorkWithContext() {
        // given
        String contextKey1 = "contextKey1";
        String contextValue1 = "contextValue1";
        WorkContext workContext = WorkContext.of(contextKey1, contextValue1);
        Pipeline pipeline = Pipeline
            .builder("name")
            .finalWork(Work
                .builder("test", ContainerisedWorkDefinition.builder("image").command("cmd").build())
                .workContext(workContext)
                .build()
            )
            .build();

        // when
        PipelineOuterClass.Pipeline pipelineProtobuf = pipeline.toProtobuf();

        // then
        assertThat(pipelineProtobuf.getFinalWorkList().size(), equalTo(1));
        assertThat(pipelineProtobuf.getFinalWorkList().get(0).getWorkContextMap(), equalTo(workContext.getWorkContextMap()));
    }

    @Test
    public void shouldSerialiseWithContainerisedWork() {
        // given
        String image = "image";
        String cmd = "cmd";
        Pipeline pipeline = Pipeline
            .builder("name")
            .finalWork(Work.builder("test", ContainerisedWorkDefinition.builder(image).command(cmd).build()).build())
            .build();

        // when
        PipelineOuterClass.Pipeline pipelineProtobuf = pipeline.toProtobuf();

        // then
        assertThat(pipelineProtobuf.getFinalWorkList().size(), equalTo(1));
        assertThat(pipelineProtobuf.getFinalWorkList().get(0).hasContainerisedWorkDefinition(), equalTo(true));
        assertThat(pipelineProtobuf.getFinalWorkList().get(0).getContainerisedWorkDefinition().getImage(), equalTo(image));
        assertThat(pipelineProtobuf.getFinalWorkList().get(0).getContainerisedWorkDefinition().getCmdList().size(), equalTo(1));
        assertThat(pipelineProtobuf.getFinalWorkList().get(0).getContainerisedWorkDefinition().getCmdList().get(0), equalTo(cmd));
    }

    @Test
    public void shouldSerialiseWithCustomWork() {
        // given
        Pipeline pipeline = Pipeline
            .builder("name")
            .finalWork(
                Work
                    .builder("test1", new CustomWorkDefinition() {
                        @Override
                        public void run() {
                            System.out.println("test1");
                        }
                    })
                    .build(),
                Work
                    .builder("test2", new CustomWorkDefinition() {
                        @Override
                        public void run() {
                            System.out.println("test2");
                        }
                    })
                    .build()
            )
            .build();

        // when
        PipelineOuterClass.Pipeline pipelineProtobuf = pipeline.toProtobuf();

        // then
        assertThat(pipelineProtobuf.getFinalWorkList().size(), equalTo(2));
        assertThat(pipelineProtobuf.getFinalWorkList().get(0).hasCustomWorkDefinition(), equalTo(true));
        String executionId1 = pipelineProtobuf.getFinalWorkList().get(0).getCustomWorkDefinition().getExecutionId();
        assertThat(executionId1, notNullValue());
        String executionId2 = pipelineProtobuf.getFinalWorkList().get(1).getCustomWorkDefinition().getExecutionId();
        assertThat(executionId2, notNullValue());
        assertThat(executionId1, not(equalTo(executionId2)));
    }

    @Test
    public void shouldSerialiseWithCompoundWork() {
        // given
        Pipeline pipeline = Pipeline
            .builder("name")
            .finalWork(
                Work.builder("test", CompoundWorkDefinition
                        .builder(
                            Work
                                .builder("test1", new CustomWorkDefinition() {
                                    @Override
                                    public void run() {
                                        System.out.println("test1");
                                    }
                                })
                                .build(),
                            Work
                                .builder("test2", new CustomWorkDefinition() {
                                    @Override
                                    public void run() {
                                        System.out.println("test2");
                                    }
                                })
                                .build()
                        )
                        .build()
                    )
                    .build()
            )
            .build();

        // when
        PipelineOuterClass.Pipeline pipelineProtobuf = pipeline.toProtobuf();

        // then
        assertThat(pipelineProtobuf.getFinalWorkList().size(), equalTo(1));
        assertThat(pipelineProtobuf.getFinalWorkList().get(0).hasCompoundWorkDefinition(), equalTo(true));
        List<PipelineOuterClass.Work> compoundWorkList = pipelineProtobuf.getFinalWorkList().get(0).getCompoundWorkDefinition().getFinalWorkList();
        assertThat(compoundWorkList.size(), equalTo(2));
    }

    @Test
    public void shouldHaveSingleIdentityForPreviousWorkInFanOut() {
        // given
        Work work1 = Work
            .builder("work1", ContainerisedWorkDefinition.builder(UUID.randomUUID().toString()).command(UUID.randomUUID().toString()).build())
            .build();
        Work work2 = Work
            .builder("work2", ContainerisedWorkDefinition.builder(UUID.randomUUID().toString()).command(UUID.randomUUID().toString()).build())
            .previousWork(PreviousWork.builder(work1).build())
            .build();
        Work work3 = Work
            .builder("work3", ContainerisedWorkDefinition.builder(UUID.randomUUID().toString()).command(UUID.randomUUID().toString()).build())
            .previousWork(PreviousWork.builder(work1).build())
            .build();
        Pipeline pipeline = Pipeline
            .builder("name")
            .finalWork(work2, work3)
            .build();

        // when
        PipelineOuterClass.Pipeline pipelineProtobuf = pipeline.toProtobuf();

        // then
        assertThat(pipelineProtobuf.getFinalWorkList().size(), equalTo(2));
        assertThat(pipelineProtobuf.getFinalWorkList().get(0).hasContainerisedWorkDefinition(), equalTo(true));
        assertThat(pipelineProtobuf.getFinalWorkList().get(0).getPreviousWorkList().size(), equalTo(1));
        String work2PreviousWorkId = pipelineProtobuf.getFinalWorkList().get(0).getPreviousWorkList().get(0).getWork().getId();
        assertThat(pipelineProtobuf.getFinalWorkList().get(1).hasContainerisedWorkDefinition(), equalTo(true));
        assertThat(pipelineProtobuf.getFinalWorkList().get(1).getPreviousWorkList().size(), equalTo(1));
        String work3PreviousWorkId = pipelineProtobuf.getFinalWorkList().get(1).getPreviousWorkList().get(0).getWork().getId();
        assertThat(work2PreviousWorkId, equalTo(work3PreviousWorkId));
    }

    @Test
    public void shouldHaveSameIdentityForEquivalentCopiesOfWork() {
        // given
        String description = "work1";
        String image = UUID.randomUUID().toString();
        String cmd = UUID.randomUUID().toString();
        Work work1a = Work
            .builder(description, ContainerisedWorkDefinition.builder(image).command(cmd).build())
            .build();
        Work work1b = Work
            .builder(description, ContainerisedWorkDefinition.builder(image).command(cmd).build())
            .build();
        Work work2 = Work
            .builder("work2", ContainerisedWorkDefinition.builder(UUID.randomUUID().toString()).command(UUID.randomUUID().toString()).build())
            .previousWork(PreviousWork.builder(work1a).build())
            .build();
        Work work3 = Work
            .builder("work3", ContainerisedWorkDefinition.builder(UUID.randomUUID().toString()).command(UUID.randomUUID().toString()).build())
            .previousWork(PreviousWork.builder(work1b).build())
            .build();
        Pipeline pipeline = Pipeline
            .builder("name")
            .finalWork(work2, work3)
            .build();

        // when
        PipelineOuterClass.Pipeline pipelineProtobuf = pipeline.toProtobuf();

        // then
        assertThat(pipelineProtobuf.getFinalWorkList().size(), equalTo(2));

        assertThat(pipelineProtobuf.getFinalWorkList().get(0).hasContainerisedWorkDefinition(), equalTo(true));
        assertThat(pipelineProtobuf.getFinalWorkList().get(0).getPreviousWorkList().size(), equalTo(1));
        String work2PreviousWorkId = pipelineProtobuf.getFinalWorkList().get(0).getPreviousWorkList().get(0).getWork().getId();

        assertThat(pipelineProtobuf.getFinalWorkList().get(1).hasContainerisedWorkDefinition(), equalTo(true));
        assertThat(pipelineProtobuf.getFinalWorkList().get(1).getPreviousWorkList().size(), equalTo(1));
        String work3PreviousWorkId = pipelineProtobuf.getFinalWorkList().get(1).getPreviousWorkList().get(0).getWork().getId();

        assertThat(work2PreviousWorkId, equalTo(work3PreviousWorkId));
    }

    @Test
    public void shouldHaveDistinctIdentityForEquivalentWorkWhenTheContextIsDifferent() {
        // given
        Work work1 = Work.builder("work1", ContainerisedWorkDefinition.builder(UUID.randomUUID().toString()).command(UUID.randomUUID().toString()).build()).build();
        Pipeline pipeline = Pipeline
            .builder("name")
            .finalWork(
                Work
                    .builder("compound1", CompoundWorkDefinition.builder(work1).build())
                    .workContext(WorkContext.of("key", "value1"))
                    .build(),
                Work
                    .builder("compound2", CompoundWorkDefinition.builder(work1).build())
                    .workContext(WorkContext.of("key", "value2"))
                    .build()
            )
            .build();

        // when
        PipelineOuterClass.Pipeline pipelineProtobuf = pipeline.toProtobuf();

        // then
        assertThat(pipelineProtobuf.getFinalWorkList().size(), equalTo(2));

        assertThat(pipelineProtobuf.getFinalWorkList().get(0).hasCompoundWorkDefinition(), equalTo(true));
        List<PipelineOuterClass.Work> compoundWork1FinalWorkList = pipelineProtobuf.getFinalWorkList().get(0).getCompoundWorkDefinition().getFinalWorkList();
        assertThat(compoundWork1FinalWorkList.size(), equalTo(1));
        String compoundWork1FinalWorkId = compoundWork1FinalWorkList.get(0).getId();

        assertThat(pipelineProtobuf.getFinalWorkList().get(1).hasCompoundWorkDefinition(), equalTo(true));
        List<PipelineOuterClass.Work> compoundWork2FinalWorkList = pipelineProtobuf.getFinalWorkList().get(1).getCompoundWorkDefinition().getFinalWorkList();
        assertThat(compoundWork2FinalWorkList.size(), equalTo(1));
        String compoundWork2FinalWorkId = compoundWork2FinalWorkList.get(0).getId();

        assertThat(compoundWork1FinalWorkId, not(equalTo(compoundWork2FinalWorkId)));
    }

    @Test
    public void shouldHaveSameIdentityForTheEquivalentWorkWhenTheContextIsEquivalent() {
        // given
        Work work1 = Work.builder("work1", ContainerisedWorkDefinition.builder(UUID.randomUUID().toString()).command(UUID.randomUUID().toString()).build()).build();
        Pipeline pipeline = Pipeline
            .builder("name")
            .finalWork(
                Work
                    .builder("compound1", CompoundWorkDefinition.builder(work1).build())
                    .workContext(WorkContext.of("key", "value1"))
                    .build(),
                Work
                    .builder("compound2", CompoundWorkDefinition.builder(work1).build())
                    .workContext(WorkContext.of("key", "value1"))
                    .build()
            )
            .build();

        // when
        PipelineOuterClass.Pipeline pipelineProtobuf = pipeline.toProtobuf();

        // then
        assertThat(pipelineProtobuf.getFinalWorkList().size(), equalTo(2));

        assertThat(pipelineProtobuf.getFinalWorkList().get(0).hasCompoundWorkDefinition(), equalTo(true));
        List<PipelineOuterClass.Work> compoundWork1FinalWorkList = pipelineProtobuf.getFinalWorkList().get(0).getCompoundWorkDefinition().getFinalWorkList();
        assertThat(compoundWork1FinalWorkList.size(), equalTo(1));
        String compoundWork1FinalWorkId = compoundWork1FinalWorkList.get(0).getId();

        assertThat(pipelineProtobuf.getFinalWorkList().get(1).hasCompoundWorkDefinition(), equalTo(true));
        List<PipelineOuterClass.Work> compoundWork2FinalWorkList = pipelineProtobuf.getFinalWorkList().get(1).getCompoundWorkDefinition().getFinalWorkList();
        assertThat(compoundWork2FinalWorkList.size(), equalTo(1));
        String compoundWork2FinalWorkId = compoundWork2FinalWorkList.get(0).getId();

        assertThat(compoundWork1FinalWorkId, equalTo(compoundWork2FinalWorkId));
    }

    @Test
    public void shouldSerialiseDefaultCondition() {
        // given
        Pipeline pipeline = Pipeline
            .builder("name")
            .finalWork(Work
                .builder("test", ContainerisedWorkDefinition
                    .builder("image")
                    .command("cmd")
                    .build()
                )
                .build()
            )
            .build();

        // when
        PipelineOuterClass.Pipeline pipelineProtobuf = pipeline.toProtobuf();

        // then
        assertThat(pipelineProtobuf.getFinalWorkList().size(), equalTo(1));
        assertThat(pipelineProtobuf.getFinalWorkList().get(0).getCondition().hasPreviousWorkStatusCondition(), equalTo(true));
        assertThat(pipelineProtobuf.getFinalWorkList().get(0).getCondition().getPreviousWorkStatusCondition().getStatus(), equalTo(PipelineOuterClass.PreviousWorkStatusCondition.Status.SUCCESS));
    }

    @Test
    public void shouldSerialiseComplexCondition() {
        // given
        String key1 = "key1";
        String value1 = "value1";
        String key2 = "key2";
        String value2 = "value2";
        Pipeline pipeline = Pipeline
            .builder("name")
            .finalWork(Work
                .builder("test", ContainerisedWorkDefinition
                    .builder("image")
                    .command("cmd")
                    .build()
                )
                .condition(
                    or(
                        and(
                            workContextCondition(key1, WorkContextCondition.Operand.EQUALS, value1),
                            workContextCondition(key2, WorkContextCondition.Operand.EQUALS, value2)
                        ),
                        previousWorkStatusCondition(PreviousWorkStatusCondition.Status.FAILURE)
                    ))
                .build()
            )
            .build();

        // when
        PipelineOuterClass.Pipeline pipelineProtobuf = pipeline.toProtobuf();

        // then
        assertThat(pipelineProtobuf.getFinalWorkList().size(), equalTo(1));
        assertThat(pipelineProtobuf.getFinalWorkList().get(0).getCondition().hasOrCondition(), equalTo(true));
        PipelineOuterClass.OrCondition orCondition = pipelineProtobuf.getFinalWorkList().get(0).getCondition().getOrCondition();

        assertThat(orCondition.getLeft().hasAndCondition(), equalTo(true));
        PipelineOuterClass.AndCondition and = orCondition.getLeft().getAndCondition();

        assertThat(and.getLeft().hasWorkContextCondition(), equalTo(true));
        assertThat(and.getLeft().getWorkContextCondition().getKey(), equalTo(key1));
        assertThat(and.getLeft().getWorkContextCondition().getOperand(), equalTo(PipelineOuterClass.WorkContextCondition.Operand.EQUALS));
        assertThat(and.getLeft().getWorkContextCondition().getValue(), equalTo(value1));

        assertThat(and.getRight().hasWorkContextCondition(), equalTo(true));
        assertThat(and.getRight().getWorkContextCondition().getKey(), equalTo(key2));
        assertThat(and.getRight().getWorkContextCondition().getOperand(), equalTo(PipelineOuterClass.WorkContextCondition.Operand.EQUALS));
        assertThat(and.getRight().getWorkContextCondition().getValue(), equalTo(value2));

        assertThat(orCondition.getRight().hasPreviousWorkStatusCondition(), equalTo(true));
        PipelineOuterClass.PreviousWorkStatusCondition previousWorkStatusCondition = orCondition.getRight().getPreviousWorkStatusCondition();
        assertThat(previousWorkStatusCondition.getStatus(), equalTo(PipelineOuterClass.PreviousWorkStatusCondition.Status.FAILURE));
    }
}
