package yeetcd.controller.execution;

import yeetcd.controller.config.Config;
import yeetcd.controller.testinfra.RustFsClient;
import yeetcd.controller.testinfra.TestClusterFixture;
import io.kubernetes.client.openapi.ApiClient;
import lombok.SneakyThrows;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

import java.io.ByteArrayOutputStream;
import java.nio.charset.StandardCharsets;
import java.util.*;
import java.util.concurrent.CompletableFuture;

import static org.hamcrest.CoreMatchers.*;
import static org.hamcrest.MatcherAssert.assertThat;

/**
 * Integration tests for KubernetesExecutionEngine.
 * 
 * Uses TestClusterFixture to manage k3d cluster lifecycle.
 * The cluster is automatically created if missing, and test resources are cleaned up.
 * 
 * Note: Input/output files are stored in S3-backed PVCs, not on local disk.
 * Tests use RustFsClient to upload/download files from S3.
 */
@ExtendWith(TestClusterFixture.class)
public class KubernetesExecutionEngineTest extends AbstractExecutionEngineTest {

    private PipelinePvcManager pvcManager;

    @Override
    String builtImagePullAddress() {
        return TestClusterFixture.getRegistryPullAddress();
    }

    @Override
    ExecutionEngine executionEngine() {
        Config.Kubernetes config = getKubernetesConfig();
        ApiClient apiClient = testApiClient();
        
        pvcManager = new PipelinePvcManager(apiClient);
        
        return new KubernetesExecutionEngine(
            config,
            apiClient,
            true,
            pvcManager
        );
    }

    private static Config.Kubernetes getKubernetesConfig() {
        // Build config dynamically from test infrastructure
        Config.Kubernetes config = new Config.Kubernetes();
        
        Config.Kubernetes.Registry registry = new Config.Kubernetes.Registry();
        registry.setPushAddress(TestClusterFixture.getRegistryPushAddress());
        registry.setPullAddress(TestClusterFixture.getRegistryPullAddress());
        config.setRegistry(registry);
        
        Config.Kubernetes.S3 s3 = new Config.Kubernetes.S3();
        s3.setEndpoint(TestClusterFixture.getRustFsEndpoint());
        s3.setAccessKey(TestClusterFixture.getRustFsAccessKey());
        s3.setSecretKey(TestClusterFixture.getRustFsSecretKey());
        s3.setBucketName("yeetcd-pipelines");
        config.setS3(s3);
        
        return config;
    }

    @SneakyThrows
    private ApiClient testApiClient() {
        return TestClusterFixture.getApiClient();
    }
    
    /**
     * Override: Kubernetes uses PVC-based mounts, not on-disk files.
     * 
     * Flow:
     * 1. Create PVC (which creates an S3 bucket via CSI driver)
     * 2. Get PVC UID - bucket name is "pvc-<uid>"
     * 3. Upload input file to bucket using RustFsClient
     * 4. Run job with PvcMountInput
     * 5. Verify output
     */
    @SneakyThrows
    @Test
    @Override
    public void shouldMakeInputFilesAvailable() {
        // given
        ExecutionEngine executionEngine = executionEngine();
        
        // Create PVC for this test
        String pipelineRunId = UUID.randomUUID().toString();
        String pvcName = pvcManager.createPvc(pipelineRunId, "yeetcd-s3");
        
        // Get PVC UID - the CSI driver creates buckets with name "pvc-<uid>"
        String pvcUid = pvcManager.getPvcUid(pvcName);
        if (pvcUid == null) {
            throw new AssertionError("PVC UID is null - PVC may not be fully provisioned");
        }
        String bucketName = "pvc-" + pvcUid;
        
        // Create RustFsClient and upload input file
        RustFsClient rustFsClient = new RustFsClient(
            TestClusterFixture.getRustFsEndpoint(),
            TestClusterFixture.getRustFsAccessKey(),
            TestClusterFixture.getRustFsSecretKey(),
            bucketName
        );
        rustFsClient.configureAlias();
        
        String fileName = UUID.randomUUID().toString();
        String fileContents = UUID.randomUUID().toString();
        String workId = UUID.randomUUID().toString();
        String inputSubPath = "inputs/" + workId + "/" + fileName;
        
        rustFsClient.uploadFile(inputSubPath, fileContents);
        
        // Create job with PVC mount
        String mountDirectory = "/var/test/";
        String filePath = mountDirectory + fileName;
        
        ByteArrayOutputStream stdOut = new ByteArrayOutputStream();
        JobStreams jobStreams = new JobStreams(stdOut, System.err);
        JobDefinition jobDefinition = new JobDefinition(
            TEST_IMAGE,
            new String[]{"cat", filePath},
            "/",
            Collections.emptyMap(),
            Map.of(mountDirectory, new PvcMountInput(pvcName, "inputs/" + workId)),
            Collections.emptyMap(),
            jobStreams
        );
        
        try {
            // when
            CompletableFuture<JobResult> jobResultFuture = executionEngine.runJob(jobDefinition);
            JobResult jobResult = jobResultFuture.get();
            
            // then
            assertThat(jobResult.exitCode(), equalTo(0));
            assertThat(stdOut.toString(StandardCharsets.UTF_8), equalTo(fileContents));
        } finally {
            // Cleanup PVC
            pvcManager.deletePvc(pvcName);
        }
    }
    
