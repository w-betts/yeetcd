package yeetcd.test;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;

/**
 * Main class that orchestrates mock server lifecycle, behavior registration,
 * CLI spawning, and expectation verification.
 * 
 * Provides builder pattern with methods:
 * - startMockServer()
 * - defineBehavior()
 * - runCLI()
 * - verifyExpectations()
 */
public class YeetcdMockRunner {

    private final MockServer mockServer;
    private final List<MockBehavior> behaviors = new ArrayList<>();
    private final List<WorkExecution> executions = new ArrayList<>();
    private Process cliProcess;
    private String classpath;
    private String sourcePath;
    private String pipelineName;

    private YeetcdMockRunner(Builder builder) throws IOException {
        this.mockServer = new MockServer(builder.port);
        this.classpath = builder.classpath;
        this.sourcePath = builder.sourcePath;
        this.pipelineName = builder.pipelineName;
    }

    /**
     * Starts the embedded mock server.
     */
    public YeetcdMockRunner startMockServer() throws IOException {
        mockServer.start();
        return this;
    }

    /**
     * Registers a mock behavior.
     */
    public YeetcdMockRunner defineBehavior(MockBehavior behavior) {
        behaviors.add(behavior);
        
        // Register with mock server
        mockServer.registerBehavior(
            behavior.getImage(),
            behavior.getCmd(),
            behavior.toMockWorkResponse()
        );
        
        return this;
    }

    /**
     * Runs the CLI with the mock server.
     */
    public YeetcdMockRunner runCLI() throws IOException, InterruptedException {
        // Get the port the mock server is listening on
        String mockAddress = "localhost:" + mockServer.getPort();
        
        // Spawn CLI
        CLISpawner spawner = CLISpawner.builder()
                .classpath(classpath)
                .mockAddress(mockAddress)
                .build();
        
        cliProcess = spawner.spawn();
        
        // Wait for CLI to complete
        int exitCode = cliProcess.waitFor();
        
        // Capture executions from mock server (convert to WorkExecution)
        for (MockServer.WorkExecution exec : mockServer.getExecutedWork()) {
            executions.add(new WorkExecution(
                exec.image(),
                exec.cmd(),
                exec.envVars(),
                exec.workingDir(),
                exec.exitCode(),
                exec.stdout(),
                exec.stderr()
            ));
        }
        
        return this;
    }

    /**
     * Runs the CLI with custom arguments.
     */
    public YeetcdMockRunner runCLI(List<String> arguments) throws IOException, InterruptedException {
        // Similar to runCLI() but with custom arguments
        // For now, just delegate to basic runCLI
        return runCLI();
    }

    /**
     * Verifies expectations against recorded executions.
     */
    public ExpectationVerifier verifyExpectations() {
        return new ExpectationVerifier(executions);
    }

    /**
     * Gets all recorded work executions.
     */
    public List<WorkExecution> getExecutedWork() {
        return new ArrayList<>(executions);
    }

    /**
     * Gets the mock server.
     */
    public MockServer getMockServer() {
        return mockServer;
    }

    /**
     * Stops the mock server and cleans up.
     */
    public void close() throws InterruptedException {
        if (cliProcess != null && cliProcess.isAlive()) {
            cliProcess.destroy();
        }
        
        if (mockServer != null) {
            mockServer.stop();
        }
    }

    /**
     * Builder for YeetcdMockRunner.
     */
    public static class Builder {
        private int port = 50051;
        private String classpath;
        private String sourcePath;
        private String pipelineName;

        public Builder() {}

        public Builder port(int port) {
            this.port = port;
            return this;
        }

        public Builder classpath(String classpath) {
            this.classpath = classpath;
            return this;
        }

        public Builder sourcePath(String sourcePath) {
            this.sourcePath = sourcePath;
            return this;
        }

        public Builder pipelineName(String pipelineName) {
            this.pipelineName = pipelineName;
            return this;
        }

        public YeetcdMockRunner build() throws IOException {
            return new YeetcdMockRunner(this);
        }
    }

    /**
     * Creates a new Builder for YeetcdMockRunner.
     */
    public static Builder builder() {
        return new Builder();
    }
}