package yeetcd.controller.execution;

import io.kubernetes.client.openapi.ApiClient;
import io.kubernetes.client.openapi.apis.CoreV1Api;
import io.kubernetes.client.openapi.models.V1PersistentVolumeClaim;
import io.kubernetes.client.openapi.models.V1PersistentVolumeClaimSpec;
import io.kubernetes.client.openapi.models.V1ObjectMeta;
import io.kubernetes.client.util.ClientBuilder;
import lombok.SneakyThrows;
import lombok.extern.slf4j.Slf4j;

import java.io.IOException;
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
        
        log.info("PVC '{}' created successfully", pvcName);
        return pvcName;
    }

    /**
     * Deletes a PVC.
     * 
     * @param pvcName the name of the PVC to delete
     */
    @SneakyThrows
    public void deletePvc(String pvcName) {
        log.info("Deleting PVC '{}'", pvcName);
        
        try {
            coreV1Api.deleteNamespacedPersistentVolumeClaim(pvcName, NAMESPACE).execute();
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
}