    /**
     * Override: Kubernetes stores output files in S3-backed PVCs.
     * 
     * Flow:
     * 1. Create PVC (which creates an S3 bucket via CSI driver)
     * 2. Get PVC UID - bucket name is "pvc-<uid>"
     * 3. Upload input file to bucket
     * 4. Run job that copies input to output directory
     * 5. Download output file from bucket using RustFsClient
     * 6. Verify contents
     */
    @SneakyThrows
    @Test
    @Override
    public void shouldExtractOutputFiles() {
        // given
        ExecutionEngine executionEngine = executionEngine();
        
        // Create PVC for this test
        String pipelineRunId = UUID.randomUUID().toString();
        String pvcName = pvcManager.createPvc(pipelineRunId, "yeetcd-s3");
        
        // Get PVC UID - the CSI driver creates buckets with name "pvc-<uid>"
        String pvcUid = pvcManager.getPvcUid(pvcName);
        if (pvcUid == null) {
            throw new AssertionError("PVC UID is null - PVC may not be fully provisioned");
        }
        String bucketName = "pvc-" + pvcUid;
        
        // Create RustFsClient and upload input file
        RustFsClient rustFsClient = new RustFsClient(
            TestClusterFixture.getRustFsEndpoint(),
            TestClusterFixture.getRustFsAccessKey(),
            TestClusterFixture.getRustFsSecretKey(),
            bucketName
        );
        rustFsClient.configureAlias();
        
        String fileName = UUID.randomUUID().toString();
        String fileContents = UUID.randomUUID().toString();
        String workId = UUID.randomUUID().toString();
        String inputSubPath = "inputs/" + workId + "/" + fileName;
        
        rustFsClient.uploadFile(inputSubPath, fileContents);
        
        // Create job with PVC mount for input and output
        String mountDirectory = "/var/test/";
        String filePath = mountDirectory + fileName;
        
        String outputName = "output_name";
        String outputDirectory = "/var/out";
        
        ByteArrayOutputStream stdOut = new ByteArrayOutputStream();
        JobStreams jobStreams = new JobStreams(stdOut, System.err);
        JobDefinition jobDefinition = new JobDefinition(
            TEST_IMAGE,
            new String[]{"bash", "-c", String.format("mkdir -p %s && cp %s %s", outputDirectory, filePath, outputDirectory + "/" + fileName)},
            "/",
            Collections.emptyMap(),
            Map.of(mountDirectory, new PvcMountInput(pvcName, "inputs/" + workId)),
            Map.of(outputName, outputDirectory),
            jobStreams
        );
        
        try {
            // when
            CompletableFuture<JobResult> jobResultFuture = executionEngine.runJob(jobDefinition);
            JobResult jobResult = jobResultFuture.get();
            
            // then
            assertThat(jobResult.exitCode(), equalTo(0));
            
            // Download output file from S3 and verify
            // The execution engine generates its own workId (job name) for outputs
            // We need to discover the actual workId by listing the outputs directory
            java.util.List<String> outputFiles = rustFsClient.listFiles("outputs/");
            assertThat("Expected at least one output file", outputFiles.size(), org.hamcrest.Matchers.greaterThan(0));
            
            // Find the file that matches our expected output name
            String outputFile = outputFiles.stream()
                .filter(f -> f.contains("/" + outputName + "/") && f.endsWith("/" + fileName))
                .findFirst()
                .orElseThrow(() -> new AssertionError(
                    "Expected output file not found. Files found: " + outputFiles + 
                    ", expected file matching: */" + outputName + "/" + fileName
                ));
            
            String downloadedContents = rustFsClient.downloadFile(outputFile);
            assertThat(downloadedContents, equalTo(fileContents));
        } finally {
            // Cleanup PVC
            pvcManager.deletePvc(pvcName);
        }
    }
}
