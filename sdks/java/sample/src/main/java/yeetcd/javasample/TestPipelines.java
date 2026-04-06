package yeetcd.javasample;

import yeetcd.sdk.*;
import yeetcd.sdk.condition.*;

import java.util.List;
import java.util.Map;
import java.util.stream.IntStream;
import java.util.stream.Stream;

public class TestPipelines {

    private static CustomWorkDefinition cachedCustomWork;

    public static CustomWorkDefinition getCustomWorkForPipeline() {
        if (cachedCustomWork == null) {
            cachedCustomWork = new CustomWorkDefinition() {
                @Override
                public void run() {
                    System.out.println("Custom work executed");
                }
            };
        }
        return cachedCustomWork;
    }

    @PipelineGenerator
    public static Pipeline containerisedWorkPipeline() {
        Work work = Work
            .builder("containerised-work", ContainerisedWorkDefinition
                .builder("maven:3.9.9-eclipse-temurin-17")
                .command("bash", "-c", "echo 'containerised work'")
                .build()
            )
            .build();
        return Pipeline.builder("containerisedWorkPipeline").finalWork(work).build();
    }

    @PipelineGenerator
    public static Pipeline customWorkPipeline() {
        Work work = Work
            .builder("custom-work", getCustomWorkForPipeline())
            .build();
        return Pipeline.builder("customWorkPipeline").finalWork(work).build();
    }

    @PipelineGenerator
    public static Pipeline compoundWorkPipeline() {
        Work subWork1 = Work
            .builder("sub-work-1", new CustomWorkDefinition() {
                @Override
                public void run() {
                    System.out.println("Sub work 1");
                }
            })
            .build();
        Work subWork2 = Work
            .builder("sub-work-2", new CustomWorkDefinition() {
                @Override
                public void run() {
                    System.out.println("Sub work 2");
                }
            })
            .build();
        Work compoundWork = Work
            .builder("compound-work", CompoundWorkDefinition.builder(subWork1, subWork2).build())
            .build();
        return Pipeline.builder("compoundWorkPipeline").finalWork(compoundWork).build();
    }

    @PipelineGenerator
    public static Pipeline dynamicWorkPipeline() {
        String workCountEnvVar = "WORK_COUNT";
        Work work = Work
            .builder("dynamic-work", new DynamicWorkGeneratingWorkDefinition() {
                private final String workInstanceEnvVar = "WORK_INSTANCE";

                private final CustomWorkDefinition workDefinition = new CustomWorkDefinition() {
                    @Override
                    public void run() {
                        System.out.printf("Dynamic work instance %s%n", System.getenv(workInstanceEnvVar));
                    }
                };

                @Override
                public Work createWork() {
                    int count = Integer.parseInt(System.getenv(workCountEnvVar));
                    return Work
                        .builder("dynamicChild", CompoundWorkDefinition.builder(
                            IntStream.range(0, count)
                                .mapToObj(i -> Work
                                    .builder("work-" + i, workDefinition)
                                    .workContext(WorkContext.of(workInstanceEnvVar, Integer.toString(i)))
                                    .build())
                                .toArray(Work[]::new))
                            .build())
                        .build();
                }

                @Override
                protected Stream<CustomWorkDefinition> dynamicCustomWorkDefinitions() {
                    return Stream.of(workDefinition);
                }
            })
            .build();
        return Pipeline.builder("dynamicWorkPipeline")
            .parameters(Parameters.of(workCountEnvVar, Parameter.builder(Parameter.TypeCheck.NUMBER).required(true).build()))
            .finalWork(work)
            .build();
    }

    @PipelineGenerator
    public static Pipeline dependentWorkPipeline() {
        Work firstWork = Work
            .builder("first-work", new CustomWorkDefinition() {
                @Override
                public void run() {
                    System.out.println("First work");
                }
            })
            .build();
        Work secondWork = Work
            .builder("second-work", new CustomWorkDefinition() {
                @Override
                public void run() {
                    System.out.println("Second work");
                }
            })
            .previousWork(PreviousWork.builder(firstWork).build())
            .build();
        return Pipeline.builder("dependentWorkPipeline").finalWork(secondWork).build();
    }

    @PipelineGenerator
    public static Pipeline contextWorkPipeline() {
        WorkContext pipelineContext = WorkContext.of("PIPELINE_VAR", "pipeline-value");
        WorkContext workContext = WorkContext.of("WORK_VAR", "work-value");
        Work work = Work
            .builder("context-work", new CustomWorkDefinition() {
                @Override
                public void run() {
                    String pipelineVar = System.getenv("PIPELINE_VAR");
                    String workVar = System.getenv("WORK_VAR");
                    System.out.printf("Pipeline context: %s, Work context: %s%n", pipelineVar, workVar);
                }
            })
            .workContext(workContext)
            .build();
        return Pipeline.builder("contextWorkPipeline").workContext(pipelineContext).finalWork(work).build();
    }

    @PipelineGenerator
    public static Pipeline multiBehaviorPipeline() {
        Work work1 = Work
            .builder("containerised-1", ContainerisedWorkDefinition
                .builder("maven:3.9.9-eclipse-temurin-17")
                .command("bash", "-c", "echo 'containerised 1'")
                .build()
            )
            .build();
        Work work2 = Work
            .builder("custom-1", new CustomWorkDefinition() {
                @Override
                public void run() {
                    System.out.println("Custom work 1");
                }
            })
            .previousWork(PreviousWork.builder(work1).build())
            .build();
        Work work3 = Work
            .builder("containerised-2", ContainerisedWorkDefinition
                .builder("maven:3.9.9-eclipse-temurin-17")
                .command("bash", "-c", "echo 'containerised 2'")
                .build()
            )
            .previousWork(PreviousWork.builder(work2).build())
            .build();
        return Pipeline.builder("multiBehaviorPipeline").finalWork(work3).build();
    }
}
