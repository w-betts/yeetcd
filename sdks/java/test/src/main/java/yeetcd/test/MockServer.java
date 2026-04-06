package yeetcd.test;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.stub.StreamObserver;
import yeetcd.protocol.mock.MockExecutionServiceGrpc;
import yeetcd.protocol.mock.Mock.MockImageBuildRequest;
import yeetcd.protocol.mock.Mock.MockImageBuildResponse;
import yeetcd.protocol.mock.Mock.MockWorkRequest;
import yeetcd.protocol.mock.Mock.MockWorkResponse;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.TimeUnit;
import java.util.logging.Logger;

public class MockServer {
    private static final Logger logger = Logger.getLogger(MockServer.class.getName());
    private static final String GENERATED_CUSTOM_WORK_RUNNER = "yeetcd.sdk.GeneratedCustomWorkRunner";
    private static final String GENERATED_PIPELINE_DEFINITIONS = "yeetcd.sdk.GeneratedPipelineDefinitions";

    private final int port;
    private Server server;
    private final List<WorkExecution> executions = new ArrayList<>();
    private final List<MatchedBehavior> matchedBehaviors = new ArrayList<>();
    private final Map<String, WorkResponse> containerisedBehaviors = new HashMap<>();
    private final Map<String, WorkResponse> customWorkBehaviors = new HashMap<>();
    private WorkResponse defaultContainerisedResponse = new WorkResponse(0, "", "");
    private WorkResponse defaultCustomResponse = new WorkResponse(0, "", "");
    private String classpath;

    public MockServer(int port) {
        this.port = port;
    }

    public void setClasspath(String classpath) {
        this.classpath = classpath;
    }

