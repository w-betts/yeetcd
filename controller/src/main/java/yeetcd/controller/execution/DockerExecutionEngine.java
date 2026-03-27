package yeetcd.controller.execution;

import com.github.dockerjava.api.DockerClient;
import com.github.dockerjava.api.async.ResultCallback;
import com.github.dockerjava.api.command.*;
import com.github.dockerjava.api.exception.NotFoundException;
import com.github.dockerjava.api.model.*;
import com.github.dockerjava.core.DefaultDockerClientConfig;
import com.github.dockerjava.core.DockerClientConfig;
import com.github.dockerjava.core.DockerClientImpl;
import com.github.dockerjava.transport.DockerHttpClient;
import com.github.dockerjava.zerodep.ZerodepDockerHttpClient;
import lombok.SneakyThrows;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.compress.archivers.tar.TarArchiveEntry;
import org.apache.commons.compress.archivers.tar.TarArchiveInputStream;

import java.io.*;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.Duration;
import java.util.*;
import java.util.concurrent.CompletableFuture;
import java.util.stream.Collectors;

@Slf4j
public class DockerExecutionEngine extends AbstractExecutionEngine {

    private final DockerClient dockerClient;
    private final DockerDaemonImageBuilder dockerDaemonImageBuilder;

    public DockerExecutionEngine() {
        DockerClientConfig config = DefaultDockerClientConfig.createDefaultConfigBuilder()
                .build();

        DockerHttpClient httpClient = new ZerodepDockerHttpClient.Builder()
                .dockerHost(config.getDockerHost())
                .sslConfig(config.getSSLConfig())
                .maxConnections(100)
                .connectionTimeout(Duration.ofSeconds(30))
                .responseTimeout(Duration.ofSeconds(45))
                .build();

        dockerClient = DockerClientImpl.getInstance(config, httpClient);
        dockerDaemonImageBuilder = new DockerDaemonImageBuilder(dockerClient);
    }

    @Override
    @SneakyThrows
    public CompletableFuture<BuildImageResult> buildImage(BuildImageDefinition buildImageDefinition) {
        return doAsync(() -> dockerDaemonImageBuilder.buildImage(buildImageDefinition));
    }

    @Override
    @SneakyThrows
    public CompletableFuture<Void> removeImage(String image) {
        try {
            dockerClient
                    .removeImageCmd(image)
                    .exec();
            return CompletableFuture.completedFuture(null);
        } catch (Throwable throwable) {
            return CompletableFuture.failedFuture(throwable);
        }
    }

    @Override
    @SneakyThrows
    public CompletableFuture<JobResult> runJob(JobDefinition jobDefinition) {
        CompletableFuture<JobResult> taskResultFuture = new CompletableFuture<>();

        executor.submit(() -> {
            CreateContainerResponse container = null;
            int exitCode = -1;
            File outputDirectoriesParent = null;
            try {
                outputDirectoriesParent = Files.createTempDirectory("yeetcd_dockerengine").toFile();
                pullImage(jobDefinition.image());
                container = createContainer(jobDefinition);
                exitCode = runContainer(container, jobDefinition.jobStreams());
                outputDirectoriesParent.deleteOnExit();
                if (exitCode == 0) {
                    for (Map.Entry<String, String> entry : jobDefinition.outputDirectoryPaths().entrySet()) {
                        extractArchive(outputDirectoriesParent, container, entry.getKey(), entry.getValue());
                    }
                }

            } catch (Throwable throwable) {
                taskResultFuture.completeExceptionally(throwable);
            } finally {
                if (container != null) {
                    try {
                        dockerClient.removeContainerCmd(container.getId()).exec();
                    } catch (NotFoundException exception) {
                        log.debug("already removed");
                    }
                }
            }
            taskResultFuture.complete(new JobResult(exitCode, outputDirectoriesParent));
        });

        return taskResultFuture;
    }

    @SneakyThrows
    private void pullImage(String imageTag) {
        try {
            dockerClient.inspectImageCmd(imageTag).exec();
        } catch (NotFoundException ex) {
            PullImageResultCallback pullImageCallback = new PullImageResultCallback();
            dockerClient.pullImageCmd(imageTag).exec(pullImageCallback);
            pullImageCallback.awaitCompletion();
        }
    }

    @SneakyThrows
    private CreateContainerResponse createContainer(JobDefinition jobDefinition) {

        List<Bind> binds = new LinkedList<>();
        CreateContainerCmd createContainerCmd = dockerClient
                .createContainerCmd(jobDefinition.image())
                .withCmd(jobDefinition.cmd())
                .withWorkingDir(jobDefinition.workingDir());

        for (Map.Entry<String, MountInput> entry : jobDefinition.inputFilePaths().entrySet()) {
            String path = entry.getKey();
            MountInput mountInput = entry.getValue();
            Volume volume = new Volume(path);
            binds.add(new Bind(mountInput.directory().getPath(), volume));
            createContainerCmd = createContainerCmd
                    .withVolumes(volume);
        }
        HostConfig hostConfig = HostConfig
                .newHostConfig()
                .withBinds(binds)
                .withLogConfig(new LogConfig(LogConfig.LoggingType.LOCAL));
        return createContainerCmd
                .withHostConfig(hostConfig)
                .withEnv(jobDefinition.environment() == null ? Collections.emptyList() : jobDefinition.environment()
                        .entrySet().stream()
                        .map(entry -> "%s=%s".formatted(entry.getKey(), entry.getValue()))
                        .collect(Collectors.toList()))
                .exec();
    }

