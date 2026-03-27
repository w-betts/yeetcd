package yeetcd.controller.execution;

import lombok.SneakyThrows;
import org.junit.jupiter.api.Test;

import javax.tools.*;
import java.io.ByteArrayOutputStream;
import java.io.File;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.*;
import java.util.concurrent.CompletableFuture;

import static javax.tools.StandardLocation.CLASS_OUTPUT;
import static org.apache.commons.lang3.StringUtils.isBlank;
import static org.hamcrest.CoreMatchers.*;
import static org.hamcrest.MatcherAssert.assertThat;

public abstract class AbstractExecutionEngineTest {

    public static final String TEST_IMAGE = "maven:3.9.9-eclipse-temurin-17";

    abstract ExecutionEngine executionEngine();

    abstract String builtImagePullAddress();

    @SneakyThrows
    @Test
    public void shouldRunAJob() {
        // given
        ExecutionEngine executionEngine = executionEngine();

        ByteArrayOutputStream stdOut = new ByteArrayOutputStream();
        String output = "some output";
        JobDefinition jobDefinition = new JobDefinition(
            // TODO some plain image
            TEST_IMAGE,
            new String[]{"echo", "-n", output},
            "/",
            Collections.emptyMap(),
            Collections.emptyMap(),
            Collections.emptyMap(),
            new JobStreams(stdOut, System.err)
        );

        // when
        CompletableFuture<JobResult> jobResultFuture = executionEngine.runJob(jobDefinition);
        JobResult jobResult = jobResultFuture.get();

        // then
        assertThat(jobResult.exitCode(), equalTo(0));
        assertThat(stdOut.toString(StandardCharsets.UTF_8), equalTo(output));
    }

    @SneakyThrows
    @Test
    public void shouldMakeEnvironmentVariablesAvailable() {
        // given
        ExecutionEngine executionEngine = executionEngine();

        String envVarName = "ENV_VAR";
        String envVarValue = UUID.randomUUID().toString();

        ByteArrayOutputStream stdOut = new ByteArrayOutputStream();
        JobStreams jobStreams = new JobStreams(stdOut, System.err);
        JobDefinition jobDefinition = new JobDefinition(
            // TODO some plain image
            TEST_IMAGE,
            new String[]{"env"},
            "/",
            Map.of(envVarName, envVarValue),
            Collections.emptyMap(),
            Collections.emptyMap(),
            jobStreams
        );

        // when
        CompletableFuture<JobResult> jobResultFuture = executionEngine.runJob(jobDefinition);
        JobResult jobResult = jobResultFuture.get();

        // then
        assertThat(jobResult.exitCode(), equalTo(0));
        assertThat(stdOut.toString(StandardCharsets.UTF_8), containsString("%s=%s".formatted(envVarName, envVarValue)));
    }

    @SneakyThrows
    @Test
    public void shouldMakeInputFilesAvailable() {
        // given
        ExecutionEngine executionEngine = executionEngine();

        String fileName = UUID.randomUUID().toString();
        File hostDirectory = Files.createTempDirectory("yeetcd_test").toFile();
        File file = Files.createFile(Path.of(hostDirectory.toPath().toString(), fileName)).toFile();
        String fileContents = UUID.randomUUID().toString();
        Files.writeString(file.toPath(), fileContents);

        String mountDirectory = "/var/test/";
        String filePath = mountDirectory + fileName;

        ByteArrayOutputStream stdOut = new ByteArrayOutputStream();
        JobStreams jobStreams = new JobStreams(stdOut, System.err);
        JobDefinition jobDefinition = new JobDefinition(
            // TODO some plain image
            TEST_IMAGE,
            new String[]{"cat", filePath},
            "/",
            Collections.emptyMap(),
            Map.of(mountDirectory, new OnDiskMountInput(hostDirectory)),
            Collections.emptyMap(),
            jobStreams
        );

        // when
        CompletableFuture<JobResult> jobResultFuture = executionEngine.runJob(jobDefinition);
        JobResult jobResult = jobResultFuture.get();

        // then
        assertThat(jobResult.exitCode(), equalTo(0));
        assertThat(stdOut.toString(StandardCharsets.UTF_8), equalTo(fileContents));
    }