    public void start() throws IOException {
        server = ServerBuilder.forPort(port)
                .addService(new MockExecutionServiceImpl())
                .build()
                .start();

        logger.info("MockServer started on port " + port);

        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            try {
                MockServer.this.stop();
            } catch (InterruptedException e) {
                logger.severe("Error shutting down MockServer: " + e.getMessage());
            }
        }));
    }

    public void stop() throws InterruptedException {
        if (server != null) {
            server.shutdown().awaitTermination(5, TimeUnit.SECONDS);
            logger.info("MockServer stopped");
        }
    }

    public void registerContainerisedBehavior(String image, WorkResponse response) {
        containerisedBehaviors.put(image, response);
        logger.info("Registered containerised behavior for image: " + image);
    }

    public void registerCustomWorkBehavior(String executionId, WorkResponse response) {
        customWorkBehaviors.put(executionId, response);
        logger.info("Registered custom work behavior for executionId: " + executionId);
    }

    public void setDefaultContainerisedResponse(WorkResponse response) {
        this.defaultContainerisedResponse = response;
    }

    public void setDefaultCustomResponse(WorkResponse response) {
        this.defaultCustomResponse = response;
    }

    public List<WorkExecution> getExecutedWork() {
        return new ArrayList<>(executions);
    }

    public List<MatchedBehavior> getMatchedBehaviors() {
        return new ArrayList<>(matchedBehaviors);
    }

    public int getPort() {
        return port;
    }

    private class MockExecutionServiceImpl extends MockExecutionServiceGrpc.MockExecutionServiceImplBase {

        @Override
        public void runWork(MockWorkRequest request, StreamObserver<MockWorkResponse> responseObserver) {
            String image = request.getImage();
            String[] cmd = request.getCmdList().toArray(new String[0]);
            Map<String, String> envVars = request.getEnvVarsMap();
            String workingDir = request.getWorkingDir();

            WorkExecution execution;
            WorkResponse workResponse;
            WorkBehaviorType workType;
            String matchKey;

            if (isPipelineGenerator(cmd)) {
                workType = WorkBehaviorType.CONTAINERISED;
                matchKey = image;
                workResponse = runPipelineGenerator(cmd, workingDir, envVars);
                execution = WorkExecution.containerised(image, cmd, envVars, workingDir,
                        workResponse.exitCode(), workResponse.stdout(), workResponse.stderr());
            } else if (isNativeWork(cmd)) {
                String executionId = extractExecutionId(cmd);
                
                if (isDynamicWork(request)) {
                    workType = WorkBehaviorType.DYNAMIC;
                    matchKey = null;
                    
                    workResponse = defaultCustomResponse;
                    
                    execution = WorkExecution.dynamic(image, cmd, envVars, workingDir, 
                            workResponse.exitCode(), workResponse.stdout(), workResponse.stderr());
                } else {
                    workType = WorkBehaviorType.CUSTOM;
                    matchKey = executionId;
                    
                    workResponse = customWorkBehaviors.get(executionId);
                    if (workResponse == null) {
                        workResponse = defaultCustomResponse;
                    }
                    
                    execution = WorkExecution.custom(executionId, image, cmd, envVars, workingDir,
                            workResponse.exitCode(), workResponse.stdout(), workResponse.stderr());
                }
            } else {
                workType = WorkBehaviorType.CONTAINERISED;
                matchKey = image;
                
                workResponse = containerisedBehaviors.get(image);
                if (workResponse == null) {
                    workResponse = defaultContainerisedResponse;
                }
                
                execution = WorkExecution.containerised(image, cmd, envVars, workingDir,
                        workResponse.exitCode(), workResponse.stdout(), workResponse.stderr());
            }

            executions.add(execution);
            matchedBehaviors.add(new MatchedBehavior(workType, matchKey, execution, workResponse));

            MockWorkResponse response = MockWorkResponse.newBuilder()
                    .setExitCode(workResponse.exitCode())
                    .setStdout(workResponse.stdout())
                    .setStderr(workResponse.stderr())
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }

        private boolean isPipelineGenerator(String[] cmd) {
            if (cmd.length >= 1 && GENERATED_PIPELINE_DEFINITIONS.equals(cmd[0])) {
                return true;
            }
            return false;
        }

        private WorkResponse runPipelineGenerator(String[] cmd, String workingDir, Map<String, String> envVars) {
            if (classpath == null || classpath.isEmpty()) {
                logger.warning("No classpath set for pipeline generator");
                return new WorkResponse(1, "", "No classpath configured");
            }

            try {
                List<String> javaCmd = new ArrayList<>();
                javaCmd.add("java");
                javaCmd.add("-cp");
                javaCmd.add(classpath);
                for (String arg : cmd) {
                    javaCmd.add(arg);
                }

                ProcessBuilder pb = new ProcessBuilder(javaCmd);
                pb.environment().putAll(envVars);
                if (workingDir != null && !workingDir.isEmpty()) {
                    pb.directory(new java.io.File(workingDir));
                }
                pb.redirectErrorStream(true);

                Process process = pb.start();

                ByteArrayOutputStream stdout = new ByteArrayOutputStream();
                process.getInputStream().transferTo(stdout);

                boolean finished = process.waitFor(60, TimeUnit.SECONDS);
                if (!finished) {
                    process.destroyForcibly();
                    return new WorkResponse(1, "", "Generator timed out");
                }

                int exitCode = process.exitValue();
                String output = stdout.toString(StandardCharsets.UTF_8);

                logger.info("Pipeline generator completed with exit code " + exitCode);
                return new WorkResponse(exitCode, output, "");
            } catch (Exception e) {
                logger.severe("Failed to run pipeline generator: " + e.getMessage());
                return new WorkResponse(1, "", e.getMessage());
            }
        }

        private boolean isNativeWork(String[] cmd) {
            return cmd.length >= 1 && GENERATED_CUSTOM_WORK_RUNNER.equals(cmd[0]);
        }

        private String extractExecutionId(String[] cmd) {
            if (cmd.length >= 3) {
                return cmd[2];
            }
            return "";
        }

        private boolean isDynamicWork(MockWorkRequest request) {
            return false;
        }

        @Override
        public void buildImage(MockImageBuildRequest request, StreamObserver<MockImageBuildResponse> responseObserver) {
            MockImageBuildResponse response = MockImageBuildResponse.newBuilder()
                    .setSuccess(true)
                    .setImageRef(request.getImage() + ":" + request.getTag())
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        }
    }
}
