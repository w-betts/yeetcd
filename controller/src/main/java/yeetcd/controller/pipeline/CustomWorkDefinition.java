package yeetcd.controller.pipeline;

public final class CustomWorkDefinition extends AbstractNativeWorkDefinition {
    private final String executionId;

    public CustomWorkDefinition(String executionId) {
        this.executionId = executionId;
    }

    public String executionId() {
        return executionId;
    }
}
