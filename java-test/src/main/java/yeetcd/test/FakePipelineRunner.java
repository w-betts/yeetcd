package yeetcd.test;

import yeetcd.sdk.*;
import yeetcd.sdk.condition.*;
import lombok.EqualsAndHashCode;
import lombok.SneakyThrows;
import lombok.ToString;

import java.io.File;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.*;
import java.util.stream.Collectors;
import java.util.stream.Stream;

@SuppressWarnings("SwitchStatementWithTooFewBranches")
@EqualsAndHashCode
@ToString
public final class FakePipelineRunner {

    private final FakeWorkResult defaultWorkResult;
    private final List<FakeWorkMatcherResult> specifiedWorkResults;

    private FakePipelineRunner(FakeWorkResult defaultWorkResult, List<FakeWorkMatcherResult> specifiedWorkResults) {
        this.defaultWorkResult = defaultWorkResult;
        this.specifiedWorkResults = specifiedWorkResults;
    }

    public FakePipelineRunResult run(FakePipelineRun pipelineRun) {
        Work[] finalWork = pipelineRun.getPipeline().getFinalWork();
        WorkContext pipelineContext = WorkContext.of(pipelineRun.getArguments()).mergeInto(pipelineRun.getPipeline().getWorkContext());
        Map<WorkKey, FakeWorkExecution> completedWork = new HashMap<>();
        ExecutionResult executionResult = getExecutionResult(finalWork, pipelineContext, completedWork);
        return FakePipelineRunResult.builder(executionResult.workExecutionStages()).status(executionResult.status()).build();
    }

    private ExecutionResult getExecutionResult(Work[] finalWork, WorkContext containingContext, Map<WorkKey, FakeWorkExecution> completedWork) {
        List<FakeWorkExecutionStage> workExecutionStages = new LinkedList<>();
        Set<WorkKey> remainingWork = getAllWork(containingContext, finalWork);
        remainingWork.removeAll(completedWork.keySet());
        FakePipelineStatus status = FakePipelineStatus.SUCCESS;
        while (!remainingWork.isEmpty()) {
            Set<FakeWorkExecution> workExecutions = new HashSet<>();
            Set<Work> unblockedWork = remainingWork.stream()
                .map(WorkKey::work)
                .filter(work -> Arrays
                    .stream(work.getPreviousWork())
                    .allMatch(previousWork -> completedWork.containsKey(WorkKey.workKey(containingContext, previousWork.getWork())))
                )
                .collect(Collectors.toSet());
            for (Work work : unblockedWork) {
                FakeWorkExecution fakeWorkExecution = fakeWorkExecution(work, containingContext, completedWork);
                workExecutions.add(fakeWorkExecution);
                WorkKey workKey = WorkKey.workKey(containingContext, work);
                completedWork.put(workKey, fakeWorkExecution);
                remainingWork.remove(workKey);
                if (fakeWorkExecution.getStatus() == FakeWorkStatus.FAILURE) {
                    status = FakePipelineStatus.FAILURE;
                }
            }
            workExecutionStages.add(FakeWorkExecutionStage.builder(workExecutions).build());
        }
        return new ExecutionResult(workExecutionStages, status);
    }

