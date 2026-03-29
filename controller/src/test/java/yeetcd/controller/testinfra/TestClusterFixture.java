package yeetcd.controller.testinfra;

import io.kubernetes.client.openapi.ApiClient;
import io.kubernetes.client.util.ClientBuilder;
import io.kubernetes.client.util.KubeConfig;
import lombok.Getter;
import lombok.extern.slf4j.Slf4j;
import org.junit.jupiter.api.extension.AfterAllCallback;
import org.junit.jupiter.api.extension.BeforeAllCallback;
import org.junit.jupiter.api.extension.ExtensionContext;

import java.io.InputStreamReader;
import java.nio.file.Path;
import java.time.Duration;
import java.util.concurrent.atomic.AtomicReference;

/**
 * JUnit 5 extension that manages test cluster lifecycle.
 * 
 * Defensive coding: Throws clear, actionable errors with full context on failure.
 * Handles beforeAll/afterAll lifecycle for cluster setup and cleanup.
 * 
 * Usage:
 * <pre>
 * @ExtendWith(TestClusterFixture.class)
 * public class MyIntegrationTest {
 *     // Tests run with cluster available
 * }
 * </pre>
 */
@Slf4j
public class TestClusterFixture implements BeforeAllCallback, AfterAllCallback {

    private static final String TEST_NAMESPACE = "yeetcd-test";
    private static final Duration STALE_RESOURCE_MAX_AGE = Duration.ofHours(2);

    @Getter
    private static K3dClusterManager clusterManager;
    
    @Getter
    private static TestResourceCleaner resourceCleaner;
    
    @Getter
    private static Path kubeconfigPath;
    
    @Getter
    private static ApiClient apiClient;
    
    /**
     * Port-forward process for RustFS.
     * Started on-demand when getRustFsEndpoint() is called.
     */
    private static AtomicReference<Process> rustfsPortForwardProcess = new AtomicReference<>();
    
    /**
     * Local port for RustFS port-forward.
     */
    private static final int RUSTFS_LOCAL_PORT = 19000;

    private static volatile boolean initialized = false;
    private static final Object initLock = new Object();

    /**
     * Called before all tests in a class.
     * Ensures cluster exists, cleans up stale resources, and creates API client.
     * 
     * Defensive: Throws clear error with context if any step fails.
     */
    @Override
    public void beforeAll(ExtensionContext context) throws Exception {
        String testClass = context.getRequiredTestClass().getName();
        log.info("Setting up test cluster for {}...", testClass);
        
        synchronized (initLock) {
            if (!initialized) {
                try {
                    initializeCluster();
                    initialized = true;
                } catch (TestInfrastructureException e) {
                    // Re-throw with additional context about the test class
                    throw new TestInfrastructureException(
                        e.getErrorCode(),
                        "Test cluster setup failed for " + testClass + ": " + e.getOperation(),
                        e.getResource(),
                        e.getState(),
                        e.getSuggestedFix(),
                        e
                    );
                } catch (Exception e) {
                    throw new TestInfrastructureException(
                        "CLUSTER_SETUP_UNEXPECTED_ERROR",
                        "Unexpected error during cluster setup for " + testClass,
                        "test cluster initialization",
                        "unexpected exception: " + e.getClass().getName() + ": " + e.getMessage(),
                        "Check logs for details. Ensure Docker is running and k3d is installed.",
                        e
                    );
                }
            }
        }
        
        // Clean up test resources before running tests (even if already initialized)
        try {
            cleanupBeforeTest();
        } catch (TestInfrastructureException e) {
            throw new TestInfrastructureException(
                e.getErrorCode(),
                "Pre-test cleanup failed for " + testClass + ": " + e.getOperation(),
                e.getResource(),
                e.getState(),
                e.getSuggestedFix(),
                e
            );
        }
        
        log.info("Test cluster setup complete for {}", testClass);
    }

