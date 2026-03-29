package yeetcd.controller.execution;

import yeetcd.controller.config.Config.Kubernetes;
import com.google.cloud.tools.jib.api.Containerizer;
import com.google.cloud.tools.jib.api.ImageReference;
import com.google.cloud.tools.jib.api.Jib;
import com.google.cloud.tools.jib.api.RegistryImage;
import com.google.cloud.tools.jib.api.buildplan.AbsoluteUnixPath;
import com.google.common.annotations.VisibleForTesting;
import io.kubernetes.client.PodLogs;
import io.kubernetes.client.openapi.ApiClient;
import io.kubernetes.client.openapi.Configuration;
import io.kubernetes.client.openapi.apis.BatchV1Api;
import io.kubernetes.client.openapi.apis.CoreV1Api;
import io.kubernetes.client.openapi.models.*;
import lombok.SneakyThrows;
import lombok.extern.slf4j.Slf4j;


import java.io.*;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.Duration;
import java.util.*;
import java.util.concurrent.CompletableFuture;
import java.util.function.BiFunction;
import java.util.stream.Collectors;

@Slf4j
public class KubernetesExecutionEngine extends AbstractExecutionEngine {
    private static final String namespace = "yeetcd";

    private final JibImageBuilder jibImageBuilder;
    private final BatchV1Api batchV1Api;
    private final PipelinePvcManager pvcManager;
    private static final String yeetcdJobNameLabel = "yeetcdJobName";

//    @SneakyThrows
//    public KubernetesExecutionEngine(Kubernetes config) {
//        this(config, ClientBuilder.cluster().build(), false);
//    }

    @VisibleForTesting
    public KubernetesExecutionEngine(Kubernetes config, ApiClient apiClient, boolean allowInsecureRegistries) {
        this(config, apiClient, allowInsecureRegistries, null);
    }

    @VisibleForTesting
    public KubernetesExecutionEngine(Kubernetes config, ApiClient apiClient, boolean allowInsecureRegistries, 
                                       PipelinePvcManager pvcManager) {
        jibImageBuilder = new JibImageBuilder(config.getRegistry().getPushAddress(), allowInsecureRegistries);
        this.batchV1Api = new BatchV1Api(apiClient);
        this.pvcManager = pvcManager;
        Configuration.setDefaultApiClient(apiClient);
    }

    @Override
    public CompletableFuture<BuildImageResult> buildImage(BuildImageDefinition buildImageDefinition) {
        return doAsync(() -> jibImageBuilder.buildImage(buildImageDefinition));
    }

    @Override
    public CompletableFuture<Void> removeImage(String image) {
        return doAsync(() -> {
            log.info("Removing image '{}' from registry", image);
            // For k3d's registry, we use the registry API to delete the image
            // The registry API endpoint is at /v2/<name>/manifests/<reference>
            String registryUrl = "http://" + jibImageBuilder.registry + "/v2/" + image + "/manifests/latest";
            
            try {
                java.net.URL url = new java.net.URL(registryUrl);
                java.net.HttpURLConnection conn = (java.net.HttpURLConnection) url.openConnection();
                conn.setRequestMethod("DELETE");
                conn.setConnectTimeout(5000);
                conn.setReadTimeout(5000);
                
                int responseCode = conn.getResponseCode();
                if (responseCode == 202 || responseCode == 404) {
                    // 202 = Accepted (deleted), 404 = Not found (already deleted)
                    log.info("Image '{}' removed successfully (response code: {})", image, responseCode);
                } else {
                    // Log but don't fail - image may not exist or registry may not support deletion
                    log.warn("Unexpected response when removing image '{}': {}", image, responseCode);
                }
                conn.disconnect();
            } catch (IOException e) {
                // Log but don't fail - this is best-effort cleanup
                log.warn("Failed to remove image '{}': {}", image, e.getMessage());
            }
            return null;
        });
    }

