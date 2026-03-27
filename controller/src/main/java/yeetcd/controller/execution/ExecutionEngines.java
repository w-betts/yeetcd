package yeetcd.controller.execution;

public class ExecutionEngines {
    public static ExecutionEngine createForRuntime() {
        return new DockerExecutionEngine();
    }
}
