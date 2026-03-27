package yeetcd.controller.pipeline;

public final class DynamicWorkGeneratingWorkDefinition extends AbstractNativeWorkDefinition {
    private final String executionId;

    public DynamicWorkGeneratingWorkDefinition(String executionId) {
        this.executionId = executionId;
    }

    @Override
    public String executionId() {
        return executionId;
    }
}
