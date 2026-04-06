package yeetcd.test;

import yeetcd.sdk.DynamicWorkGeneratingWorkDefinition;

public class DynamicWorkBehaviorBuilder {
    private final BehaviorChain chain;

    DynamicWorkBehaviorBuilder(BehaviorChain chain) {
        this.chain = chain;
    }

    public Behavior replace(DynamicWorkGeneratingWorkDefinition instance) {
        chain.setDynamicWorkBehavior(instance);
        return chain;
    }
}
