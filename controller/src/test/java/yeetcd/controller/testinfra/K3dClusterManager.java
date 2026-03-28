package yeetcd.controller.testinfra;

import lombok.extern.slf4j.Slf4j;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.concurrent.TimeUnit;

/**
 * Manages k3d cluster lifecycle for integration tests.
 * 
 * Defensive coding: All operations throw clear, actionable errors with full context.
 * Never swallow exceptions or fail silently.
 */
@Slf4j
public class K3dClusterManager {

    private static final String CLUSTER_NAME = "yeetcd";
    private static final int K3D_TIMEOUT_SECONDS = 120;
    private static final int REGISTRY_PORT_TIMEOUT_SECONDS = 30;

    private final Path kubeconfigPath;
    private volatile boolean clusterChecked = false;

    public K3dClusterManager() {
        this.kubeconfigPath = Path.of(System.getProperty("java.io.tmpdir"), "yeetcd-k3d-kubeconfig.yaml");
    }

    /**
     * Ensures the k3d cluster exists, creating it if necessary.
     * 
     * Defensive: Validates all preconditions before operations.
     * Throws clear errors with context (operation, resource, state, suggested fix).
     */
    public synchronized void ensureClusterExists() {
        if (clusterChecked) {
            log.debug("Cluster already checked, skipping");
            return;
        }

        log.info("Ensuring k3d cluster '{}' exists...", CLUSTER_NAME);

        // Validate k3d is installed
        validateK3dInstalled();

        // Check if cluster exists
        boolean clusterExists = checkClusterExists();
        
        if (!clusterExists) {
            log.info("Cluster '{}' does not exist, creating...", CLUSTER_NAME);
            createCluster();
        } else {
            log.info("Cluster '{}' already exists", CLUSTER_NAME);
        }

        // Ensure kubeconfig is available
        updateKubeconfig();
        
        // Validate cluster is accessible
        validateClusterAccessible();

        clusterChecked = true;
        log.info("Cluster '{}' is ready", CLUSTER_NAME);
    }

    /**
     * Validates that k3d is installed and available on PATH.
     * 
     * Defensive: Fails fast with actionable error message if k3d is not installed.
     */
    private void validateK3dInstalled() {
        log.debug("Validating k3d installation...");
        
        ProcessResult result = executeCommand("k3d", "version");
        
        if (result.exitCode != 0) {
            throw new TestInfrastructureException(
                "K3D_NOT_INSTALLED",
                "k3d is not installed or not on PATH",
                "k3d binary",
                "not found or not executable",
                "Install k3d: https://k3d.io/v5.7.4/#installation (e.g., 'brew install k3d' on macOS, 'curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash' on Linux)"
            );
        }
        
        log.debug("k3d version: {}", result.stdout.trim());
    }

    /**
     * Checks if the k3d cluster exists.
     * 
     * Defensive: Throws clear error if cluster check fails unexpectedly.
     */
    private boolean checkClusterExists() {
        log.debug("Checking if cluster '{}' exists...", CLUSTER_NAME);
        
        ProcessResult result = executeCommand("k3d", "cluster", "list", "--no-headers");
        
        if (result.exitCode != 0) {
            throw new TestInfrastructureException(
                "K3D_CLUSTER_LIST_FAILED",
                "Failed to list k3d clusters",
                "k3d cluster list command",
                "exit code " + result.exitCode,
                "Check k3d installation and Docker daemon. Error: " + result.stderr
            );
        }
        
        boolean exists = result.stdout.contains(CLUSTER_NAME);
        log.debug("Cluster '{}' exists: {}", CLUSTER_NAME, exists);
        return exists;
    }

    /**
     * Creates the k3d cluster with a registry.
     * 
     * Defensive: Throws clear error with full context if creation fails.
     */
    private void createCluster() {
        log.info("Creating k3d cluster '{}' with registry...", CLUSTER_NAME);
        
        // Build command with registry
        ProcessResult result = executeCommand(
            "k3d", "cluster", "create", CLUSTER_NAME,
            "--registry-create", CLUSTER_NAME + "-registry",
            "--agents", "1",
            "--wait"
        );
        
        if (result.exitCode != 0) {
            throw new TestInfrastructureException(
                "K3D_CLUSTER_CREATE_FAILED",
                "Failed to create k3d cluster",
                "k3d cluster '" + CLUSTER_NAME + "'",
                "creation failed with exit code " + result.exitCode,
                "Check Docker is running and ports are available. Delete existing cluster with 'k3d cluster delete " + CLUSTER_NAME + "' and retry. Error: " + result.stderr
            );
        }
        
        log.info("Successfully created cluster '{}'", CLUSTER_NAME);
    }

