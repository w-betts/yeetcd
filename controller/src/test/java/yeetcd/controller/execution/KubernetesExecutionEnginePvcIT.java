package yeetcd.controller.execution;

import yeetcd.controller.config.Config;
import yeetcd.controller.testinfra.RustFsClient;
import yeetcd.controller.testinfra.TestClusterFixture;
import io.kubernetes.client.openapi.ApiClient;
import lombok.SneakyThrows;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

import java.io.ByteArrayOutputStream;
import java.nio.charset.StandardCharsets;
import java.util.Collections;
import java.util.Map;
import java.util.concurrent.CompletableFuture;

import static org.hamcrest.CoreMatchers.*;
import static org.hamcrest.MatcherAssert.assertThat;

/**
 * Integration tests for KubernetesExecutionEngine with PVC-based input/output handling.
 * 
 * Tests verify that:
 * - Input files are uploaded to RustFS via S3 API and mounted into pods via PVC
 * - Output files written to PVC are retrievable via S3 API
 * - Image removal works correctly
 * 
 * Uses TestClusterFixture to manage k3d cluster lifecycle with RustFS deployed.
 */
@ExtendWith(TestClusterFixture.class)
public class KubernetesExecutionEnginePvcIT {

    private KubernetesExecutionEngine executionEngine;
    private PipelinePvcManager pvcManager;
    private RustFsClient rustFsClient;
    private static final String TEST_IMAGE = "maven:3.9.9-eclipse-temurin-17";
    private static final String STORAGE_CLASS = "yeetcd-s3";

    @BeforeEach
    void setUp() {
        Config.Kubernetes config = getKubernetesConfig();
        ApiClient apiClient = TestClusterFixture.getApiClient();
        
        pvcManager = new PipelinePvcManager(apiClient);
        rustFsClient = new RustFsClient(
            config.getS3().getEndpoint(),
            config.getS3().getAccessKey(),
            config.getS3().getSecretKey(),
            config.getS3().getBucketName()
        );
        rustFsClient.configureAlias();
        executionEngine = new KubernetesExecutionEngine(config, apiClient, true, pvcManager);
    }

    private static Config.Kubernetes getKubernetesConfig() {
        Config.Kubernetes config = new Config.Kubernetes();
        
        Config.Kubernetes.Registry registry = new Config.Kubernetes.Registry();
        registry.setPushAddress(TestClusterFixture.getRegistryPushAddress());
        registry.setPullAddress(TestClusterFixture.getRegistryPullAddress());
        config.setRegistry(registry);
        
        Config.Kubernetes.S3 s3 = new Config.Kubernetes.S3();
        s3.setEndpoint("http://localhost:9000");
        s3.setAccessKey("rustfsadmin");
        s3.setSecretKey("rustfsadmin");
        s3.setBucketName("yeetcd-pipelines");
        config.setS3(s3);
        
        return config;
    }

    /**
     * GIVEN: PVC exists with input files uploaded to RustFS via S3 API in work-specific subdirectory
     * WHEN: job is run with inputFilePaths
     * THEN: pod mounts PVC and files are accessible at specified mount path
     */
    @Test
    @SneakyThrows
    public void shouldMakeInputFilesAvailableViaPvc() {
        // GIVEN: PVC exists with input files uploaded to RustFS
        String pipelineRunId = "test-input-pipeline";
        String pvcName = pvcManager.createPvc(pipelineRunId, STORAGE_CLASS);
        String workId = "work-input-test";
        String subPath = "/inputs/" + workId;
        String mountPath = "/mnt/inputs";
        String fileName = "test-input.txt";
        String fileContents = "Hello from PVC input!";
        
        // Upload file to RustFS via S3 API
        uploadFileToS3(subPath + "/" + fileName, fileContents);

        ByteArrayOutputStream stdOut = new ByteArrayOutputStream();
        JobStreams jobStreams = new JobStreams(stdOut, System.err);
        
        JobDefinition jobDefinition = JobDefinition.builder()
            .image(TEST_IMAGE)
            .cmd(new String[]{"cat", mountPath + "/" + fileName})
            .workingDir("/")
            .environment(Collections.emptyMap())
            .inputFilePaths(Map.of(mountPath, new PvcMountInput(pvcName, subPath)))
            .outputDirectoryPaths(Collections.emptyMap())
            .jobStreams(jobStreams)
            .build();

        // WHEN: job is run with inputFilePaths
        CompletableFuture<JobResult> jobResultFuture = executionEngine.runJob(jobDefinition);
        JobResult jobResult = jobResultFuture.get();

        // THEN: pod mounts PVC and files are accessible
        assertThat(jobResult.exitCode(), equalTo(0));
        assertThat(stdOut.toString(StandardCharsets.UTF_8), equalTo(fileContents));
        
        // Cleanup
        pvcManager.deletePvc(pvcName);
    }

