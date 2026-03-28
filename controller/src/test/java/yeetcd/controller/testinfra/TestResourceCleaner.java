package yeetcd.controller.testinfra;

import io.kubernetes.client.openapi.ApiClient;
import io.kubernetes.client.openapi.ApiException;
import io.kubernetes.client.openapi.Configuration;
import io.kubernetes.client.openapi.apis.CoreV1Api;
import io.kubernetes.client.openapi.models.V1DeleteOptions;
import io.kubernetes.client.openapi.models.V1Pod;
import io.kubernetes.client.openapi.models.V1PodList;
import io.kubernetes.client.util.ClientBuilder;
import io.kubernetes.client.util.KubeConfig;
import lombok.extern.slf4j.Slf4j;

import java.io.InputStreamReader;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.Duration;
import java.time.Instant;
import java.time.OffsetDateTime;

/**
 * Cleans up test resources in Kubernetes.
 * 
 * Defensive coding: All operations throw clear, actionable errors with full context.
 * Validates inputs and preconditions before operations. Logs operations for debugging.
 */
@Slf4j
public class TestResourceCleaner {

    private static final String TEST_LABEL_KEY = "yeetcd-test";
    private static final String TEST_LABEL_VALUE = "true";
    private static final String TEST_LABEL_SELECTOR = TEST_LABEL_KEY + "=" + TEST_LABEL_VALUE;
    private static final Duration DEFAULT_MAX_AGE = Duration.ofHours(2);

    private final ApiClient apiClient;
    private final CoreV1Api coreApi;

    /**
     * Creates a TestResourceCleaner using the provided kubeconfig.
     * 
     * Defensive: Validates kubeconfig exists and is readable before creating client.
     */
    public TestResourceCleaner(Path kubeconfigPath) {
        validateKubeconfigPath(kubeconfigPath);
        
        log.debug("Creating TestResourceCleaner with kubeconfig: {}", kubeconfigPath);
        
        try {
            InputStreamReader kubeconfigReader = new InputStreamReader(Files.newInputStream(kubeconfigPath));
            this.apiClient = ClientBuilder.kubeconfig(KubeConfig.loadKubeConfig(kubeconfigReader)).build();
            Configuration.setDefaultApiClient(apiClient);
            this.coreApi = new CoreV1Api();
            log.debug("Successfully created Kubernetes API client");
        } catch (Exception e) {
            throw new TestInfrastructureException(
                "K8S_CLIENT_CREATE_FAILED",
                "Failed to create Kubernetes API client",
                "kubeconfig at '" + kubeconfigPath + "'",
                "ApiException: " + e.getMessage(),
                "Check kubeconfig is valid and cluster is accessible. Verify with 'kubectl --kubeconfig=" + kubeconfigPath + " version'",
                e
            );
        }
    }

    /**
     * Validates the kubeconfig path.
     * 
     * Defensive: Fails fast with clear error if kubeconfig is invalid.
     */
    private void validateKubeconfigPath(Path kubeconfigPath) {
        if (kubeconfigPath == null) {
            throw new TestInfrastructureException(
                "KUBECONFIG_PATH_NULL",
                "Kubeconfig path is null",
                "kubeconfig path",
                "null",
                "Provide a valid kubeconfig path (e.g., from K3dClusterManager.getKubeconfigPath())"
            );
        }
        
        if (!Files.exists(kubeconfigPath)) {
            throw new TestInfrastructureException(
                "KUBECONFIG_NOT_FOUND",
                "Kubeconfig file does not exist",
                "kubeconfig at '" + kubeconfigPath + "'",
                "file not found",
                "Ensure cluster is created with K3dClusterManager.ensureClusterExists() before creating TestResourceCleaner"
            );
        }
        
        if (!Files.isReadable(kubeconfigPath)) {
            throw new TestInfrastructureException(
                "KUBECONFIG_NOT_READABLE",
                "Kubeconfig file is not readable",
                "kubeconfig at '" + kubeconfigPath + "'",
                "file not readable",
                "Check file permissions with 'ls -l " + kubeconfigPath + "'"
            );
        }
    }

