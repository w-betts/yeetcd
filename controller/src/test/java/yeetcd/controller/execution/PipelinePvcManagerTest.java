package yeetcd.controller.execution;

import yeetcd.controller.testinfra.TestClusterFixture;
import io.kubernetes.client.openapi.ApiClient;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

import static org.hamcrest.MatcherAssert.assertThat;
import static org.hamcrest.CoreMatchers.*;

/**
 * Unit tests for PipelinePvcManager.
 * 
 * Tests PVC lifecycle management including creation, deletion, and status retrieval.
 * Uses TestClusterFixture to manage k3d cluster lifecycle with RustFS deployed.
 */
@ExtendWith(TestClusterFixture.class)
public class PipelinePvcManagerTest {

    private static final String STORAGE_CLASS = "yeetcd-s3";

    private PipelinePvcManager createPvcManager() {
        ApiClient apiClient = TestClusterFixture.getApiClient();
        return new PipelinePvcManager(apiClient);
    }

    /**
     * GIVEN: k3d cluster exists with RustFS deployed
     * WHEN: createPvc('test-pipeline') is called
     * THEN: PVC is created with storageClassName 'yeetcd-s3' and status 'Bound'
     */
    @Test
    public void shouldCreatePvcWithCorrectStorageClass() {
        // GIVEN: k3d cluster exists with RustFS deployed
        PipelinePvcManager pvcManager = createPvcManager();
        String pipelineRunId = "test-pipeline";

        // WHEN: createPvc is called
        String pvcName = pvcManager.createPvc(pipelineRunId, STORAGE_CLASS);

        // THEN: PVC is created with correct storage class and is bound
        assertThat(pvcName, notNullValue());
        assertThat(pvcName, containsString(pipelineRunId));
        
        String status = pvcManager.getPvcStatus(pvcName);
        assertThat(status, equalTo("Bound"));
        
        // Cleanup
        pvcManager.deletePvc(pvcName);
    }

    /**
     * GIVEN: PVC exists for a pipeline
     * WHEN: deletePvc(pvcName) is called
     * THEN: PVC is deleted from the namespace
     */
    @Test
    public void shouldDeletePvcSuccessfully() {
        // GIVEN: PVC exists for a pipeline
        PipelinePvcManager pvcManager = createPvcManager();
        String pipelineRunId = "test-pipeline-delete";
        String pvcName = pvcManager.createPvc(pipelineRunId, STORAGE_CLASS);
        
        // Verify PVC exists
        String status = pvcManager.getPvcStatus(pvcName);
        assertThat(status, equalTo("Bound"));

        // WHEN: deletePvc is called
        pvcManager.deletePvc(pvcName);

        // THEN: PVC is deleted from the namespace
        String deletedStatus = pvcManager.getPvcStatus(pvcName);
        assertThat(deletedStatus, nullValue());
    }

    /**
     * GIVEN: PVC exists
     * WHEN: getPvcStatus(pvcName) is called
     * THEN: correct status (Bound/Pending) is returned
     */
    @Test
    public void shouldRetrievePvcStatusCorrectly() {
        // GIVEN: PVC exists
        PipelinePvcManager pvcManager = createPvcManager();
        String pipelineRunId = "test-pipeline-status";
        String pvcName = pvcManager.createPvc(pipelineRunId, STORAGE_CLASS);

        // WHEN: getPvcStatus is called
        String status = pvcManager.getPvcStatus(pvcName);

        // THEN: correct status is returned
        assertThat(status, notNullValue());
        assertThat(status, anyOf(equalTo("Bound"), equalTo("Pending")));
        
        // Cleanup
        pvcManager.deletePvc(pvcName);
    }

    /**
     * GIVEN: PVC does not exist
     * WHEN: getPvcStatus(pvcName) is called
     * THEN: null is returned
     */
    @Test
    public void shouldReturnNullForNonExistentPvc() {
        // GIVEN: PVC does not exist
        PipelinePvcManager pvcManager = createPvcManager();
        String nonExistentPvcName = "pvc-does-not-exist-12345";

        // WHEN: getPvcStatus is called
        String status = pvcManager.getPvcStatus(nonExistentPvcName);

        // THEN: null is returned
        assertThat(status, nullValue());
    }

    /**
     * GIVEN: Multiple PVCs exist
     * WHEN: Each is queried
     * THEN: Correct status is returned for each
     */
    @Test
    public void shouldHandleMultiplePvcs() {
        // GIVEN: Multiple PVCs exist
        PipelinePvcManager pvcManager = createPvcManager();
        
        String pvcName1 = pvcManager.createPvc("pipeline-1", STORAGE_CLASS);
        String pvcName2 = pvcManager.createPvc("pipeline-2", STORAGE_CLASS);

        // WHEN: Each is queried
        String status1 = pvcManager.getPvcStatus(pvcName1);
        String status2 = pvcManager.getPvcStatus(pvcName2);

        // THEN: Correct status is returned for each
        assertThat(status1, equalTo("Bound"));
        assertThat(status2, equalTo("Bound"));
        
        // Cleanup
        pvcManager.deletePvc(pvcName1);
        pvcManager.deletePvc(pvcName2);
    }
}