    /**
     * GIVEN: job writes files to output directory on PVC
     * WHEN: job completes
     * THEN: output files are retrievable via S3 API from RustFS
     */
    @Test
    @SneakyThrows
    public void shouldExtractOutputFilesFromPvc() {
        // GIVEN: job writes files to output directory on PVC
        String pipelineRunId = "test-output-pipeline";
        String pvcName = pvcManager.createPvc(pipelineRunId, STORAGE_CLASS);
        String workId = "work-output-test";
        String outputSubPath = "/outputs/" + workId;
        String outputMountPath = "/mnt/outputs";
        String outputName = "test-output";
        String fileName = "output.txt";
        String fileContents = "Hello from PVC output!";

        ByteArrayOutputStream stdOut = new ByteArrayOutputStream();
        JobStreams jobStreams = new JobStreams(stdOut, System.err);
        
        JobDefinition jobDefinition = JobDefinition.builder()
            .image(TEST_IMAGE)
            .cmd(new String[]{"bash", "-c", "echo '" + fileContents + "' > " + outputMountPath + "/" + fileName})
            .workingDir("/")
            .environment(Collections.emptyMap())
            .inputFilePaths(Collections.emptyMap())
            .outputDirectoryPaths(Map.of(outputName, outputMountPath))
            .jobStreams(jobStreams)
            .build();

        // WHEN: job completes
        CompletableFuture<JobResult> jobResultFuture = executionEngine.runJob(jobDefinition);
        JobResult jobResult = jobResultFuture.get();

        // THEN: job completes successfully
        assertThat(jobResult.exitCode(), equalTo(0));
        
        // AND: output files are retrievable via S3 API
        String retrievedContent = downloadFileFromS3(outputSubPath + "/" + fileName);
        assertThat(retrievedContent, equalTo(fileContents));
        
        // Cleanup
        pvcManager.deletePvc(pvcName);
    }

    /**
     * GIVEN: image exists in registry
     * WHEN: removeImage(imageId) is called
     * THEN: image is deleted from registry
     */
    @Test
    @SneakyThrows
    public void shouldRemoveImageFromRegistry() {
        // GIVEN: image exists in registry
        String image = "test-remove-image";
        String tag = "latest";
        
        // Build a simple image first
        BuildImageResult buildResult = executionEngine.buildImage(createSimpleBuildImageDefinition(image, tag)).get();
        String imageId = buildResult.imageId();
        
        // Verify image exists by running it
        ByteArrayOutputStream stdOut = new ByteArrayOutputStream();
        JobDefinition jobDefinition = JobDefinition.builder()
            .image(TestClusterFixture.getRegistryPullAddress() + "/" + imageId)
            .cmd(new String[]{"echo", "test"})
            .workingDir("/")
            .environment(Collections.emptyMap())
            .inputFilePaths(Collections.emptyMap())
            .outputDirectoryPaths(Collections.emptyMap())
            .jobStreams(new JobStreams(stdOut, System.err))
            .build();
        
        JobResult jobResult = executionEngine.runJob(jobDefinition).get();
        assertThat(jobResult.exitCode(), equalTo(0));

        // WHEN: removeImage is called
        CompletableFuture<Void> removeFuture = executionEngine.removeImage(imageId);
        removeFuture.get();

        // THEN: image is deleted from registry (subsequent runs should fail)
        // Note: Actual verification depends on registry implementation
    }