    /**
     * Cleans up test resources (jobs and pods) in the specified namespace.
     * Only removes resources with the test label.
     * 
     * Defensive: Validates namespace, throws clear errors with context on failure.
     */
    public void cleanupTestResources(String namespace) {
        validateNamespace(namespace);
        
        log.info("Cleaning up test resources in namespace '{}'...", namespace);
        
        cleanupJobsAndPods(namespace, TEST_LABEL_SELECTOR);
        
        log.info("Successfully cleaned up test resources in namespace '{}'", namespace);
    }

    /**
     * Cleans up stale test resources older than maxAge.
     * Useful for cleaning up resources from failed/aborted test runs.
     * 
     * Defensive: Validates inputs, throws clear errors with context.
     */
    public void cleanupStaleResources(String namespace, Duration maxAge) {
        validateNamespace(namespace);
        
        if (maxAge == null) {
            maxAge = DEFAULT_MAX_AGE;
            log.debug("Using default max age: {}", maxAge);
        }
        
        if (maxAge.isNegative() || maxAge.isZero()) {
            throw new TestInfrastructureException(
                "INVALID_MAX_AGE",
                "Max age must be positive",
                "maxAge parameter",
                "negative or zero: " + maxAge,
                "Provide a positive duration (e.g., Duration.ofHours(2))"
            );
        }
        
        log.info("Cleaning up stale resources in namespace '{}' older than {}...", namespace, maxAge);
        
        Instant cutoffTime = Instant.now().minus(maxAge);
        
        try {
            // Clean up stale pods
            V1PodList pods = coreApi.listNamespacedPod(namespace)
                .labelSelector(TEST_LABEL_SELECTOR)
                .execute();
            
            int deletedCount = 0;
            for (V1Pod pod : pods.getItems()) {
                if (isResourceStale(pod.getMetadata().getCreationTimestamp(), cutoffTime)) {
                    String podName = pod.getMetadata().getName();
                    log.debug("Deleting stale pod: {}", podName);
                    deletePod(namespace, podName);
                    deletedCount++;
                }
            }
            
            log.info("Deleted {} stale resources in namespace '{}'", deletedCount, namespace);
        } catch (ApiException e) {
            throw new TestInfrastructureException(
                "STALE_CLEANUP_FAILED",
                "Failed to clean up stale resources",
                "namespace '" + namespace + "'",
                "ApiException: " + e.getMessage() + " (code: " + e.getCode() + ")",
                "Check cluster connectivity and permissions. Verify with 'kubectl get pods -n " + namespace + "'",
                e
            );
        }
    }

    /**
     * Cleans up jobs and pods matching the label selector.
     * 
     * Defensive: Throws clear error with full context on failure.
     */
    public void cleanupJobsAndPods(String namespace, String labelSelector) {
        validateNamespace(namespace);
        
        if (labelSelector == null || labelSelector.isEmpty()) {
            throw new TestInfrastructureException(
                "INVALID_LABEL_SELECTOR",
                "Label selector cannot be null or empty",
                "labelSelector parameter",
                labelSelector == null ? "null" : "empty",
                "Provide a valid label selector (e.g., 'yeetcd-test=true')"
            );
        }
        
        log.debug("Cleaning up jobs and pods in namespace '{}' with selector '{}'...", namespace, labelSelector);
        
        try {
            // Delete pods with label selector
            V1PodList pods = coreApi.listNamespacedPod(namespace)
                .labelSelector(labelSelector)
                .execute();
            
            for (V1Pod pod : pods.getItems()) {
                String podName = pod.getMetadata().getName();
                log.debug("Deleting pod: {}", podName);
                deletePod(namespace, podName);
            }
            
            log.debug("Deleted {} pods", pods.getItems().size());
        } catch (ApiException e) {
            throw new TestInfrastructureException(
                "POD_CLEANUP_FAILED",
                "Failed to clean up pods",
                "pods in namespace '" + namespace + "' with selector '" + labelSelector + "'",
                "ApiException: " + e.getMessage() + " (code: " + e.getCode() + ")",
                "Check cluster connectivity and permissions. Verify with 'kubectl get pods -n " + namespace + " -l " + labelSelector + "'",
                e
            );
        }
    }