    private FakeWorkExecution fakeWorkExecution(Work work, WorkContext containingContext, Map<WorkKey, FakeWorkExecution> completedWork) {
        boolean shouldExecute = evaluate(work.getCondition(), work.getWorkContext().mergeInto(containingContext), completedWork);
        WorkContext workContext = work.getWorkContext().mergeInto(containingContext);
        if (work.getWorkDefinition() instanceof ContainerisedWorkDefinition || work.getWorkDefinition() instanceof CustomWorkDefinition) {
            Map<String, byte[]> exportedFiles = Collections.emptyMap();
            FakeWorkStatus status = FakeWorkStatus.SKIPPED;
            String stdOut = "";
            if (shouldExecute) {
                Optional<FakeWorkMatcherResult> matcherResult = specifiedWorkResults.stream()
                    .filter(it -> it.getWorkMatcher().matches(work, workContext))
                    .findFirst();
                if (matcherResult.isPresent()) {
                    status = matcherResult.get().getWorkResult().getStatus();
                    stdOut = matcherResult.get().getWorkResult().getStdOut();
                    exportedFiles = matcherResult.get().getWorkResult().getExportedFiles();
                }
                else {
                    status = defaultWorkResult.getStatus();
                    stdOut = defaultWorkResult.getStdOut();
                    exportedFiles = defaultWorkResult.getExportedFiles();
                }
            }
            return FakeSimpleWorkExecution
                .builder(work)
                .status(status)
                .envVars(envVars(work, containingContext, completedWork))
                .inputFiles(inputFiles(work, containingContext, completedWork))
                .stdOut(stdOut)
                .exportedFiles(exportedFiles)
                .build();
        }
        else if (work.getWorkDefinition() instanceof CompoundWorkDefinition compoundWorkDefinition) {
            return FakeCompoundWorkExecution.builder(getExecutionResult(compoundWorkDefinition.getFinalWork(), workContext, completedWork).workExecutionStages()).build();
        }
        else if (work.getWorkDefinition() instanceof DynamicWorkGeneratingWorkDefinition dynamicWorkGeneratingWorkDefinition) {
            Arrays.stream(work.getPreviousWork()).forEach(previousWork -> {
                envVars(work, containingContext, completedWork).forEach(System::setProperty);
                inputFiles(work, containingContext, completedWork).forEach(FakePipelineRunner::sneakyWrite);
            });
            return fakeWorkExecution(dynamicWorkGeneratingWorkDefinition.createWork(), containingContext, completedWork);
        }
        else {
            throw new IllegalArgumentException("Work argument has unsupported work definition type %s".formatted(work.getWorkDefinition().getClass().getName()));
        }
    }

    @SuppressWarnings("ResultOfMethodCallIgnored")
    @SneakyThrows
    private static void sneakyWrite(String path, byte[] value) {
        File file = new File(path);
        file.getParentFile().mkdirs();
        if (file.createNewFile()) {
            Files.write(Path.of(path), value);
        } else {
            throw new IllegalArgumentException("Not safe to proceed as input file already exists");
        }
    }

    private static Map<String, String> envVars(Work work, WorkContext containingContext, Map<WorkKey, FakeWorkExecution> completedWork) {
        Map<String, String> envVars = new HashMap<>();
        Arrays
            .stream(work.getPreviousWork())
            .forEach(previousWork -> {
                FakeWorkExecution previousWorkExecution = completedWork.get(WorkKey.workKey(containingContext, previousWork.getWork()));
                if (previousWorkExecution instanceof FakeSimpleWorkExecution && previousWork.getStdOutEnvVar() != null) {
                    envVars.put(previousWork.getStdOutEnvVar(), ((FakeSimpleWorkExecution) previousWorkExecution).getStdOut());
                }
            });
        envVars.putAll(work.getWorkContext().mergeInto(containingContext).getWorkContextMap());
        return envVars;
    }

    private static Map<String, byte[]> inputFiles(Work work, WorkContext containingContext, Map<WorkKey, FakeWorkExecution> completedWork) {
        Map<String, byte[]> inputFiles = new HashMap<>();
        Arrays
            .stream(work.getPreviousWork())
            .filter(previousWork -> previousWork.getOutputPathsMount() != null)
            .forEach(previousWork -> {
                FakeWorkExecution previousWorkExecution = completedWork.get(WorkKey.workKey(containingContext, previousWork.getWork()));
                if (previousWorkExecution instanceof FakeSimpleWorkExecution) {
                    Arrays.stream(previousWork.getWork().getWorkOutputPaths()).forEach(workOutputPath -> {
                        byte[] value = ((FakeSimpleWorkExecution) previousWorkExecution).getExportedFiles().get(workOutputPath.getPath());
                        if (value != null) {
                            inputFiles.put(
                                "%s/%s".formatted(previousWork.getOutputPathsMount(), workOutputPath.getName()),
                                value
                            );
                        }
                    });
                }
            });
        return inputFiles;
    }

