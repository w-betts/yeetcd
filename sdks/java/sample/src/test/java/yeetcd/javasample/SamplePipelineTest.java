package yeetcd.javasample;

import org.junit.jupiter.api.Test;
import yeetcd.sdk.CustomWorkDefinition;
import yeetcd.test.*;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Map;
import java.util.stream.Stream;

import static org.hamcrest.MatcherAssert.assertThat;
import static org.hamcrest.Matchers.*;
import static org.junit.jupiter.api.Assertions.*;

/**
 * Integration tests for sample pipelines using PipelineTestRun.
 * 
 * These tests verify the PipelineTestRun API for testing pipelines:
 * - Builder configuration with containerisedWork, customWork, dynamicWork
 * - Result querying with hasExecution, hasNoExecution, getExecutionCount
 * - Multiple behaviors and defaults
 */
public class SamplePipelineTest {

    private String buildClasspath() throws IOException {
        StringBuilder cp = new StringBuilder();
        
        // Add sample's own classes
        cp.append("target/classes:");
        
        // Add SDK classes from sibling module
        Path sdkClasses = Path.of("../sdk/target/classes");
        if (Files.exists(sdkClasses)) {
            cp.append(sdkClasses.toAbsolutePath()).append(":");
        }
        
        // Add protocol classes from sibling module
        Path protoClasses = Path.of("../protocol/target/classes");
        if (Files.exists(protoClasses)) {
            cp.append(protoClasses.toAbsolutePath()).append(":");
        }
        
        // Add test classes (for MockServer)
        cp.append("../test/target/classes:");
        
        // Add generated protobuf classes
        Path protoGen = Path.of("../protocol/target/generated-sources/protobuf/java");
        if (Files.exists(protoGen)) {
            cp.append(protoGen.toAbsolutePath()).append(":");
        }
        
        // Add generated gRPC classes
        Path grpcGen = Path.of("../protocol/target/generated-sources/protobuf/grpc-java");
        if (Files.exists(grpcGen)) {
            cp.append(grpcGen.toAbsolutePath()).append(":");
        }
        
        // Add generated annotation classes
        Path annotationGen = Path.of("../sdk/target/generated-sources/annotations");
        if (Files.exists(annotationGen)) {
            cp.append(annotationGen.toAbsolutePath()).append(":");
        }
        
        return cp.toString();
    }

    @Test
    void testPipelineTestRunBuilderCreatesInstance() {
        PipelineTestRun run = PipelineTestRun.builder()
                .pipelineName("testPipeline")
                .containerisedWork("maven:3.9.9-eclipse-temurin-17")
                        .result(0, "output", "")
                .build();
        
        assertThat(run, notNullValue());
    }

    @Test
    void testContainerisedWorkBehaviorBuilderResult() {
        PipelineTestRun run = PipelineTestRun.builder()
                .pipelineName("test")
                .containerisedWork("maven:3.9.9-eclipse-temurin-17")
                        .result(0, "hello", "error")
                .build();
        
        assertThat(run, notNullValue());
    }

    @Test
    void testCustomWorkBehaviorBuilderResult() {
        CustomWorkDefinition instance = new CustomWorkDefinition() {
            @Override
            public void run() {
                System.out.println("Test work");
            }
        };
        
        PipelineTestRun run = PipelineTestRun.builder()
                .pipelineName("test")
                .customWork(instance)
                        .result(0, "output", "")
                .build();
        
        assertThat(run, notNullValue());
    }

    @Test
    void testDefaultBehaviors() {
        PipelineTestRun run = PipelineTestRun.builder()
                .pipelineName("test")
                .defaultContainerisedWork()
                        .result(0, "default", "")
                .build();
        
        PipelineTestRun.builder()
                .pipelineName("test")
                .defaultCustomWork()
                        .result(0, "default", "")
                .build();
        
        assertThat(run, notNullValue());
    }

