package yeetcd.javasample;

import yeetcd.sdk.*;
import yeetcd.sdk.condition.*;

import java.io.IOException;
import java.net.URI;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.List;
import java.util.stream.IntStream;
import java.util.stream.Stream;

import static yeetcd.sdk.condition.Conditions.workContextCondition;

public class SamplePipelines {

    @PipelineGenerator
    public static Pipeline sample() {
        Work containerisedWork = Work
            .builder("containerised-work-definition", ContainerisedWorkDefinition
                .builder("maven:3.9.9-eclipse-temurin-17")
                .command("bash", "-c", "echo 'Hello from a containerised task'")
                .build()
            )
            .build();
        Work customWork = Work
            .builder("custom-work-definition", new CustomWorkDefinition() {
                @Override
                public void run() {
                    System.out.println("Hello from a custom work definition");
                }
            })
            .previousWork(PreviousWork.builder(containerisedWork).build())
            .build();
        return Pipeline.builder("sample").finalWork(customWork).build();
    }

    @PipelineGenerator
    public static Pipeline sampleCompound() {
        Work compoundWork1 = sample()
            .asWorkBuilder("sample-pipeline-work-1")
            .workContext(WorkContext.of("part", "1"))
            .build();
        Work compoundWork2 = sample()
            .asWorkBuilder("sample-pipeline-work-2")
            .workContext(WorkContext.of("part", "2"))
            .previousWork(PreviousWork.builder(compoundWork1).build())
            .build();
        return Pipeline.builder("sampleCompound").finalWork(compoundWork2).build();
    }

    @PipelineGenerator
    public static Pipeline sampleWithWorkContext() {
        String pipelineWorkContextValue = "pipelineWorkContext";
        WorkContext pipelineWorkContext = WorkContext.of("PIPELINE_WORK_CONTEXT", pipelineWorkContextValue);
        String workWorkContextValue = "workWorkContext";
        WorkContext workWorkContext = WorkContext.of("WORK_WORK_CONTEXT", workWorkContextValue);
        Work work = Work
            .builder("containerised-work-definition", new CustomWorkDefinition() {
                @Override
                public void run() {
                    String pipelineWorkContextActualValue = System.getenv("PIPELINE_WORK_CONTEXT");
                    System.out.printf("PIPELINE_WORK_CONTEXT has value '%s'%n", pipelineWorkContextActualValue);
                    if (!pipelineWorkContextValue.equals(pipelineWorkContextActualValue)) {
                        System.out.printf("Expected PIPELINE_WORK_CONTEXT to have value '%s'%n", pipelineWorkContextValue);
                        System.exit(1);
                    }
                    String workWorkContextActualValue = System.getenv("WORK_WORK_CONTEXT");
                    System.out.printf("WORK_WORK_CONTEXT has value '%s'%n", workWorkContextActualValue);
                    if (!workWorkContextValue.equals(workWorkContextActualValue)) {
                        System.out.printf("Expected WORK_WORK_CONTEXT to have value '%s'%n", workWorkContextValue);
                        System.exit(1);
                    }
                }
            })
            .workContext(workWorkContext)
            .build();
        return Pipeline.builder("sampleWithWorkContext").workContext(pipelineWorkContext).finalWork(work).build();
    }

    @PipelineGenerator
    public static Pipeline sampleWithParameters() {
        String defaultValue = "default";
        List<String> choices = List.of(defaultValue, "other");
        Parameter parameter = Parameter
            .builder(Parameter.TypeCheck.STRING)
            .required(true)
            .defaultValue(defaultValue)
            .choices(choices)
            .build();
        String parameterName = "PARAMETER_NAME";
        Work work = Work
            .builder("containerised-work-definition", new CustomWorkDefinition() {
                @Override
                public void run() {
                    String envValue = System.getenv(parameterName);
                    String parameterActualValue = envValue == null ? "" : envValue;
                    System.out.printf("PARAMETER_NAME has value '%s'%n", parameterActualValue);
                    if (!choices.contains(parameterActualValue)) {
                        System.out.printf("Expected PARAMETER_NAME to have value in [%s]%n", String.join(", ", choices));
                        System.exit(1);
                    }
                }
            })
            .build();
        return Pipeline.builder("sampleWithParameters").parameters(Parameters.of(parameterName, parameter)).finalWork(work).build();
    }

