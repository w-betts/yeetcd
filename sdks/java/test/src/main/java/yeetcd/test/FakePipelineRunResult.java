package yeetcd.test;

import lombok.EqualsAndHashCode;

import java.util.Arrays;
import java.util.List;

@EqualsAndHashCode
public final class FakePipelineRunResult {
    private final List<FakeWorkExecutionStage> workExecutionStages;
    private final FakePipelineStatus status;

    private FakePipelineRunResult(List<FakeWorkExecutionStage> workExecutionStages, FakePipelineStatus status) {
        this.workExecutionStages = workExecutionStages;
        this.status = status;
    }

    public static Builder builder(List<FakeWorkExecutionStage> workExecutionStages) {
        return new Builder(workExecutionStages);
    }

    public static Builder builder(FakeWorkExecutionStage... workExecutionStages) {
        return new Builder(Arrays.stream(workExecutionStages).toList());
    }

    public List<FakeWorkExecutionStage> getWorkExecutionStages() {
        return workExecutionStages;
    }

    public FakePipelineStatus getStatus() {
        return status;
    }

    public static class Builder {
        private final List<FakeWorkExecutionStage> workExecutionStages;
        private FakePipelineStatus status = FakePipelineStatus.SUCCESS;

        private Builder(List<FakeWorkExecutionStage> workExecutionStages) {
            this.workExecutionStages = workExecutionStages;
        }

        public Builder status(FakePipelineStatus status) {
            this.status = status;
            return this;
        }

        public FakePipelineRunResult build() {
            return new FakePipelineRunResult(workExecutionStages, status);
        }
    }

    @Override
    public String toString() {
        return """
            
            %s
            %s
            
            """.formatted(FakeWorkExecutionStage.toString(workExecutionStages), status.toString());
    }
}