    @Test
    void testChainedBehaviors() {
        PipelineTestRun.builder()
                .pipelineName("test")
                .containerisedWork("image1")
                        .result(0, "out1", "")
                .containerisedWork("image2")
                        .result(0, "out2", "")
                .build();
    }

    @Test
    void testPipelineTestRunResultHasExecution() {
        WorkExecution exec = WorkExecution.containerised(
                "maven:3.9.9-eclipse-temurin-17",
                new String[]{"bash", "-c", "echo hello"},
                Map.of("VAR", "value"),
                "/workspace",
                0,
                "hello",
                ""
        );
        
        PipelineTestRunResult result = new PipelineTestRunResult(
                PipelineStatus.SUCCESS,
                0,
                java.util.List.of(exec),
                java.util.List.of()
        );
        
        assertTrue(result.hasExecution("maven:3.9.9-eclipse-temurin-17"));
        assertFalse(result.hasExecution("other-image"));
    }

    @Test
    void testPipelineTestRunResultHasNoExecution() {
        WorkExecution exec = WorkExecution.containerised(
                "maven:3.9.9-eclipse-temurin-17",
                new String[]{"bash"},
                null,
                "/workspace",
                0,
                "",
                ""
        );
        
        PipelineTestRunResult result = new PipelineTestRunResult(
                PipelineStatus.SUCCESS,
                0,
                java.util.List.of(exec),
                java.util.List.of()
        );
        
        assertTrue(result.hasNoExecution("other-image"));
    }

    @Test
    void testPipelineTestRunResultGetExecutionCount() {
        WorkExecution exec1 = WorkExecution.containerised("image", new String[]{"cmd"}, null, "/wd", 0, "", "");
        WorkExecution exec2 = WorkExecution.containerised("image", new String[]{"cmd"}, null, "/wd", 0, "", "");
        WorkExecution exec3 = WorkExecution.containerised("image", new String[]{"cmd"}, null, "/wd", 0, "", "");
        
        PipelineTestRunResult result = new PipelineTestRunResult(
                PipelineStatus.SUCCESS,
                0,
                java.util.List.of(exec1, exec2, exec3),
                java.util.List.of()
        );
        
        assertEquals(3, result.getExecutionCount("image"));
        assertEquals(0, result.getExecutionCount("other-image"));
    }

    @Test
    void testPipelineTestRunResultGetExecutions() {
        WorkExecution exec1 = WorkExecution.containerised("image1", new String[]{"cmd1"}, null, "/wd", 0, "", "");
        WorkExecution exec2 = WorkExecution.containerised("image2", new String[]{"cmd2"}, null, "/wd", 0, "", "");
        
        PipelineTestRunResult result = new PipelineTestRunResult(
                PipelineStatus.SUCCESS,
                0,
                java.util.List.of(exec1, exec2),
                java.util.List.of()
        );
        
        assertThat(result.getExecutions(), hasSize(2));
        assertEquals("image1", result.getExecutions().get(0).image());
        assertEquals("image2", result.getExecutions().get(1).image());
    }

    @Test
    void testPipelineTestRunResultFindByImage() {
        WorkExecution exec1 = WorkExecution.containerised("maven", new String[]{"cmd"}, null, "/wd", 0, "", "");
        WorkExecution exec2 = WorkExecution.containerised("other", new String[]{"cmd"}, null, "/wd", 0, "", "");
        WorkExecution exec3 = WorkExecution.containerised("maven", new String[]{"other"}, null, "/wd", 0, "", "");
        
        PipelineTestRunResult result = new PipelineTestRunResult(
                PipelineStatus.SUCCESS,
                0,
                java.util.List.of(exec1, exec2, exec3),
                java.util.List.of()
        );
        
        assertThat(result.findByImage("maven"), hasSize(2));
        assertThat(result.findByImage("other"), hasSize(1));
    }