    @PipelineGenerator
    public static Pipeline sampleWithWorkOutputs() {
        String expectedOutput = "expected output";
        String name = "expected-output";
        String filePath = "/var/log/expected-output";
        String stdOut = "stdOutOutput";
        Work produceOutputWork = Work
            .builder(
                "generate-outputs", new CustomWorkDefinition() {
                    @Override
                    public void run() {
                        try {
                            Files.writeString(Path.of(URI.create("file://%s".formatted(filePath))), expectedOutput);
                            System.out.print(stdOut);
                        } catch (IOException e) {
                            throw new RuntimeException(e);
                        }
                    }
                }
            )
            .workOutputPaths(WorkOutputPath
                .builder(name, filePath)
                .build()
            )
            .build();

        String mountPath = "/var/output-generator/mount";
        String stdOutEnvVar = "stdOutEnvVarName";
        Work customWork = Work
            .builder("consume-outputs", new CustomWorkDefinition() {
                @Override
                public void run() {
                    String previousWorkStdOut = System.getenv(stdOutEnvVar);
                    System.out.printf("Work stdout has value '%s'%n", previousWorkStdOut);
                    if (!stdOut.equals(previousWorkStdOut)) {
                        System.out.printf("Expected stdout to have value %s%n", stdOut);
                        System.exit(1);
                    }
                    try {
                        String containerisedWorkOutput = Files.readString(Path.of(URI.create("file://%s/%s".formatted(mountPath, name))));
                        System.out.printf("Work file output has value '%s'%n", containerisedWorkOutput);
                        if (!expectedOutput.equals(containerisedWorkOutput)) {
                            System.out.printf("Expected work file output to have value %s%n", expectedOutput);
                            System.exit(1);
                        }
                    } catch (IOException e) {
                        throw new RuntimeException(e);
                    }
                }
            })
            .previousWork(PreviousWork
                .builder(produceOutputWork)
                .outputsMountPath(mountPath)
                .stdOutEnvVar(stdOutEnvVar)
                .build()
            )
            .build();
        return Pipeline.builder("sampleWithWorkOutputs").finalWork(customWork).build();
    }

    @PipelineGenerator
    public static Pipeline sampleWithConditions() {
        Work unconditionalWork = Work
            .builder("conditional-work-definition", new CustomWorkDefinition() {
                @Override
                public void run() {
                    System.out.println("This should run");
                }
            })
            .workContext(WorkContext.of("key", "value"))
            .condition(workContextCondition("key", WorkContextCondition.Operand.EQUALS, "value"))
            .build();
        Work conditionalWork = Work
            .builder("conditional-work-definition", new CustomWorkDefinition() {
                @Override
                public void run() {
                    System.out.println("This shouldn't run thanks to the condition");
                    System.exit(1);
                }
            })
            .condition(workContextCondition("missingKey", WorkContextCondition.Operand.EQUALS, "value"))
            .previousWork(PreviousWork.builder(unconditionalWork).build())
            .build();
        Work workDependentOnConditionalWork = Work
            .builder("conditional-work-definition", new CustomWorkDefinition() {
                @Override
                public void run() {
                    System.out.println("This shouldn't run because its dependency didn't run");
                    System.exit(1);
                }
            })
            .previousWork(PreviousWork.builder(conditionalWork).build())
            .build();
        return Pipeline.builder("sampleWithConditions").finalWork(workDependentOnConditionalWork).build();
    }

    @PipelineGenerator
    public static Pipeline sampleWithDynamicWork() {
        String workCountEnvVar = "WORK_COUNT";
        return Pipeline.builder("sampleDynamicWork")
            .finalWork(Work
                .builder("dynamic-work-generating-work-definition", new DynamicWorkGeneratingWorkDefinition() {
                    private final String workInstanceEnvVar = "WORK_INSTANCE";

                    private final CustomWorkDefinition workDefinition = new CustomWorkDefinition() {
                        @Override
                        public void run() {
                            System.out.printf("Work definition %s%n", System.getenv(workInstanceEnvVar));
                        }
                    };

                    @Override
                    public Work createWork() {
                        return Work
                            .builder("dynamicWork", CompoundWorkDefinition.builder(IntStream
                                .range(0, Integer.parseInt(System.getenv(workCountEnvVar)))
                                .mapToObj(i -> Work
                                    .builder(
                                        "work-%d".formatted(i),
                                        workDefinition
                                    )
                                    .workContext(WorkContext.of(workInstanceEnvVar, Integer.toString(i)))
                                    .build()
                                )
                                .toArray(Work[]::new)).build())
                            .build();
                    }

                    @Override
                    protected Stream<CustomWorkDefinition> dynamicCustomWorkDefinitions() {
                        return Stream.of(workDefinition);
                    }
                })
                .build()
            )
            .parameters(Parameters.of(workCountEnvVar, Parameter.builder(Parameter.TypeCheck.NUMBER).required(true).build()))
            .build();
    }
}