    private record WorkKey(Work work, WorkContext workContext) {
        private static WorkKey workKey(WorkContext containingContext, Work work) {
            return new WorkKey(work, work.getWorkContext().mergeInto(containingContext));
        }
    }

    private record ExecutionResult(List<FakeWorkExecutionStage> workExecutionStages, FakePipelineStatus status) {
    }

    private static Set<WorkKey> getAllWork(WorkContext containingContext, Work... works) {
        return Stream
            .concat(
                Arrays
                    .stream(works)
                    .flatMap(workItem -> Arrays.stream(workItem.getPreviousWork())
                        .flatMap(previousWork -> getAllWork(containingContext, previousWork.getWork()).stream())),
                Arrays
                    .stream(works)
                    .map(work -> new WorkKey(work, work.getWorkContext().mergeInto(containingContext)))
            )
            .collect(Collectors.toCollection(HashSet::new));
    }

    private static boolean evaluate(Condition condition, WorkContext workContext, Map<WorkKey, FakeWorkExecution> completedWork) {
        if (condition instanceof AndCondition andCondition) {
            return evaluate(andCondition.getLeft(), workContext, completedWork) &&
                   evaluate(andCondition.getRight(), workContext, completedWork);
        }
        else if (condition instanceof OrCondition orCondition) {
            return evaluate(orCondition.getLeft(), workContext, completedWork) ||
                   evaluate(orCondition.getRight(), workContext, completedWork);
        }
        else if (condition instanceof NotCondition notCondition) {
            return !evaluate(notCondition.getCondition(), workContext, completedWork);
        }
        else if (condition instanceof WorkContextCondition workContextCondition) {
            WorkContextCondition.Operand operand = workContextCondition.getOperand();
            switch (operand) {
                case EQUALS -> {
                    return workContextCondition.getValue().equals(workContext.getWorkContextMap().get(workContextCondition.getKey()));
                }
                default -> throw new IllegalArgumentException("Unsupported WorkContextCondition.Operand type %s".formatted(operand));
            }
        } else if (condition instanceof PreviousWorkStatusCondition previousWorkStatusCondition) {
            switch (previousWorkStatusCondition.getStatus()) {
                case SUCCESS -> {
                    return completedWork.values().stream().allMatch(result -> result.getStatus() == FakeWorkStatus.SUCCESS);
                }
                case FAILURE -> {
                    return completedWork.values().stream().anyMatch(result -> result.getStatus() == FakeWorkStatus.FAILURE);
                }
                case ANY -> {
                    return true;
                }
                default -> throw new IllegalArgumentException("Unsupported PreviousWorkStatusCondition.Status type %s".formatted(previousWorkStatusCondition.getStatus()));
            }
        } else {
            throw new IllegalArgumentException("Unexpected Condition type %s".formatted(condition));
        }
    }


    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {

        private FakeWorkResult defaultFakeWorkResult = FakeWorkResult.builder().build();
        private List<FakeWorkMatcherResult> specifiedWorkResults = Collections.emptyList();

        public Builder defaultWorkResult(FakeWorkResult defaultFakeWorkResult) {
            this.defaultFakeWorkResult = defaultFakeWorkResult;
            return this;
        }

        public Builder specifiedWorkResults(List<FakeWorkMatcherResult> specifiedWorkResults) {
            this.specifiedWorkResults = specifiedWorkResults;
            return this;
        }

        public Builder specifiedWorkResults(FakeWorkMatcherResult... specifiedWorkResults) {
            this.specifiedWorkResults = Arrays.stream(specifiedWorkResults).toList();
            return this;
        }

        public FakePipelineRunner build() {
            return new FakePipelineRunner(defaultFakeWorkResult, specifiedWorkResults);
        }
    }
}