    @Test
    void testPipelineTestRunResultPipelineStatus() {
        PipelineTestRunResult successResult = new PipelineTestRunResult(
                PipelineStatus.SUCCESS,
                0,
                java.util.List.of(),
                java.util.List.of()
        );
        
        assertEquals(PipelineStatus.SUCCESS, successResult.getPipelineStatus());
        
        PipelineTestRunResult failureResult = new PipelineTestRunResult(
                PipelineStatus.FAILURE,
                1,
                java.util.List.of(),
                java.util.List.of()
        );
        
        assertEquals(PipelineStatus.FAILURE, failureResult.getPipelineStatus());
    }

    @Test
    void testPipelineTestRunResultGetMatchedBehaviors() {
        WorkExecution exec = WorkExecution.containerised("image", new String[]{"cmd"}, null, "/wd", 0, "out", "");
        MatchedBehavior matched = new MatchedBehavior(
                WorkBehaviorType.CONTAINERISED,
                "image",
                exec,
                new WorkResponse(0, "out", "")
        );
        
        PipelineTestRunResult result = new PipelineTestRunResult(
                PipelineStatus.SUCCESS,
                0,
                java.util.List.of(exec),
                java.util.List.of(matched)
        );
        
        assertThat(result.getMatchedBehaviors(), hasSize(1));
        assertEquals(WorkBehaviorType.CONTAINERISED, result.getMatchedBehaviors().get(0).type());
    }

    @Test
    void testCustomWorkExecution() {
        WorkExecution exec = WorkExecution.custom(
                "exec-id-123",
                "built-image",
                new String[]{"java", "yeetcd.sdk.GeneratedCustomWorkRunner", "pipeline", "exec-id-123"},
                Map.of("VAR", "value"),
                "/",
                0,
                "output",
                ""
        );
        
        assertEquals(WorkBehaviorType.CUSTOM, exec.type());
        assertEquals("exec-id-123", exec.matchKey());
    }

    @Test
    void testDynamicWorkExecution() {
        WorkExecution exec = WorkExecution.dynamic(
                "built-image",
                new String[]{"java", "yeetcd.sdk.GeneratedCustomWorkRunner", "pipeline", "dynamic-id"},
                null,
                "/",
                0,
                "generated-work",
                ""
        );
        
        assertEquals(WorkBehaviorType.DYNAMIC, exec.type());
        assertNull(exec.matchKey());
    }

    @Test
    void testWorkResponseBuilder() {
        WorkResponse response = WorkResponse.builder()
                .exitCode(0)
                .stdout("output")
                .stderr("error")
                .build();
        
        assertEquals(0, response.exitCode());
        assertEquals("output", response.stdout());
        assertEquals("error", response.stderr());
    }

    @Test
    void testWorkResponseSuccess() {
        WorkResponse response = WorkResponse.success();
        
        assertEquals(0, response.exitCode());
        assertEquals("", response.stdout());
        assertEquals("", response.stderr());
    }

    @Test
    void testTestPipelinesExist() {
        assertNotNull(TestPipelines.containerisedWorkPipeline());
        assertNotNull(TestPipelines.customWorkPipeline());
        assertNotNull(TestPipelines.compoundWorkPipeline());
        assertNotNull(TestPipelines.dynamicWorkPipeline());
        assertNotNull(TestPipelines.dependentWorkPipeline());
        assertNotNull(TestPipelines.contextWorkPipeline());
        assertNotNull(TestPipelines.multiBehaviorPipeline());
    }

    @Test
    @org.junit.jupiter.api.Disabled("E2E tests require Docker and proper classpath setup")
    void testE2E_ContainerisedWork_HappyPath() throws IOException, InterruptedException {
        String projectDir = System.getProperty("user.dir");
        
        PipelineTestRunResult result = PipelineTestRun.builder()
                .pipelineName("containerisedWorkPipeline")
                .sourcePath(projectDir)
                .containerisedWork("maven:3.9.9-eclipse-temurin-17")
                        .result(0, "containerised work", "")
                .build()
                .start();
        
        System.out.println("CLI Output:\n" + result.getCliOutput());
        assertEquals(PipelineStatus.SUCCESS, result.getPipelineStatus());
        assertTrue(result.hasExecution("maven:3.9.9-eclipse-temurin-17"));
    }