    @SneakyThrows
    @Test
    public void shouldExtractOutputFiles() {
        // given
        ExecutionEngine executionEngine = executionEngine();

        String fileName = UUID.randomUUID().toString();
        File hostDirectory = Files.createTempDirectory("yeetcd_test").toFile();
        File file = Files.createFile(Path.of(hostDirectory.toPath().toString(), fileName)).toFile();
        String fileContents = UUID.randomUUID().toString();
        Files.writeString(file.toPath(), fileContents);

        String mountDirectory = "/var/test/";
        String filePath = mountDirectory + fileName;

        String outputName = "output_name";
        String outputDirectory = "/var/out";

        ByteArrayOutputStream stdOut = new ByteArrayOutputStream();
        JobStreams jobStreams = new JobStreams(stdOut, System.err);
        JobDefinition jobDefinition = new JobDefinition(
            // TODO some plain image
            TEST_IMAGE,
            new String[]{"bash", "-c", String.format("mkdir -p %s && cp %s %s", outputDirectory, filePath, outputDirectory + "/" + fileName)},
            "/",
            Collections.emptyMap(),
            Map.of(mountDirectory, new OnDiskMountInput(hostDirectory)),
            Map.of(outputName, outputDirectory),
            jobStreams
        );

        // when
        CompletableFuture<JobResult> jobResultFuture = executionEngine.runJob(jobDefinition);
        JobResult jobResult = jobResultFuture.get();

        // then
        assertThat(jobResult.exitCode(), equalTo(0));
        File outputDirectoriesParent = jobResult.outputDirectoriesParent();
        File[] outputFiles = outputDirectoriesParent.listFiles();
        assertThat(outputFiles, notNullValue());
        assertThat(outputFiles.length, equalTo(1));
        File[] outputDirectoryContents = outputFiles[0].listFiles();
        assertThat(outputDirectoryContents, notNullValue());
        assertThat(outputDirectoryContents.length, equalTo(1));
        assertThat(Files.readString(outputDirectoryContents[0].toPath()), equalTo(fileContents));
    }

    @SneakyThrows
    @Test
    public void shouldBuildAndRunAJavaImage() {
        // given
        ExecutionEngine executionEngine = executionEngine();

        String output = UUID.randomUUID().toString();
        String mainClassName = "TestMain";
        String image = this.getClass().getSimpleName().toLowerCase();
        String tag = UUID.randomUUID().toString();
        try {
            File artifactParentDirectory = Files.createTempDirectory("yeetcd_test").toFile();
            artifactParentDirectory.deleteOnExit();
            String artifactDefinitionName = "classes";
            givenCompiledSource(artifactParentDirectory, artifactDefinitionName, Map.of(
                mainClassName + ".java", """
                    public class %s {
                        public static void main(String[] args) {
                            System.out.print("%s");
                        }
                    }
                    """.formatted(mainClassName, output)
            ));

            // when
            BuildImageResult buildImageResult = executionEngine.buildImage(new BuildImageDefinition(
                image,
                tag,
                ImageBase.JAVA,
                artifactParentDirectory,
                List.of(artifactDefinitionName),
                mainClassName
            )).get();

            // then
            String imageId = isBlank(builtImagePullAddress()) ? buildImageResult.imageId() : "%s/%s".formatted(builtImagePullAddress(), buildImageResult.imageId());

            // and
            ByteArrayOutputStream stdOut = new ByteArrayOutputStream();
            CompletableFuture<JobResult> jobResultFuture = executionEngine.runJob(new JobDefinition(
                imageId,
                new String[]{},
                "/",
                Collections.emptyMap(),
                Collections.emptyMap(),
                Collections.emptyMap(),
                new JobStreams(stdOut, System.err)
            ));
            JobResult jobResult = jobResultFuture.get();
            assertThat(jobResult.exitCode(), equalTo(0));
            assertThat(stdOut.toString(StandardCharsets.UTF_8), equalTo(output));

        }
        finally {
            executionEngine.removeImage(tag);
        }
    }

    @SuppressWarnings({"UnusedReturnValue", "ResultOfMethodCallIgnored"})
    @SneakyThrows
    private static File givenCompiledSource(File artifactParentDirectory, String artifactDefinitionName, Map<String, String> sourceFileContents) {
        Path sourceDir = Files.createTempDirectory("yeetcd_test");
        sourceDir.toFile().deleteOnExit();

        File classDir = Path.of(artifactParentDirectory.getPath(), artifactDefinitionName).toFile();
        classDir.mkdir();
        classDir.deleteOnExit();

        List<JavaFileObject> compilationUnits = new LinkedList<>();
        for (Map.Entry<String, String> entry : sourceFileContents.entrySet()) {
            String fileName = entry.getKey();
            String contents = entry.getValue();
            Path sourceFile = Files.createFile(Path.of(sourceDir.toString(), fileName));
            Files.writeString(sourceFile, contents);
            compilationUnits.add(new SimpleJavaFileObject(sourceFile.toUri(), JavaFileObject.Kind.SOURCE) {
                @Override
                public CharSequence getCharContent(boolean ignoreEncodingErrors) {
                    return contents;
                }
            });
        }

        JavaCompiler javaCompiler = ToolProvider.getSystemJavaCompiler();
        StandardJavaFileManager standardFileManager = javaCompiler.getStandardFileManager(null, null, null);
        standardFileManager.setLocation(CLASS_OUTPUT, List.of(classDir));


        JavaCompiler.CompilationTask task = javaCompiler.getTask(
            null,
            standardFileManager,
            null,
            null,
            null,
            compilationUnits
        );

        Boolean compilationSuccessful = task.call();

        assertThat(compilationSuccessful, equalTo(true));
        return classDir;
    }
}