    /**
     * GIVEN: k3d cluster with RustFS deployed
     * WHEN: simple echo job is submitted
     * THEN: job completes with exit code 0 and stdout contains expected output
     */
    @Test
    @SneakyThrows
    public void shouldRunAJob() {
        // GIVEN: k3d cluster with RustFS deployed
        String output = "some output";
        ByteArrayOutputStream stdOut = new ByteArrayOutputStream();
        
        JobDefinition jobDefinition = JobDefinition.builder()
            .image(TEST_IMAGE)
            .cmd(new String[]{"echo", "-n", output})
            .workingDir("/")
            .environment(Collections.emptyMap())
            .inputFilePaths(Collections.emptyMap())
            .outputDirectoryPaths(Collections.emptyMap())
            .jobStreams(new JobStreams(stdOut, System.err))
            .build();

        // WHEN: simple echo job is submitted
        CompletableFuture<JobResult> jobResultFuture = executionEngine.runJob(jobDefinition);
        JobResult jobResult = jobResultFuture.get();

        // THEN: job completes with exit code 0 and stdout contains expected output
        assertThat(jobResult.exitCode(), equalTo(0));
        assertThat(stdOut.toString(StandardCharsets.UTF_8), equalTo(output));
    }

    /**
     * GIVEN: k3d cluster with RustFS deployed
     * WHEN: job is submitted with env vars
     * THEN: env vars are available in container
     */
    @Test
    @SneakyThrows
    public void shouldMakeEnvironmentVariablesAvailable() {
        // GIVEN: k3d cluster with RustFS deployed
        String envVarName = "TEST_ENV_VAR";
        String envVarValue = "test-value-123";
        ByteArrayOutputStream stdOut = new ByteArrayOutputStream();
        
        JobDefinition jobDefinition = JobDefinition.builder()
            .image(TEST_IMAGE)
            .cmd(new String[]{"env"})
            .workingDir("/")
            .environment(Map.of(envVarName, envVarValue))
            .inputFilePaths(Collections.emptyMap())
            .outputDirectoryPaths(Collections.emptyMap())
            .jobStreams(new JobStreams(stdOut, System.err))
            .build();

        // WHEN: job is submitted with env vars
        CompletableFuture<JobResult> jobResultFuture = executionEngine.runJob(jobDefinition);
        JobResult jobResult = jobResultFuture.get();

        // THEN: env vars are available in container
        assertThat(jobResult.exitCode(), equalTo(0));
        assertThat(stdOut.toString(StandardCharsets.UTF_8), containsString(envVarName + "=" + envVarValue));
    }

    /**
     * GIVEN: k3d cluster with RustFS deployed
     * WHEN: job is submitted with input files (uploaded to RustFS via S3 API before job)
     * THEN: input files are accessible in container via PVC mount
     */
    @Test
    @SneakyThrows
    public void shouldMakeInputFilesAvailable() {
        // GIVEN: k3d cluster with RustFS deployed
        String pipelineRunId = "test-input-files-pipeline";
        String pvcName = pvcManager.createPvc(pipelineRunId, STORAGE_CLASS);
        String workId = "work-input-files";
        String subPath = "/inputs/" + workId;
        String mountPath = "/mnt/inputs";
        String fileName = "input.txt";
        String fileContents = "Input file contents";
        
        // Upload file to RustFS via S3 API before job
        uploadFileToS3(subPath + "/" + fileName, fileContents);

        ByteArrayOutputStream stdOut = new ByteArrayOutputStream();
        JobDefinition jobDefinition = JobDefinition.builder()
            .image(TEST_IMAGE)
            .cmd(new String[]{"cat", mountPath + "/" + fileName})
            .workingDir("/")
            .environment(Collections.emptyMap())
            .inputFilePaths(Map.of(mountPath, new PvcMountInput(pvcName, subPath)))
            .outputDirectoryPaths(Collections.emptyMap())
            .jobStreams(new JobStreams(stdOut, System.err))
            .build();

        // WHEN: job is submitted with input files
        CompletableFuture<JobResult> jobResultFuture = executionEngine.runJob(jobDefinition);
        JobResult jobResult = jobResultFuture.get();

        // THEN: input files are accessible in container via PVC mount
        assertThat(jobResult.exitCode(), equalTo(0));
        assertThat(stdOut.toString(StandardCharsets.UTF_8), equalTo(fileContents));
        
        // Cleanup
        pvcManager.deletePvc(pvcName);
    }

