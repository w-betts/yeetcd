package yeetcd.test;

public class ContainerisedWorkBehaviorBuilder {
    private final String image;
    private final BehaviorChain chain;

    ContainerisedWorkBehaviorBuilder(String image, BehaviorChain chain) {
        this.image = image;
        this.chain = chain;
    }

    public Behavior result(int exitCode, String stdout, String stderr) {
        chain.addContainerisedWorkBehavior(image, null, new WorkResponse(exitCode, stdout, stderr));
        return chain;
    }

    public Behavior replace(String replacementImage, String... replacementCmd) {
        chain.addContainerisedWorkBehavior(image, new ReplacementImage(replacementImage, replacementCmd), null);
        return chain;
    }

    public record ReplacementImage(String image, String[] cmd) {}
}
