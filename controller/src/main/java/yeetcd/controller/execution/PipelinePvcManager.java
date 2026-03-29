package yeetcd.controller.execution;

import io.kubernetes.client.openapi.ApiClient;
import io.kubernetes.client.openapi.apis.CoreV1Api;
import io.kubernetes.client.openapi.models.V1PersistentVolumeClaim;
import io.kubernetes.client.openapi.models.V1PersistentVolumeClaimSpec;
import io.kubernetes.client.openapi.models.V1ObjectMeta;
import io.kubernetes.client.openapi.models.V1Pod;
import io.kubernetes.client.openapi.models.V1PodList;
import io.kubernetes.client.openapi.models.V1Volume;
import io.kubernetes.client.util.ClientBuilder;
import lombok.SneakyThrows;
import lombok.extern.slf4j.Slf4j;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

/**
 * Manages PVC lifecycle for pipeline runs.
 * Creates and deletes PVCs that are backed by S3 storage via k8s-csi-s3.
 */
@Slf4j
public class PipelinePvcManager {

    private static final String NAMESPACE = "yeetcd";
    private static final String PVC_PREFIX = "yeetcd-pvc-";
    private final CoreV1Api coreV1Api;

    public PipelinePvcManager() {
        try {
            this.coreV1Api = new CoreV1Api(ClientBuilder.cluster().build());
        } catch (IOException e) {
            throw new RuntimeException("Failed to create Kubernetes API client", e);
        }
    }

    public PipelinePvcManager(ApiClient apiClient) {
        this.coreV1Api = new CoreV1Api(apiClient);
    }

    /**
     * Creates a PVC for a pipeline run.
     * If a PVC with the same name already exists, it is deleted first.
     * 
     * @param pipelineRunId the unique identifier for the pipeline run
     * @param storageClassName the StorageClass to use (e.g., 'yeetcd-s3')
     * @return the name of the created PVC
     */
    @SneakyThrows
    public String createPvc(String pipelineRunId, String storageClassName) {
        String pvcName = PVC_PREFIX + pipelineRunId;
        
        log.info("Creating PVC '{}' for pipeline '{}' with storage class '{}'", 
            pvcName, pipelineRunId, storageClassName);
        
        // Delete existing PVC if it exists (for idempotency)
        try {
            coreV1Api.deleteNamespacedPersistentVolumeClaim(pvcName, NAMESPACE).execute();
            log.info("Deleted existing PVC '{}' before creating new one", pvcName);
            // Wait a moment for deletion to complete
            Thread.sleep(1000);
        } catch (io.kubernetes.client.openapi.ApiException e) {
            if (e.getCode() != 404) {
                throw e;
            }
            // PVC doesn't exist, which is fine
        }
        
        V1PersistentVolumeClaim pvc = new V1PersistentVolumeClaim()
            .metadata(new V1ObjectMeta()
                .name(pvcName)
                .labels(Map.of(
                    "app", "yeetcd",
                    "pipeline-run-id", pipelineRunId
                ))
            )
            .spec(new V1PersistentVolumeClaimSpec()
                .storageClassName(storageClassName)
                .accessModes(java.util.List.of("ReadWriteMany"))
                .resources(new io.kubernetes.client.openapi.models.V1VolumeResourceRequirements()
                    .requests(Map.of("storage", new io.kubernetes.client.custom.Quantity("1Gi")))
                )
            );
        
        coreV1Api.createNamespacedPersistentVolumeClaim(NAMESPACE, pvc).execute();
        
        log.info("PVC '{}' created, waiting for it to become Bound", pvcName);
        waitForPvcBound(pvcName, 30);
        
        log.info("PVC '{}' is now Bound", pvcName);
        return pvcName;
    }
    
    /**
     * Waits for a PVC to become Bound.
     * 
     * @param pvcName the name of the PVC
     * @param timeoutSeconds the maximum time to wait in seconds
     * @throws RuntimeException if the PVC doesn't become Bound within the timeout
     */
    @SneakyThrows
    public void waitForPvcBound(String pvcName, int timeoutSeconds) {
        long startTime = System.currentTimeMillis();
        long timeoutMs = timeoutSeconds * 1000L;
        
        while (System.currentTimeMillis() - startTime < timeoutMs) {
            String status = getPvcStatus(pvcName);
            if ("Bound".equals(status)) {
                return;
            }
            if (status == null) {
                throw new RuntimeException("PVC '" + pvcName + "' not found while waiting for Bound status");
            }
            log.debug("PVC '{}' status: {}, waiting...", pvcName, status);
            Thread.sleep(500);
        }
        
        String finalStatus = getPvcStatus(pvcName);
        throw new RuntimeException("PVC '" + pvcName + "' did not become Bound within " + timeoutSeconds + 
            " seconds. Final status: " + finalStatus);
    }

    /**
     * Deletes a PVC and waits for it to be fully removed.
     * First deletes any pods that are using the PVC to avoid the PVC protection finalizer blocking deletion.
     * 
     * @param pvcName the name of the PVC to delete
     */
    @SneakyThrows
    public void deletePvc(String pvcName) {
        log.info("Deleting PVC '{}'", pvcName);
        
        // First, delete any pods that are using this PVC
        deletePodsUsingPvc(pvcName);
        
        try {
            coreV1Api.deleteNamespacedPersistentVolumeClaim(pvcName, NAMESPACE).execute();
            log.info("PVC '{}' deletion initiated, waiting for removal", pvcName);
            
            // Wait for PVC to be fully deleted
            waitForPvcDeleted(pvcName, 30);
            log.info("PVC '{}' deleted successfully", pvcName);
        } catch (io.kubernetes.client.openapi.ApiException e) {
            if (e.getCode() == 404) {
                log.info("PVC '{}' not found, already deleted", pvcName);
            } else {
                throw e;
            }
        }
    }
    
