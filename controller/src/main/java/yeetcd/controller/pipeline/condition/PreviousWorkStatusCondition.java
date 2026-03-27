package yeetcd.controller.pipeline.condition;

import yeetcd.controller.pipeline.WorkContext;
import yeetcd.controller.pipeline.WorkResult;
import yeetcd.controller.pipeline.WorkResultTracker;
import yeetcd.controller.pipeline.WorkStatus;
import yeetcd.protocol.pipeline.PipelineOuterClass;
import lombok.SneakyThrows;

import java.util.concurrent.CompletableFuture;

public record PreviousWorkStatusCondition(Status status) implements Condition {

    @Override
    @SneakyThrows
    public boolean evaluate(WorkContext workContext, WorkResultTracker workResultTracker) {
        switch (status) {
            case SUCCESS -> {
                return workResultTracker.getWorkResultMap().values().stream()
                    .filter(CompletableFuture::isDone)
                    .map(PreviousWorkStatusCondition::sneakyGet)
                    .allMatch(result -> result.workStatus() == WorkStatus.SUCCESS);
            }
            case FAILURE -> {
                return workResultTracker.getWorkResultMap().values().stream()
                    .filter(CompletableFuture::isDone)
                    .map(PreviousWorkStatusCondition::sneakyGet)
                    .anyMatch(result -> result.workStatus() == WorkStatus.FAILURE);
            }
            default -> {
                return true;
            }
        }
    }

    @SneakyThrows
    private static WorkResult sneakyGet(CompletableFuture<WorkResult> workResultCompletableFuture) {
        return workResultCompletableFuture.get();
    }

    public enum Status {
        SUCCESS,
        FAILURE,
        ANY;

        static Status fromProtobuf(PipelineOuterClass.PreviousWorkStatusCondition.Status status) {
            switch (status) {
                case SUCCESS -> {
                    return SUCCESS;
                }
                case FAILURE -> {
                    return FAILURE;
                }
                case ANY -> {
                    return ANY;
                }
                default -> throw new IllegalArgumentException("Unrecognised status %s".formatted(status.name()));
            }
        }
    }

    public static Condition fromProtobuf(PipelineOuterClass.PreviousWorkStatusCondition previousWorkStatusCondition) {
        return new PreviousWorkStatusCondition(
            Status.fromProtobuf(previousWorkStatusCondition.getStatus())
        );
    }
}