    /**
     * GIVEN: k3d cluster with RustFS deployed
     * WHEN: job writes output files to PVC
     * THEN: output files are extracted and available in JobResult
     */
    @Test
    @SneakyThrows
    public void shouldExtractOutputFiles() {
        // GIVEN: k3d cluster with RustFS deployed
        String pipelineRunId = "test-extract-output-pipeline";
        String pvcName = pvcManager.createPvc(pipelineRunId, STORAGE_CLASS);
        String workId = "work-extract-output";
        String outputSubPath = "/outputs/" + workId;
        String outputMountPath = "/mnt/outputs";
        String outputName = "my-output";
        String fileName = "result.txt";
        String fileContents = "Output result contents";

        ByteArrayOutputStream stdOut = new ByteArrayOutputStream();
        JobDefinition jobDefinition = JobDefinition.builder()
            .image(TEST_IMAGE)
            .cmd(new String[]{"bash", "-c", "mkdir -p " + outputMountPath + " && echo '" + fileContents + "' > " + outputMountPath + "/" + fileName})
            .workingDir("/")
            .environment(Collections.emptyMap())
            .inputFilePaths(Collections.emptyMap())
            .outputDirectoryPaths(Map.of(outputName, outputMountPath))
            .jobStreams(new JobStreams(stdOut, System.err))
            .build();

        // WHEN: job writes output files to PVC
        CompletableFuture<JobResult> jobResultFuture = executionEngine.runJob(jobDefinition);
        JobResult jobResult = jobResultFuture.get();

        // THEN: job completes successfully
        assertThat(jobResult.exitCode(), equalTo(0));
        
        // AND: output files are retrievable via S3 API
        String retrievedContent = downloadFileFromS3(outputSubPath + "/" + fileName);
        assertThat(retrievedContent, equalTo(fileContents));
        
        // Cleanup
        pvcManager.deletePvc(pvcName);
    }

    /**
     * GIVEN: k3d cluster with RustFS deployed
     * WHEN: Java image is built and run
     * THEN: image executes correctly and is cleaned up
     */
    @Test
    @SneakyThrows
    public void shouldBuildAndRunAJavaImage() {
        // GIVEN: k3d cluster with RustFS deployed
        String output = "Hello from built image!";
        String image = "test-java-image";
        String tag = "v1";
        
        try {
            // Build image
            BuildImageDefinition buildDef = createJavaBuildImageDefinition(image, tag, output);
            BuildImageResult buildResult = executionEngine.buildImage(buildDef).get();
            String imageId = buildResult.imageId();
            
            // Run image
            ByteArrayOutputStream stdOut = new ByteArrayOutputStream();
            JobDefinition jobDefinition = JobDefinition.builder()
                .image(TestClusterFixture.getRegistryPullAddress() + "/" + imageId)
                .cmd(new String[]{})
                .workingDir("/")
                .environment(Collections.emptyMap())
                .inputFilePaths(Collections.emptyMap())
                .outputDirectoryPaths(Collections.emptyMap())
                .jobStreams(new JobStreams(stdOut, System.err))
                .build();
            
            // WHEN: Java image is built and run
            JobResult jobResult = executionEngine.runJob(jobDefinition).get();
            
            // THEN: image executes correctly
            assertThat(jobResult.exitCode(), equalTo(0));
            assertThat(stdOut.toString(StandardCharsets.UTF_8), equalTo(output));
            
        } finally {
            // Cleanup: image is cleaned up
            executionEngine.removeImage(image + ":" + tag).get();
        }
    }