    /**
     * Initializes the cluster and creates supporting infrastructure.
     * 
     * Defensive: Each step validates preconditions and throws clear errors.
     */
    private void initializeCluster() {
        log.info("Initializing test cluster...");
        
        // Step 1: Ensure cluster exists
        clusterManager = new K3dClusterManager();
        clusterManager.ensureClusterExists();
        
        // Step 2: Get kubeconfig
        kubeconfigPath = clusterManager.getKubeconfigPath();
        log.debug("Using kubeconfig: {}", kubeconfigPath);
        
        // Step 3: Create resource cleaner
        resourceCleaner = new TestResourceCleaner(kubeconfigPath);
        
        // Step 4: Create Kubernetes API client
        try {
            InputStreamReader kubeconfigReader = new InputStreamReader(
                new java.io.FileInputStream(kubeconfigPath.toFile())
            );
            apiClient = ClientBuilder.kubeconfig(KubeConfig.loadKubeConfig(kubeconfigReader)).build();
            log.debug("Created Kubernetes API client");
        } catch (Exception e) {
            throw new TestInfrastructureException(
                "API_CLIENT_CREATE_FAILED",
                "Failed to create Kubernetes API client during cluster initialization",
                "kubeconfig at '" + kubeconfigPath + "'",
                "exception: " + e.getMessage(),
                "Check kubeconfig is valid. Verify with 'kubectl --kubeconfig=" + kubeconfigPath + " version'",
                e
            );
        }
        
        log.info("Test cluster initialized successfully");
    }

    /**
     * Cleans up resources before running tests.
     * Detects and fixes bad state, cleans up stale resources.
     * 
     * Defensive: Logs operations, throws clear errors on failure.
     */
    private void cleanupBeforeTest() {
        log.info("Cleaning up before test run...");
        
        // Detect and fix bad state
        try {
            resourceCleaner.detectAndFixBadState(TEST_NAMESPACE);
        } catch (TestInfrastructureException e) {
            log.warn("Bad state detection/fix failed (continuing): {}", e.getMessage());
            // Don't fail - this is best-effort cleanup
        }
        
        // Clean up stale resources
        try {
            resourceCleaner.cleanupStaleResources(TEST_NAMESPACE, STALE_RESOURCE_MAX_AGE);
        } catch (TestInfrastructureException e) {
            log.warn("Stale resource cleanup failed (continuing): {}", e.getMessage());
            // Don't fail - this is best-effort cleanup
        }
        
        // Clean up current test resources
        try {
            resourceCleaner.cleanupTestResources(TEST_NAMESPACE);
        } catch (TestInfrastructureException e) {
            // This is more serious - current test resources should be cleanable
            throw new TestInfrastructureException(
                "TEST_CLEANUP_FAILED",
                "Failed to clean up current test resources",
                e.getResource(),
                e.getState(),
                "Check cluster connectivity and permissions. Try manual cleanup with 'kubectl delete pods,jobs -n " + TEST_NAMESPACE + " -l yeetcd-test=true'",
                e
            );
        }
        
        log.info("Pre-test cleanup complete");
    }

    /**
     * Called after all tests in a class.
     * Cleans up test resources but leaves cluster running for other tests.
     * 
     * Defensive: Logs operations, throws clear errors on failure.
     */
    @Override
    public void afterAll(ExtensionContext context) throws Exception {
        String testClass = context.getRequiredTestClass().getName();
        log.info("Tearing down test resources for {}...", testClass);
        
        // Stop RustFS port-forward
        stopRustFsPortForward();
        
        if (resourceCleaner != null) {
            try {
                resourceCleaner.cleanupTestResources(TEST_NAMESPACE);
                log.info("Test resources cleaned up for {}", testClass);
            } catch (TestInfrastructureException e) {
                throw new TestInfrastructureException(
                    e.getErrorCode(),
                    "Post-test cleanup failed for " + testClass + ": " + e.getOperation(),
                    e.getResource(),
                    e.getState(),
                    "Resources may be left behind. Clean up manually with 'kubectl delete pods,jobs -n " + TEST_NAMESPACE + " -l yeetcd-test=true'",
                    e
                );
            }
        } else {
            log.warn("Resource cleaner not initialized, skipping cleanup for {}", testClass);
        }
    }

