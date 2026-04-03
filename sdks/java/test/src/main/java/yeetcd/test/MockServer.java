package yeetcd.test;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.stub.StreamObserver;
import yeetcd.protocol.mock.MockExecutionServiceGrpc;
import yeetcd.protocol.mock.Mock.MockImageBuildRequest;
import yeetcd.protocol.mock.Mock.MockImageBuildResponse;
import yeetcd.protocol.mock.Mock.MockWorkRequest;
import yeetcd.protocol.mock.Mock.MockWorkResponse;

import java.io.IOException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.TimeUnit;
import java.util.logging.Logger;

/**
 * MockServer implements the gRPC MockExecutionService for testing.
 * It runs embedded in the test JVM and accepts behavior definitions to return mock responses.
 */
public class MockServer {
    private static final Logger logger = Logger.getLogger(MockServer.class.getName());
    
    private final int port;
    private Server server;
    private final Map<String, MockWorkResponse> behaviors = new HashMap<>();
    private final List<WorkExecution> executions = new ArrayList<>();

    /**
     * Creates a MockServer on the specified port.
     */
    public MockServer(int port) {
        this.port = port;
    }

    /**
     * Starts the gRPC server.
     */
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

    /**
     * Stops the gRPC server.
     */
    public void stop() throws InterruptedException {
        if (server != null) {
            server.shutdown().awaitTermination(5, TimeUnit.SECONDS);
            logger.info("MockServer stopped");
        }
    }

    /**
     * Registers a mock behavior for matching work requests.
     * 
     * @param image the image to match
     * @param cmd the command to match
     * @param response the mock response to return
     */
    public void registerBehavior(String image, String[] cmd, MockWorkResponse response) {
        String key = image + ":" + String.join(" ", cmd);
        behaviors.put(key, response);
        logger.info("Registered behavior for: " + key);
    }

    /**
     * Gets all recorded work executions.
     */
    public List<WorkExecution> getExecutedWork() {
        return new ArrayList<>(executions);
    }

    /**
     * Gets the port the server is listening on.
     */
    public int getPort() {
        return port;
    }

    /**
     * Record representing a work execution.
     */
    public record WorkExecution(
        String image,
        String[] cmd,
        Map<String, String> envVars,
        String workingDir,
        int exitCode,
        String stdout,
        String stderr
    ) {}

    /**
     * gRPC service implementation.
     */
    private class MockExecutionServiceImpl extends MockExecutionServiceGrpc.MockExecutionServiceImplBase {
        
        @Override
        public void runWork(MockWorkRequest request, StreamObserver<MockWorkResponse> responseObserver) {
            // Record the execution
            WorkExecution execution = new WorkExecution(
                request.getImage(),
                request.getCmdList().toArray(new String[0]),
                request.getEnvVarsMap(),
                request.getWorkingDir(),
                0,  // Will be set from response
                "",
                ""
            );
            executions.add(execution);
            
            // Find matching behavior
            String key = request.getImage() + ":" + String.join(" ", request.getCmdList());
            MockWorkResponse response = behaviors.get(key);
            
            if (response == null) {
                // Default response if no behavior matched
                response = MockWorkResponse.newBuilder()
                        .setExitCode(0)
                        .setStdout("mock response")
                        .build();
            }
            
            responseObserver.onNext(response);
            responseObserver.onCompleted();
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

    /**
     * Builder for MockServer.
     */
    public static class Builder {
        private int port = 50051;

        public Builder port(int port) {
            this.port = port;
            return this;
        }

        public MockServer build() {
            return new MockServer(port);
        }
    }
}