    /**
     * Deletes a specific pod.
     * 
     * Defensive: Throws clear error with context if deletion fails.
     */
    private void deletePod(String namespace, String podName) {
        try {
            V1DeleteOptions deleteOptions = new V1DeleteOptions();
            coreApi.deleteNamespacedPod(podName, namespace)
                .body(deleteOptions)
                .execute();
            log.debug("Deleted pod: {}", podName);
        } catch (ApiException e) {
            if (e.getCode() == 404) {
                log.debug("Pod '{}' already deleted (404)", podName);
                return;
            }
            throw new TestInfrastructureException(
                "POD_DELETE_FAILED",
                "Failed to delete pod",
                "pod '" + podName + "' in namespace '" + namespace + "'",
                "ApiException: " + e.getMessage() + " (code: " + e.getCode() + ")",
                "Check pod exists with 'kubectl get pod " + podName + " -n " + namespace + "'. Check permissions.",
                e
            );
        }
    }

    /**
     * Checks if a resource is stale based on its creation timestamp.
     */
    private boolean isResourceStale(OffsetDateTime creationTimestamp, Instant cutoffTime) {
        if (creationTimestamp == null) {
            return false;
        }
        return creationTimestamp.toInstant().isBefore(cutoffTime);
    }

    /**
     * Validates the namespace parameter.
     * 
     * Defensive: Fails fast with clear error for invalid input.
     */
    private void validateNamespace(String namespace) {
        if (namespace == null || namespace.isEmpty()) {
            throw new TestInfrastructureException(
                "INVALID_NAMESPACE",
                "Namespace cannot be null or empty",
                "namespace parameter",
                namespace == null ? "null" : "empty",
                "Provide a valid namespace (e.g., 'default', 'yeetcd-test')"
            );
        }
    }

    /**
     * Detects and auto-fixes bad state by cleaning up stuck resources.
     * This is useful when tests leave resources in an inconsistent state.
     * 
     * Defensive: Logs all operations, throws clear errors with context.
     */
    public void detectAndFixBadState(String namespace) {
        validateNamespace(namespace);
        
        log.info("Detecting and fixing bad state in namespace '{}'...", namespace);
        
        try {
            // Check for pods in bad states (e.g., Evicted, NodeLost)
            V1PodList allPods = coreApi.listNamespacedPod(namespace).execute();
            
            int fixedCount = 0;
            for (V1Pod pod : allPods.getItems()) {
                String phase = pod.getStatus() != null ? pod.getStatus().getPhase() : null;
                String podName = pod.getMetadata().getName();
                
                // Check for bad states
                if ("Failed".equals(phase) || "Unknown".equals(phase) || 
                    (pod.getStatus() != null && pod.getStatus().getReason() != null && 
                     (pod.getStatus().getReason().contains("Evicted") || 
                      pod.getStatus().getReason().contains("NodeLost")))) {
                    
                    log.warn("Found pod '{}' in bad state (phase={}, reason={}), deleting...", 
                        podName, phase, 
                        pod.getStatus() != null ? pod.getStatus().getReason() : "null");
                    
                    try {
                        deletePod(namespace, podName);
                        fixedCount++;
                    } catch (TestInfrastructureException e) {
                        log.error("Failed to delete bad pod '{}': {}", podName, e.getMessage());
                        // Continue trying to fix other resources
                    }
                }
            }
            
            log.info("Fixed {} resources in bad state in namespace '{}'", fixedCount, namespace);
        } catch (ApiException e) {
            throw new TestInfrastructureException(
                "BAD_STATE_DETECTION_FAILED",
                "Failed to detect and fix bad state",
                "namespace '" + namespace + "'",
                "ApiException: " + e.getMessage() + " (code: " + e.getCode() + ")",
                "Check cluster connectivity with 'kubectl get pods -n " + namespace + "'",
                e
            );
        }
    }
}