    @SneakyThrows
    private int runContainer(CreateContainerResponse createContainerResponse, JobStreams jobStreams) {
        dockerClient.startContainerCmd(createContainerResponse.getId()).exec();
        WaitContainerResultCallback waitContainerResultCallback = new WaitContainerResultCallback();
        dockerClient.waitContainerCmd(createContainerResponse.getId()).exec(waitContainerResultCallback);
        Integer statusCode = waitContainerResultCallback.awaitStatusCode();
        logContainer(createContainerResponse, jobStreams);
        return statusCode;
    }

    @SneakyThrows
    private void extractArchive(File outputDirectoriesParent, CreateContainerResponse container, String name, String path) {
        InputStream archive = dockerClient.copyArchiveFromContainerCmd(container.getId(), path).exec();
        File destDir = Files.createTempDirectory("yeetcd_dockerengine").toFile();
        destDir.deleteOnExit();
        try (TarArchiveInputStream tarStream = new TarArchiveInputStream(archive)) {
            unTar(tarStream, destDir);
        }
        File outputFile = Path.of(outputDirectoriesParent.getPath(), name).toFile();
        Arrays.stream(Objects.requireNonNullElse(destDir.listFiles(), new File[]{})).forEach(it -> {
            if (!it.renameTo(outputFile)) {
                throw new RuntimeException("failed to move file");
            }
        });
        outputFile.deleteOnExit();
    }

    @SneakyThrows
    private void logContainer(CreateContainerResponse createContainerResponse, JobStreams jobStreams) {
        CompletableFuture<Void> done = new CompletableFuture<>();
        dockerClient.logContainerCmd(createContainerResponse.getId())
                .withStdOut(true)
                .withStdErr(true)
                .withFollowStream(true)
                .exec(new ResultCallback<Frame>() {
                    @Override
                    public void onStart(Closeable closeable) {
                    }

                    @Override
                    public void onNext(Frame object) {
                        processFrame(object, jobStreams);
                    }

                    @Override
                    public void onError(Throwable throwable) {
                        try (PrintWriter printWriter = new PrintWriter(new OutputStreamWriter(jobStreams.getStdErrOutputStream()))) {
                            throwable.printStackTrace(printWriter);
                            printWriter.flush();
                        }
                        done.completeExceptionally(throwable);
                    }

                    @Override
                    @SneakyThrows
                    public void onComplete() {
                        done.complete(null);
                    }

                    @Override
                    @SneakyThrows
                    public void close() {
                        done.complete(null);
                    }
                });
        done.get();
    }

    @SneakyThrows
    private static void processFrame(Frame object, JobStreams jobStreams) {
        switch (object.getStreamType()) {
            case STDOUT -> {
                jobStreams.getStdOutOutputStream().write(object.getPayload());
                jobStreams.getStdOutOutputStream().flush();
            }
            case STDERR -> {
                jobStreams.getStdErrOutputStream().write(object.getPayload());
                jobStreams.getStdErrOutputStream().flush();
            }
        }
    }

    @SuppressWarnings("ResultOfMethodCallIgnored")
    private static void unTar(TarArchiveInputStream tarIn, File dest) throws IOException {
        TarArchiveEntry tarEntry = tarIn.getNextEntry();
        // tarIn is a TarArchiveInputStream
        while (tarEntry != null) {// create a file with the same name as the tarEntry
            File destPath = new File(dest, tarEntry.getName());
            if (tarEntry.isDirectory()) {
                destPath.mkdirs();
            } else {
                destPath.createNewFile();
                byte[] btoRead = new byte[1024];

                BufferedOutputStream bout = new BufferedOutputStream(new FileOutputStream(destPath));
                int len;
                while ((len = tarIn.read(btoRead)) != -1) {
                    bout.write(btoRead, 0, len);
                }

                bout.close();
            }
            tarEntry = tarIn.getNextEntry();
        }
        tarIn.close();
    }

    private record DockerDaemonImageBuilder(DockerClient dockerClient) {

        @SuppressWarnings("ResultOfMethodCallIgnored")
            @SneakyThrows
            public BuildImageResult buildImage(BuildImageDefinition buildImageDefinition) {
                File dockerfile = null;
                try {
                    dockerfile = createDockerfile(buildImageDefinition, buildImageDefinition.artifactDirectory());

                    BuildImageResultCallback resultCallback = new BuildImageResultCallback();
                    dockerClient
                        .buildImageCmd()
                        .withTags(Set.of("%s:%s".formatted(buildImageDefinition.image(), buildImageDefinition.tag())))
                        .withDockerfile(dockerfile)
                        .exec(resultCallback);
                    String imageId = resultCallback.awaitImageId();
                    return new BuildImageResult(imageId);
                }
                finally {
                    if (dockerfile != null) {
                        dockerfile.delete();
                    }
                }
            }

            @SneakyThrows
            private static File createDockerfile(BuildImageDefinition buildImageDefinition, File contextDir) {
                File dockerfile = Files.createFile(Path.of(contextDir.toPath().toString(), "Dockerfile")).toFile();
                dockerfile.deleteOnExit();

                Files.writeString(
                    dockerfile.toPath(),
                    """
                        FROM %s
                        ADD / /artifacts
                        ENTRYPOINT %s
                        CMD %s
                        """.formatted(
                        buildImageDefinition.imageBase().getBaseImage(),
                        "[%s]".formatted(Arrays
                            .stream(buildImageDefinition.imageBase().entryPoint("/artifacts", buildImageDefinition.artifiactNames()))
                            .map("\"%s\""::formatted)
                            .collect(Collectors.joining(", "))
                        ),
                        String.format("[\"%s\"]", buildImageDefinition.cmd())
                    )
                );
                return dockerfile;
            }
        }
}
