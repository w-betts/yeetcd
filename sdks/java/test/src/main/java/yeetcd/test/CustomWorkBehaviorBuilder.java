package yeetcd.test;

import yeetcd.sdk.CustomWorkDefinition;

public class CustomWorkBehaviorBuilder {
    private final CustomWorkDefinition instance;
    private final String executionId;
    private final BehaviorChain chain;

    CustomWorkBehaviorBuilder(CustomWorkDefinition instance, String executionId, BehaviorChain chain) {
        this.instance = instance;
        this.executionId = executionId;
        this.chain = chain;
    }

    public Behavior result(int exitCode, String stdout, String stderr) {
        chain.addCustomWorkBehavior(executionId, new WorkResponse(exitCode, stdout, stderr), null, false);
        return chain;
    }

    public Behavior replace(CustomWorkDefinition replacement) {
        chain.addCustomWorkBehavior(executionId, null, replacement, false);
        return chain;
    }

    public Behavior run() {
        chain.addCustomWorkBehavior(executionId, null, instance, true);
        return chain;
    }
}