    /**
     * Updates the kubeconfig file for the cluster.
     * 
     * Defensive: Throws clear error if kubeconfig cannot be retrieved or written.
     */
    private void updateKubeconfig() {
        log.debug("Updating kubeconfig at {}...", kubeconfigPath);
        
        ProcessResult result = executeCommand("k3d", "kubeconfig", "get", CLUSTER_NAME);
        
        if (result.exitCode != 0) {
            throw new TestInfrastructureException(
                "K3D_KUBECONFIG_FAILED",
                "Failed to get kubeconfig from k3d",
                "kubeconfig for cluster '" + CLUSTER_NAME + "'",
                "k3d kubeconfig get failed with exit code " + result.exitCode,
                "Check cluster exists with 'k3d cluster list'. Error: " + result.stderr
            );
        }
        
        try {
            Files.writeString(kubeconfigPath, result.stdout, StandardCharsets.UTF_8);
            log.debug("Kubeconfig written to {}", kubeconfigPath);
        } catch (IOException e) {
            throw new TestInfrastructureException(
                "KUBECONFIG_WRITE_FAILED",
                "Failed to write kubeconfig file",
                "kubeconfig file at '" + kubeconfigPath + "'",
                "IOException: " + e.getMessage(),
                "Check file permissions and disk space. Ensure directory exists: " + kubeconfigPath.getParent(),
                e
            );
        }
    }

    /**
     * Validates that the cluster is accessible via kubectl.
     * 
     * Defensive: Throws clear error if cluster is not responding.
     */
    private void validateClusterAccessible() {
        log.debug("Validating cluster accessibility...");
        
        ProcessResult result = executeCommandWithEnv(
            new String[]{"KUBECONFIG", kubeconfigPath.toString()},
            "kubectl", "version"
        );
        
        if (result.exitCode != 0) {
            throw new TestInfrastructureException(
                "CLUSTER_NOT_ACCESSIBLE",
                "Cluster is not accessible via kubectl",
                "cluster '" + CLUSTER_NAME + "'",
                "kubectl version failed with exit code " + result.exitCode,
                "Check cluster is running with 'k3d cluster list'. Kubeconfig at: " + kubeconfigPath + ". Error: " + result.stderr
            );
        }
        
        log.debug("Cluster is accessible");
    }

    /**
     * Gets the path to the kubeconfig file.
     * Must call ensureClusterExists() first.
     * 
     * Defensive: Validates preconditions before returning.
     */
    public Path getKubeconfigPath() {
        if (!clusterChecked) {
            throw new TestInfrastructureException(
                "KUBECONFIG_NOT_READY",
                "Kubeconfig not available - cluster not initialized",
                "kubeconfig path",
                "cluster not checked",
                "Call ensureClusterExists() before getKubeconfigPath()"
            );
        }
        
        if (!Files.exists(kubeconfigPath)) {
            throw new TestInfrastructureException(
                "KUBECONFIG_MISSING",
                "Kubeconfig file does not exist",
                "kubeconfig at '" + kubeconfigPath + "'",
                "file not found",
                "Call ensureClusterExists() to regenerate kubeconfig"
            );
        }
        
        return kubeconfigPath;
    }

    /**
     * Gets the registry port for the cluster's registry.
     * Must call ensureClusterExists() first.
     * 
     * Defensive: Throws clear error with context if registry port cannot be determined.
     */
    public int getRegistryPort() {
        if (!clusterChecked) {
            throw new TestInfrastructureException(
                "REGISTRY_NOT_READY",
                "Registry port not available - cluster not initialized",
                "registry port",
                "cluster not checked",
                "Call ensureClusterExists() before getRegistryPort()"
            );
        }
        
        String registryName = CLUSTER_NAME + "-registry";
        log.debug("Getting registry port for {}...", registryName);
        
        // Use docker inspect to get the registry port
        ProcessResult result = executeCommand(
            "docker", "inspect", "--format",
            "{{index .NetworkSettings.Ports \"5000/tcp\" 0 \"HostPort\"}}",
            registryName
        );
        
        if (result.exitCode != 0 || result.stdout.trim().isEmpty()) {
            throw new TestInfrastructureException(
                "REGISTRY_PORT_FAILED",
                "Failed to get registry port",
                "registry container '" + registryName + "'",
                "docker inspect failed with exit code " + result.exitCode + ", stderr: " + result.stderr,
                "Check Docker is running and registry container exists with 'docker ps'. Ensure cluster was created with registry using 'k3d cluster create --registry-create'"
            );
        }
        
        try {
            int port = Integer.parseInt(result.stdout.trim());
            log.debug("Registry port: {}", port);
            return port;
        } catch (NumberFormatException e) {
            throw new TestInfrastructureException(
                "REGISTRY_PORT_INVALID",
                "Registry port is not a valid number",
                "registry port output",
                "invalid format: '" + result.stdout.trim() + "'",
                "Check registry container is healthy with 'docker logs " + registryName + "'",
                e
            );
        }
    }