    @Test
    @org.junit.jupiter.api.Disabled("E2E tests require Docker and proper classpath setup")
    void testE2E_ContainerisedWork_FailurePath() throws IOException, InterruptedException {
        String projectDir = System.getProperty("user.dir");
        
        PipelineTestRunResult result = PipelineTestRun.builder()
                .pipelineName("containerisedWorkPipeline")
                .sourcePath(projectDir)
                .containerisedWork("maven:3.9.9-eclipse-temurin-17")
                        .result(1, "", "work failed")
                .build()
                .start();
        
        assertEquals(PipelineStatus.FAILURE, result.getPipelineStatus());
    }

    @Test
    @org.junit.jupiter.api.Disabled("E2E tests require Docker and proper classpath setup")
    void testE2E_CustomWork_MockedResult() throws IOException, InterruptedException {
        String projectDir = System.getProperty("user.dir");
        
        CustomWorkDefinition customWork = TestPipelines.getCustomWorkForPipeline();
        
        PipelineTestRunResult result = PipelineTestRun.builder()
                .pipelineName("customWorkPipeline")
                .sourcePath(projectDir)
                .customWork(customWork)
                        .result(0, "mocked output", "")
                .build()
                .start();
        
        assertEquals(PipelineStatus.SUCCESS, result.getPipelineStatus());
        assertTrue(result.getExecutionCount("custom") > 0);
    }

    @Test
    @org.junit.jupiter.api.Disabled("E2E tests require Docker and proper classpath setup")
    void testE2E_MultipleContainerisedWork() throws IOException, InterruptedException {
        String projectDir = System.getProperty("user.dir");
        
        PipelineTestRunResult result = PipelineTestRun.builder()
                .pipelineName("multiBehaviorPipeline")
                .sourcePath(projectDir)
                .containerisedWork("maven:3.9.9-eclipse-temurin-17")
                        .result(0, "success", "")
                .build()
                .start();
        
        assertEquals(PipelineStatus.SUCCESS, result.getPipelineStatus());
        int execCount = result.getExecutionCount("maven:3.9.9-eclipse-temurin-17");
        assertTrue(execCount >= 2, "Should have at least 2 containerised work executions");
    }

    @Test
    @org.junit.jupiter.api.Disabled("E2E tests require Docker and proper classpath setup")
    void testE2E_DefaultBehavior() throws IOException, InterruptedException {
        String projectDir = System.getProperty("user.dir");
        
        PipelineTestRunResult result = PipelineTestRun.builder()
                .pipelineName("containerisedWorkPipeline")
                .sourcePath(projectDir)
                .defaultContainerisedWork()
                        .result(0, "default response", "")
                .build()
                .start();
        
        assertEquals(PipelineStatus.SUCCESS, result.getPipelineStatus());
    }

    @Test
    @org.junit.jupiter.api.Disabled("E2E tests require Docker and proper classpath setup")
    void testE2E_MultipleStartCalls_Isolation() throws IOException, InterruptedException {
        String projectDir = System.getProperty("user.dir");
        
        PipelineTestRun runner = PipelineTestRun.builder()
                .pipelineName("containerisedWorkPipeline")
                .sourcePath(projectDir)
                .containerisedWork("maven:3.9.9-eclipse-temurin-17")
                        .result(0, "first", "")
                .build();
        
        PipelineTestRunResult result1 = runner.start();
        assertEquals(PipelineStatus.SUCCESS, result1.getPipelineStatus());
        
        PipelineTestRunResult result2 = runner.start();
        assertEquals(PipelineStatus.SUCCESS, result2.getPipelineStatus());
        
        assertTrue(result1.getExecutions().size() > 0);
        assertTrue(result2.getExecutions().size() > 0);
    }
}
