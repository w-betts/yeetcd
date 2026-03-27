package yeetcd.controller.execution;

public class DockerExecutionEngineTest extends AbstractExecutionEngineTest {
    @Override
    ExecutionEngine executionEngine() {
        return new DockerExecutionEngine();
    }

    @Override
    String builtImagePullAddress() {
        return "";
    }
}