    /**
     * Deletes the k3d cluster.
     * 
     * Defensive: Logs operations for debugging, throws clear errors on failure.
     */
    public void deleteCluster() {
        log.info("Deleting k3d cluster '{}'...", CLUSTER_NAME);
        
        ProcessResult result = executeCommand("k3d", "cluster", "delete", CLUSTER_NAME);
        
        if (result.exitCode != 0) {
            throw new TestInfrastructureException(
                "K3D_CLUSTER_DELETE_FAILED",
                "Failed to delete k3d cluster",
                "cluster '" + CLUSTER_NAME + "'",
                "delete failed with exit code " + result.exitCode,
                "Check Docker is running and cluster exists. Error: " + result.stderr
            );
        }
        
        clusterChecked = false;
        log.info("Successfully deleted cluster '{}'", CLUSTER_NAME);
    }

    /**
     * Executes a command and returns the result.
     * 
     * Defensive: Throws clear error if command cannot be executed.
     */
    private ProcessResult executeCommand(String... command) {
        return executeCommandWithEnv(null, command);
    }

    /**
     * Executes a command with environment variables and returns the result.
     * 
     * Defensive: Throws clear error with full context if command fails to execute.
     */
    private ProcessResult executeCommandWithEnv(String[] envVar, String... command) {
        String commandStr = String.join(" ", command);
        log.debug("Executing: {}", commandStr);
        
        ProcessBuilder pb = new ProcessBuilder(command);
        pb.redirectErrorStream(true);
        
        if (envVar != null && envVar.length == 2) {
            pb.environment().put(envVar[0], envVar[1]);
        }
        
        Process process;
        try {
            process = pb.start();
        } catch (IOException e) {
            throw new TestInfrastructureException(
                "COMMAND_EXECUTION_FAILED",
                "Failed to execute command",
                "command '" + commandStr + "'",
                "IOException: " + e.getMessage(),
                "Check the command is installed and on PATH",
                e
            );
        }
        
        StringBuilder stdout = new StringBuilder();
        StringBuilder stderr = new StringBuilder();
        
        try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8))) {
            String line;
            while ((line = reader.readLine()) != null) {
                stdout.append(line).append("\n");
            }
        } catch (IOException e) {
            throw new TestInfrastructureException(
                "COMMAND_OUTPUT_READ_FAILED",
                "Failed to read command output",
                "command '" + commandStr + "'",
                "IOException: " + e.getMessage(),
                "This is unexpected - check system resources",
                e
            );
        }
        
        boolean finished;
        try {
            finished = process.waitFor(K3D_TIMEOUT_SECONDS, TimeUnit.SECONDS);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw new TestInfrastructureException(
                "COMMAND_INTERRUPTED",
                "Command execution was interrupted",
                "command '" + commandStr + "'",
                "InterruptedException",
                "Retry the operation",
                e
            );
        }
        
        if (!finished) {
            process.destroyForcibly();
            throw new TestInfrastructureException(
                "COMMAND_TIMEOUT",
                "Command timed out",
                "command '" + commandStr + "'",
                "timeout after " + K3D_TIMEOUT_SECONDS + " seconds",
                "Check if Docker/k3d is responsive. Consider increasing timeout or checking system resources."
            );
        }
        
        int exitCode = process.exitValue();
        log.debug("Command '{}' exited with code {}", commandStr, exitCode);
        
        return new ProcessResult(exitCode, stdout.toString(), stderr.toString());
    }

    private record ProcessResult(int exitCode, String stdout, String stderr) {}
}
