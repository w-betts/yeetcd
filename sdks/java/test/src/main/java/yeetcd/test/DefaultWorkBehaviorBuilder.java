package yeetcd.test;

public class DefaultWorkBehaviorBuilder {
    private final boolean forContainerised;
    private final BehaviorChain chain;

    DefaultWorkBehaviorBuilder(boolean forContainerised, BehaviorChain chain) {
        this.forContainerised = forContainerised;
        this.chain = chain;
    }

    public Behavior result(int exitCode, String stdout, String stderr) {
        if (forContainerised) {
            chain.setDefaultContainerisedResponse(new WorkResponse(exitCode, stdout, stderr));
        } else {
            chain.setDefaultCustomResponse(new WorkResponse(exitCode, stdout, stderr));
        }
        return chain;
    }
}