    @SneakyThrows
    @Override
    public CompletableFuture<JobResult> runJob(JobDefinition jobDefinition) {
        String name = UUID.randomUUID().toString();
        return doAsync(() -> runJobSync(jobDefinition, name))
            .thenCompose(job -> {
                CompletableFuture<JobResult> jobResult = checkResult(name);
                return CompletableFuture
                    .allOf(
                        jobResult,
                        logPod(name, jobDefinition.jobStreams().getStdOutOutputStream())
                    )
                    .thenCompose(nothing -> jobResult);
            })
            .whenComplete((jobResult, throwable) -> {
                if (throwable != null || jobResult.exitCode() > 0) {
                    streamPodListLogsSync(listJobPodsSync(name), jobDefinition.jobStreams().getStdErrOutputStream());
                    log.error("Error in job. Logs sent to std err stream", throwable);
                }
            });

    }

    @SneakyThrows
    private V1Job runJobSync(JobDefinition jobDefinition, String name) {
        String workId = name;
        String pvcName = null;
        
        // Check if we have PVC-based mounts and extract PVC name
        boolean hasPvcMounts = jobDefinition.inputFilePaths().values().stream()
            .anyMatch(MountInput::isPvcMount);
        
        if (hasPvcMounts) {
            // Find the PVC name from the first PVC mount
            pvcName = jobDefinition.inputFilePaths().values().stream()
                .filter(MountInput::isPvcMount)
                .map(m -> ((PvcMountInput) m).pvcName())
                .findFirst()
                .orElse(null);
        }
        
        // Build pod spec with PVC volumes if needed
        V1PodSpec podSpec = buildPodSpec(jobDefinition, name, pvcName, workId);
        
        V1Job job = batchV1Api
            .createNamespacedJob(namespace, new V1Job()
                .metadata(new V1ObjectMeta().name(name))
                .spec(new V1JobSpec()
                    .backoffLimit(1)
                    .template(new V1PodTemplateSpec()
                        .metadata(new V1ObjectMeta().labels(Map.of(yeetcdJobNameLabel, name)))
                        .spec(podSpec)
                    )
                ))
            .execute();
        
        return job;
    }
    
    /**
     * Builds the pod spec with PVC volumes if needed.
     */
    private V1PodSpec buildPodSpec(JobDefinition jobDefinition, String name, String pvcName, String workId) {
        List<V1Volume> volumes = new ArrayList<>();
        List<V1VolumeMount> volumeMounts = new ArrayList<>();
        
        // Add PVC volume if we have PVC-based mounts
        if (pvcName != null) {
            volumes.add(new V1Volume()
                .name("pipeline-pvc")
                .persistentVolumeClaim(new V1PersistentVolumeClaimVolumeSource()
                    .claimName(pvcName)
                )
            );
            
            // Add volume mounts for input files
            for (Map.Entry<String, MountInput> entry : jobDefinition.inputFilePaths().entrySet()) {
                String mountPath = entry.getKey();
                MountInput mountInput = entry.getValue();
                if (mountInput.isPvcMount()) {
                    PvcMountInput pvcMount = (PvcMountInput) mountInput;
                    volumeMounts.add(new V1VolumeMount()
                        .name("pipeline-pvc")
                        .mountPath(mountPath)
                        .subPath(pvcMount.subPath().replaceFirst("^/", "")) // Remove leading slash
                    );
                }
            }
            
            // Add volume mounts for output directories
            for (Map.Entry<String, String> entry : jobDefinition.outputDirectoryPaths().entrySet()) {
                String outputName = entry.getKey();
                String mountPath = entry.getValue();
                String outputSubPath = "outputs/" + workId + "/" + outputName;
                
                volumeMounts.add(new V1VolumeMount()
                    .name("pipeline-pvc")
                    .mountPath(mountPath)
                    .subPath(outputSubPath)
                );
            }
        }
        
        V1Container container = new V1Container()
            .name(name)
            .image(jobDefinition.image())
            .command(Arrays.stream(jobDefinition.cmd()).toList())
            .env(jobDefinition.environment().entrySet().stream()
                .map(entry -> new V1EnvVar().name(entry.getKey()).value(entry.getValue()))
                .collect(Collectors.toList()))
            .workingDir(jobDefinition.workingDir());
        
        if (!volumeMounts.isEmpty()) {
            container.volumeMounts(volumeMounts);
        }
        
        V1PodSpec podSpec = new V1PodSpec()
            .containers(List.of(container))
            .restartPolicy("Never");
        
        if (!volumes.isEmpty()) {
            podSpec.volumes(volumes);
        }
        
        return podSpec;
    }

