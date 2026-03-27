package yeetcd.controller.pipeline.condition;

import yeetcd.controller.pipeline.WorkContext;
import yeetcd.controller.pipeline.WorkResultTracker;
import yeetcd.protocol.pipeline.PipelineOuterClass;

import java.util.Objects;

public record WorkContextCondition(String key, Operand operand, String value) implements Condition {

    @Override
    public boolean evaluate(WorkContext workContext, WorkResultTracker workResultTracker) {
        return operand().evaluate(workContext, key, value);
    }

    public enum Operand {
        EQUALS {
            @Override
            boolean evaluate(WorkContext workContext, String key, String value) {
                return Objects.equals(workContext.workContextMap().get(key), value);
            }
        };

        @SuppressWarnings("SwitchStatementWithTooFewBranches")
        static Operand fromProtobuf(PipelineOuterClass.WorkContextCondition.Operand operand) {
            switch (operand) {
                case EQUALS -> {
                    return EQUALS;
                }
                default -> throw new IllegalArgumentException("Unrecognised operand %s".formatted(operand.name()));
            }
        }

        abstract boolean evaluate(WorkContext workContext, String key, String value);

    }

    public static Condition fromProtobuf(PipelineOuterClass.WorkContextCondition workContextCondition) {
        return new WorkContextCondition(
            workContextCondition.getKey(),
            Operand.fromProtobuf(workContextCondition.getOperand()),
            workContextCondition.getValue()
        );
    }
}
