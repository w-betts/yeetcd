package yeetcd.test;

import java.util.List;
import java.util.Map;
import java.util.function.Predicate;
import java.util.stream.Collectors;

public class PipelineTestRunResult {
    
    private final PipelineStatus pipelineStatus;
    private final int pipelineExitCode;
    private final List<WorkExecution> executions;
    private final List<MatchedBehavior> matchedBehaviors;
    private final String cliOutput;

    public PipelineTestRunResult(
            PipelineStatus pipelineStatus,
            int pipelineExitCode,
            List<WorkExecution> executions,
            List<MatchedBehavior> matchedBehaviors) {
        this(pipelineStatus, pipelineExitCode, executions, matchedBehaviors, "");
    }

    public PipelineTestRunResult(
            PipelineStatus pipelineStatus,
            int pipelineExitCode,
            List<WorkExecution> executions,
            List<MatchedBehavior> matchedBehaviors,
            String cliOutput) {
        this.pipelineStatus = pipelineStatus;
        this.pipelineExitCode = pipelineExitCode;
        this.executions = List.copyOf(executions);
        this.matchedBehaviors = List.copyOf(matchedBehaviors);
        this.cliOutput = cliOutput != null ? cliOutput : "";
    }

    public PipelineStatus getPipelineStatus() {
        return pipelineStatus;
    }

    public int getPipelineExitCode() {
        return pipelineExitCode;
    }

    public List<WorkExecution> getExecutions() {
        return executions;
    }

    public List<MatchedBehavior> getMatchedBehaviors() {
        return matchedBehaviors;
    }

    public String getCliOutput() {
        return cliOutput;
    }

    public List<WorkExecution> findByImage(String image) {
        return executions.stream()
                .filter(e -> image.equals(e.image()))
                .collect(Collectors.toList());
    }

    public List<WorkExecution> findByInstance(String executionId) {
        return executions.stream()
                .filter(e -> e.type() == WorkBehaviorType.CUSTOM && executionId.equals(e.matchKey()))
                .collect(Collectors.toList());
    }

    public boolean hasExecution(String image) {
        return executions.stream().anyMatch(e -> image.equals(e.image()));
    }

    public boolean hasNoExecution(String image) {
        return executions.stream().noneMatch(e -> image.equals(e.image()));
    }

    public int getExecutionCount(String image) {
        return (int) executions.stream()
                .filter(e -> image.equals(e.image()))
                .count();
    }

    public List<WorkExecution> getContainerisedExecutions() {
        return executions.stream()
                .filter(e -> e.type() == WorkBehaviorType.CONTAINERISED)
                .collect(Collectors.toList());
    }

    public List<WorkExecution> getCustomExecutions() {
        return executions.stream()
                .filter(e -> e.type() == WorkBehaviorType.CUSTOM)
                .collect(Collectors.toList());
    }

    public List<WorkExecution> getDynamicExecutions() {
        return executions.stream()
                .filter(e -> e.type() == WorkBehaviorType.DYNAMIC)
                .collect(Collectors.toList());
    }
}