    /**
     * Deletes all pods that are using the specified PVC.
     * This is necessary because the PVC protection finalizer prevents deletion while pods are using it.
     * 
     * @param pvcName the name of the PVC
     */
    @SneakyThrows
    private void deletePodsUsingPvc(String pvcName) {
        // Find all pods in the namespace
        V1PodList podList = coreV1Api.listNamespacedPod(NAMESPACE).execute();
        
        List<String> podsToDelete = new ArrayList<>();
        for (V1Pod pod : podList.getItems()) {
            if (pod.getSpec() != null && pod.getSpec().getVolumes() != null) {
                for (V1Volume volume : pod.getSpec().getVolumes()) {
                    if (volume.getPersistentVolumeClaim() != null &&
                        pvcName.equals(volume.getPersistentVolumeClaim().getClaimName())) {
                        podsToDelete.add(pod.getMetadata().getName());
                        break;
                    }
                }
            }
        }
        
        if (podsToDelete.isEmpty()) {
            log.debug("No pods found using PVC '{}'", pvcName);
            return;
        }
        
        log.info("Found {} pod(s) using PVC '{}', deleting them first", podsToDelete.size(), pvcName);
        
        // Delete each pod
        for (String podName : podsToDelete) {
            try {
                coreV1Api.deleteNamespacedPod(podName, NAMESPACE).execute();
                log.debug("Deleted pod '{}'", podName);
            } catch (io.kubernetes.client.openapi.ApiException e) {
                if (e.getCode() != 404) {
                    log.warn("Failed to delete pod '{}': {}", podName, e.getMessage());
                }
            }
        }
        
        // Wait for pods to be fully terminated
        for (String podName : podsToDelete) {
            waitForPodDeleted(podName, 30);
        }
        
        log.info("All pods using PVC '{}' have been deleted", pvcName);
    }
    
    /**
     * Waits for a pod to be fully deleted.
     * 
     * @param podName the name of the pod
     * @param timeoutSeconds the maximum time to wait in seconds
     */
    @SneakyThrows
    private void waitForPodDeleted(String podName, int timeoutSeconds) {
        long startTime = System.currentTimeMillis();
        long timeoutMs = timeoutSeconds * 1000L;
        
        while (System.currentTimeMillis() - startTime < timeoutMs) {
            try {
                coreV1Api.readNamespacedPod(podName, NAMESPACE).execute();
                log.debug("Pod '{}' still exists, waiting...", podName);
                Thread.sleep(500);
            } catch (io.kubernetes.client.openapi.ApiException e) {
                if (e.getCode() == 404) {
                    log.debug("Pod '{}' has been deleted", podName);
                    return;
                }
                throw e;
            }
        }
        
        log.warn("Pod '{}' was not deleted within {} seconds, continuing anyway", podName, timeoutSeconds);
    }
    
    /**
     * Waits for a PVC to be fully deleted.
     * 
     * @param pvcName the name of the PVC
     * @param timeoutSeconds the maximum time to wait in seconds
     * @throws RuntimeException if the PVC still exists after the timeout
     */
    @SneakyThrows
    public void waitForPvcDeleted(String pvcName, int timeoutSeconds) {
        long startTime = System.currentTimeMillis();
        long timeoutMs = timeoutSeconds * 1000L;
        
        while (System.currentTimeMillis() - startTime < timeoutMs) {
            String status = getPvcStatus(pvcName);
            if (status == null) {
                return; // PVC is deleted
            }
            log.debug("PVC '{}' still exists with status: {}, waiting...", pvcName, status);
            Thread.sleep(500);
        }
        
        throw new RuntimeException("PVC '" + pvcName + "' was not deleted within " + timeoutSeconds + " seconds");
    }

    /**
     * Gets the status of a PVC.
     * 
     * @param pvcName the name of the PVC
     * @return the status of the PVC (e.g., "Bound", "Pending"), or null if PVC doesn't exist
     */
    @SneakyThrows
    public String getPvcStatus(String pvcName) {
        try {
            V1PersistentVolumeClaim pvc = coreV1Api
                .readNamespacedPersistentVolumeClaim(pvcName, NAMESPACE)
                .execute();
            
            if (pvc.getStatus() != null && pvc.getStatus().getPhase() != null) {
                return pvc.getStatus().getPhase();
            }
            return "Pending";
        } catch (io.kubernetes.client.openapi.ApiException e) {
            if (e.getCode() == 404) {
                return null;
            }
            throw e;
        }
    }
    
    /**
     * Gets the UID of a PVC.
     * The UID is used as the S3 bucket name by the CSI driver.
     * 
     * @param pvcName the name of the PVC
     * @return the UID of the PVC, or null if PVC doesn't exist
     */
    @SneakyThrows
    public String getPvcUid(String pvcName) {
        try {
            V1PersistentVolumeClaim pvc = coreV1Api
                .readNamespacedPersistentVolumeClaim(pvcName, NAMESPACE)
                .execute();
            
            if (pvc.getMetadata() != null && pvc.getMetadata().getUid() != null) {
                return pvc.getMetadata().getUid();
            }
            return null;
        } catch (io.kubernetes.client.openapi.ApiException e) {
            if (e.getCode() == 404) {
                return null;
            }
            throw e;
        }
    }
}