    /**
     * Gets the registry port for pushing images.
     * Must be called after beforeAll.
     * 
     * Defensive: Validates preconditions, throws clear error if not ready.
     */
    public static int getRegistryPort() {
        if (clusterManager == null) {
            throw new TestInfrastructureException(
                "REGISTRY_NOT_AVAILABLE",
                "Registry port not available - cluster manager not initialized",
                "registry port",
                "cluster manager is null",
                "Ensure @ExtendWith(TestClusterFixture.class) is on your test class and beforeAll has completed"
            );
        }
        return clusterManager.getRegistryPort();
    }

    /**
     * Gets the registry address for pushing images (localhost:port).
     * Must be called after beforeAll.
     */
    public static String getRegistryPushAddress() {
        return "localhost:" + getRegistryPort();
    }

    /**
     * Gets the registry address for pulling images (from within cluster).
     * Must be called after beforeAll.
     */
    public static String getRegistryPullAddress() {
        return "yeetcd-registry:5000";
    }
    
    /**
     * Gets the RustFS endpoint for S3 operations from outside the cluster.
     * Starts port-forward if not already running.
     * Must be called after beforeAll.
     * 
     * @return the RustFS endpoint (e.g., "http://localhost:19000")
     */
    public static String getRustFsEndpoint() {
        ensureRustFsPortForward();
        return "http://localhost:" + RUSTFS_LOCAL_PORT;
    }
    
    /**
     * Gets the RustFS access key.
     * Must be called after beforeAll.
     */
    public static String getRustFsAccessKey() {
        return "rustfsadmin";
    }
    
    /**
     * Gets the RustFS secret key.
     * Must be called after beforeAll.
     */
    public static String getRustFsSecretKey() {
        return "rustfsadmin";
    }
    
    /**
     * Ensures port-forward for RustFS is running.
     * Starts it if not already running.
     */
    private static void ensureRustFsPortForward() {
        Process existing = rustfsPortForwardProcess.get();
        if (existing != null && existing.isAlive()) {
            log.debug("RustFS port-forward already running");
            return;
        }
        
        log.info("Starting port-forward for RustFS...");
        
        if (kubeconfigPath == null) {
            throw new TestInfrastructureException(
                "RUSTFS_PORT_FORWARD_NOT_READY",
                "RustFS port-forward not available - kubeconfig not initialized",
                "RustFS port-forward",
                "kubeconfig is null",
                "Ensure @ExtendWith(TestClusterFixture.class) is on your test class and beforeAll has completed"
            );
        }
        
        try {
            ProcessBuilder pb = new ProcessBuilder(
                "kubectl", "port-forward",
                "--kubeconfig", kubeconfigPath.toString(),
                "svc/yeetcd-rustfs-svc", "-n", "yeetcd",
                RUSTFS_LOCAL_PORT + ":9000"
            );
            pb.redirectErrorStream(true);
            
            Process process = pb.start();
            
            // Wait a moment for port-forward to establish
            Thread.sleep(1000);
            
            if (!process.isAlive()) {
                throw new TestInfrastructureException(
                    "RUSTFS_PORT_FORWARD_FAILED",
                    "Failed to start port-forward for RustFS",
                    "RustFS port-forward",
                    "process exited immediately",
                    "Check RustFS is running with 'kubectl get pods -n yeetcd -l app.kubernetes.io/name=rustfs'"
                );
            }
            
            rustfsPortForwardProcess.set(process);
            log.info("RustFS port-forward started on port {}", RUSTFS_LOCAL_PORT);
            
        } catch (Exception e) {
            throw new TestInfrastructureException(
                "RUSTFS_PORT_FORWARD_ERROR",
                "Error starting port-forward for RustFS",
                "RustFS port-forward",
                "exception: " + e.getMessage(),
                "Check kubectl is installed and RustFS is running",
                e
            );
        }
    }
    
    /**
     * Stops the RustFS port-forward if running.
     */
    private static void stopRustFsPortForward() {
        Process process = rustfsPortForwardProcess.getAndSet(null);
        if (process != null && process.isAlive()) {
            log.info("Stopping RustFS port-forward...");
            process.destroy();
            try {
                process.waitFor(5, java.util.concurrent.TimeUnit.SECONDS);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                process.destroyForcibly();
            }
            log.info("RustFS port-forward stopped");
        }
    }
}
