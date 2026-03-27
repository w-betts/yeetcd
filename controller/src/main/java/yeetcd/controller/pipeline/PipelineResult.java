package yeetcd.controller.pipeline;

import java.util.Map;

public record PipelineResult(Map<Work, WorkResult> workResults) {

    public PipelineStatus pipelineStatus() {
        if (workResults.values().stream().anyMatch(taskResult -> taskResult.workStatus() == WorkStatus.FAILURE)) {
            return PipelineStatus.FAILURE;
        }
        return PipelineStatus.SUCCESS;
    }
}
