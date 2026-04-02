package yeetcd.test;

import yeetcd.sdk.WorkContext;
import lombok.EqualsAndHashCode;

import java.util.Arrays;
import java.util.List;

@EqualsAndHashCode
public final class FakeCompoundWorkExecution implements FakeWorkExecution {

    private final List<FakeWorkExecutionStage> workExecutionStages;
    private final WorkContext workContext;
    private final FakeWorkStatus status;

    private FakeCompoundWorkExecution(List<FakeWorkExecutionStage> workExecutionStages, WorkContext workContext) {
        this.workExecutionStages = workExecutionStages;
        this.workContext = workContext;
        if (workExecutionStages.isEmpty() || allMatch(workExecutionStages, FakeWorkStatus.SUCCESS)) {
            this.status = FakeWorkStatus.SUCCESS;
        } else if(allMatch(workExecutionStages, FakeWorkStatus.SKIPPED)) {
            this.status = FakeWorkStatus.SKIPPED;
        } else {
            this.status = FakeWorkStatus.FAILURE;
        }
    }

    private static boolean allMatch(List<FakeWorkExecutionStage> workExecutionStages, FakeWorkStatus status) {
        return workExecutionStages.stream()
            .allMatch(stage -> stage.getWorkExecutions().stream()
                .allMatch(fakeWorkExecution -> fakeWorkExecution.getStatus() == status)
            );
    }

    public static Builder builder(List<FakeWorkExecutionStage> workExecutionStages) {
        return new Builder(workExecutionStages);
    }

    public static Builder builder(FakeWorkExecutionStage... workExecutionStages) {
        return builder(Arrays.stream(workExecutionStages).toList());
    }

    @Override
    public FakeWorkStatus getStatus() {
        return status;
    }

    public static class Builder {
        private final List<FakeWorkExecutionStage> workExecutionStages;

        private WorkContext workContext = WorkContext.empty();

        private Builder(List<FakeWorkExecutionStage> workExecutionStages) {
            this.workExecutionStages = workExecutionStages;
        }

        public Builder workContext(WorkContext workContext) {
            this.workContext = workContext;
            return this;
        }

        public FakeCompoundWorkExecution build() {
            return new FakeCompoundWorkExecution(workExecutionStages, workContext);
        }
    }

    @Override
    public String toString() {
        return FakeWorkExecutionStage.toString(workExecutionStages);
    }
}