    private CompletableFuture<JobResult> checkResult(String jobName) {
        return this
            .doAsyncUntil(
                () -> getJob(jobName),
                v1Job -> v1Job.getStatus() != null && (v1Job.getStatus().getActive() == null || v1Job.getStatus().getActive() == 0),
                Duration.ofSeconds(30)
            )
            .thenApply(job -> {
                int exitCode = Objects.requireNonNull(job.getStatus()).getFailed() == null ? 0 : 1;
                return new JobResult(exitCode, null);
            });
    }

    private CompletableFuture<Void> logPod(String jobName, OutputStream outputStream) {
        return this
            .doAsyncUntil(
                () -> listJobPodsSync(jobName),
                list -> list.getItems().size() > 0 &&
                        list.getItems().stream().allMatch(pod ->
                            pod.getStatus() != null &&
                            pod.getStatus().getContainerStatuses() != null &&
                            pod.getStatus().getContainerStatuses().stream().anyMatch(containerStatus ->
                                containerStatus.getState() != null &&
                                (containerStatus.getState().getRunning() != null || containerStatus.getState().getTerminated() != null)
                            )
                        ),
                Duration.ofSeconds(30)
            )
            .thenCompose(list -> doAsync(
                () -> {
                    streamPodListLogsSync(list, outputStream);
                    return null;
                }
            ));
    }

    @SneakyThrows
    private static void streamPodListLogsSync(V1PodList list, OutputStream outputStream) {
        list.getItems().forEach(pod -> streamPodLogs(pod, outputStream));
    }

    @SneakyThrows
    private static void streamPodLogs(V1Pod pod, OutputStream outputStream) {
        try (InputStream inputStream = new PodLogs().streamNamespacedPodLog(pod)) {
            inputStream.transferTo(outputStream);
        }
    }

    @SneakyThrows
    private static V1PodList listJobPodsSync(String jobName) {
        return new CoreV1Api().listNamespacedPod(namespace).labelSelector("%s=%s".formatted(yeetcdJobNameLabel, jobName)).execute();
    }

    @SneakyThrows
    private V1Job getJob(String jobName) {
        return batchV1Api.readNamespacedJob(jobName, namespace).execute();
    }

    private static class JibImageBuilder {

        private final BiFunction<String, String, Containerizer> containerizer;
        final String registry;

        public JibImageBuilder(String registry, boolean allowInsecureRegistries) {
            this.registry = registry;
            this.containerizer = (image, tag) -> Containerizer.to(RegistryImage.named(ImageReference.of(this.registry, image, tag))).setAllowInsecureRegistries(allowInsecureRegistries);
        }

        @SneakyThrows
        public BuildImageResult buildImage(BuildImageDefinition buildImageDefinition) {
            File[] artifacts = buildImageDefinition.artifactDirectory().listFiles();
            List<Path> artifactPaths = artifacts == null ? Collections.emptyList() : Arrays.stream(artifacts).map(File::toPath).collect(Collectors.toList());
            Jib
                .from(buildImageDefinition.imageBase().getBaseImage())
                .addLayer(artifactPaths, AbsoluteUnixPath.get("/artifacts"))
                .setEntrypoint(buildImageDefinition.imageBase().entryPoint("/artifacts", buildImageDefinition.artifiactNames()))
                .setProgramArguments(buildImageDefinition.cmd())
                .containerize(containerizer.apply(buildImageDefinition.image(), buildImageDefinition.tag()));
            return new BuildImageResult("%s:%s".formatted(buildImageDefinition.image(), buildImageDefinition.tag()));
        }
    }
}