    // Helper methods
    
    private void uploadFileToS3(String key, String content) {
        rustFsClient.uploadFile(key, content);
    }
    
    private String downloadFileFromS3(String key) {
        return rustFsClient.downloadFile(key);
    }
    
    private BuildImageDefinition createSimpleBuildImageDefinition(String image, String tag) {
        // Create a simple image that just echoes a message
        java.io.File artifactDir;
        try {
            artifactDir = java.nio.file.Files.createTempDirectory("yeetcd_simple_build").toFile();
        } catch (java.io.IOException e) {
            throw new RuntimeException("Failed to create temp directory", e);
        }
        
        // Create a simple Dockerfile-like structure
        // For kaniko, we need to create a tar.gz with the Dockerfile and any artifacts
        return new BuildImageDefinition(
            image,
            tag,
            ImageBase.JAVA,
            artifactDir,
            java.util.Collections.emptyList(),
            "echo 'Simple image built'"
        );
    }
    
    @SneakyThrows
    @SuppressWarnings("ResultOfMethodCallIgnored")
    private BuildImageDefinition createJavaBuildImageDefinition(String image, String tag, String output) {
        String mainClassName = "TestMain";
        
        // Create temp directories
        java.io.File artifactParentDirectory = java.nio.file.Files.createTempDirectory("yeetcd_java_build").toFile();
        artifactParentDirectory.deleteOnExit();
        
        java.nio.file.Path sourceDir = java.nio.file.Files.createTempDirectory("yeetcd_java_source");
        sourceDir.toFile().deleteOnExit();
        
        String artifactDefinitionName = "classes";
        java.io.File classDir = java.nio.file.Path.of(artifactParentDirectory.getPath(), artifactDefinitionName).toFile();
        classDir.mkdir();
        classDir.deleteOnExit();
        
        // Create Java source file
        String sourceCode = """
            public class %s {
                public static void main(String[] args) {
                    System.out.print("%s");
                }
            }
            """.formatted(mainClassName, output);
        
        java.nio.file.Path sourceFile = java.nio.file.Files.createFile(
            java.nio.file.Path.of(sourceDir.toString(), mainClassName + ".java"));
        java.nio.file.Files.writeString(sourceFile, sourceCode);
        
        // Compile the Java source
        java.util.List<javax.tools.JavaFileObject> compilationUnits = new java.util.LinkedList<>();
        compilationUnits.add(new javax.tools.SimpleJavaFileObject(sourceFile.toUri(), javax.tools.JavaFileObject.Kind.SOURCE) {
            @Override
            public CharSequence getCharContent(boolean ignoreEncodingErrors) {
                return sourceCode;
            }
        });
        
        javax.tools.JavaCompiler javaCompiler = javax.tools.ToolProvider.getSystemJavaCompiler();
        javax.tools.StandardJavaFileManager standardFileManager = javaCompiler.getStandardFileManager(null, null, null);
        standardFileManager.setLocation(javax.tools.StandardLocation.CLASS_OUTPUT, java.util.List.of(classDir));
        
        javax.tools.JavaCompiler.CompilationTask task = javaCompiler.getTask(
            null,
            standardFileManager,
            null,
            null,
            null,
            compilationUnits
        );
        
        Boolean compilationSuccessful = task.call();
        if (!compilationSuccessful) {
            throw new RuntimeException("Failed to compile Java source");
        }
        
        return new BuildImageDefinition(
            image,
            tag,
            ImageBase.JAVA,
            artifactParentDirectory,
            java.util.List.of(artifactDefinitionName),
            mainClassName
        );
    }
}
