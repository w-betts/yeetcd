package yeetcd.test;

import yeetcd.sdk.CustomWorkDefinition;
import yeetcd.sdk.DynamicWorkGeneratingWorkDefinition;

import java.io.IOException;
import java.time.Duration;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

public class PipelineTestRun {
    
    private final String pipelineName;
    private final String[] arguments;
    private final Duration timeout;
    private final BehaviorChain behaviorChain;

    private PipelineTestRun(Builder builder) {
        this.pipelineName = builder.pipelineName;
        this.arguments = builder.arguments;
        this.timeout = builder.timeout;
        this.behaviorChain = (BehaviorChain) builder.behavior;
    }

    public static Builder builder() {
        return new Builder();
    }

    public PipelineTestRunResult start() throws IOException, InterruptedException {
        MockServer mockServer = null;
        Process cliProcess = null;
        StringBuilder cliOutput = new StringBuilder();
        
        try {
            mockServer = new MockServer(0);
            mockServer.start();
            
            int port = mockServer.getPort();
            String mockAddress = "localhost:" + port;
            
            behaviorChain.registerWith(mockServer);
            
            cliProcess = runCLI(mockAddress);
            
            // Read CLI output
            try (var reader = new java.io.BufferedReader(new java.io.InputStreamReader(cliProcess.getInputStream()))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    cliOutput.append(line).append("\n");
                }
            }
            
            int exitCode = cliProcess.waitFor();
            
            List<WorkExecution> executions = mockServer.getExecutedWork();
            List<MatchedBehavior> matchedBehaviors = mockServer.getMatchedBehaviors();
            
            PipelineStatus status = exitCode == 0 ? PipelineStatus.SUCCESS : PipelineStatus.FAILURE;
            
            return new PipelineTestRunResult(status, exitCode, executions, matchedBehaviors, cliOutput.toString());
            
        } finally {
            if (cliProcess != null && cliProcess.isAlive()) {
                cliProcess.destroy();
            }
            if (mockServer != null) {
                mockServer.stop();
            }
        }
    }

    private Process runCLI(String mockAddress) throws IOException {
        String binaryPath = findBinary();
        
        List<String> command = new ArrayList<>();
        command.add(binaryPath);
        command.add("run");
        command.add("--source");
        command.add(System.getProperty("user.dir"));
        command.add("--pipeline");
        command.add(pipelineName);
        command.add("--mock-execution-engine-address");
        command.add(mockAddress);
        
        if (arguments != null) {
            for (String arg : arguments) {
                command.add("--argument");
                command.add(arg);
            }
        }
        
        ProcessBuilder pb = new ProcessBuilder(command);
        pb.environment().putAll(System.getenv());
        pb.redirectErrorStream(true);
        
        return pb.start();
    }

    private String findBinary() throws IOException {
        String binaryName = PlatformDetector.getBinaryName();
        
        // Try multiple resource paths
        String[] resourcePaths = {
            "/cli/" + binaryName,
            "/yeetcd/test/cli/" + binaryName  // For test resources
        };
        
        for (String resourcePath : resourcePaths) {
            if (getClass().getResource(resourcePath) != null) {
                return extractBinaryToTemp(resourcePath, binaryName);
            }
        }
        
        // Try working directory
        String workingDirBinary = "bin/" + binaryName;
        java.io.File file = new java.io.File(workingDirBinary);
        if (file.exists()) {
            return file.getAbsolutePath();
        }
        
        throw new IOException("CLI binary not found: " + binaryName);
    }

    private String extractBinaryToTemp(String resourcePath, String binaryName) throws IOException {
        try (var is = getClass().getResourceAsStream(resourcePath)) {
            if (is == null) {
                throw new IOException("Resource not found: " + resourcePath);
            }
            
            java.io.File cacheDir = new java.io.File(System.getProperty("user.home"), ".cache/yeetcd");
            cacheDir.mkdirs();
            java.io.File cachedBinary = new java.io.File(cacheDir, binaryName);
            
            java.nio.file.Files.copy(is, cachedBinary.toPath(), java.nio.file.StandardCopyOption.REPLACE_EXISTING);
            cachedBinary.setExecutable(true);
            
            return cachedBinary.getAbsolutePath();
        }
    }

    public static class Builder {
        private String pipelineName;
        private String[] arguments;
        private Duration timeout = Duration.ofSeconds(60);
        private Behavior behavior = new BehaviorChain(this);

        public Builder() {}

        public Builder pipelineName(String pipelineName) {
            this.pipelineName = pipelineName;
            return this;
        }

        public Builder arguments(String... arguments) {
            this.arguments = arguments;
            return this;
        }

        public Builder timeout(Duration timeout) {
            this.timeout = timeout;
            return this;
        }

        public ContainerisedWorkBehaviorBuilder containerisedWork(String image) {
            return ((BehaviorChain) behavior).containerisedWork(image);
        }

        public CustomWorkBehaviorBuilder customWork(CustomWorkDefinition instance) {
            return ((BehaviorChain) behavior).customWork(instance);
        }

        public DynamicWorkBehaviorBuilder dynamicWork() {
            return ((BehaviorChain) behavior).dynamicWork();
        }

        public DefaultWorkBehaviorBuilder defaultContainerisedWork() {
            return ((BehaviorChain) behavior).defaultContainerisedWork();
        }

        public DefaultWorkBehaviorBuilder defaultCustomWork() {
            return ((BehaviorChain) behavior).defaultCustomWork();
        }

        public PipelineTestRun build() {
            if (pipelineName == null || pipelineName.isEmpty()) {
                throw new IllegalStateException("pipelineName is required");
            }
            return new PipelineTestRun(this);
        }
    }
}
