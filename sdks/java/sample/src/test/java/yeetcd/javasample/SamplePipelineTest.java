package yeetcd.javasample;

import org.junit.jupiter.api.Test;
import yeetcd.sdk.CustomWorkDefinition;
import yeetcd.test.*;

import java.util.Map;

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
    void testE2E_ContainerisedWork_HappyPath() throws Exception {
        PipelineTestRunResult result = PipelineTestRun.builder()
                .pipelineName("containerisedWorkPipeline")
                .containerisedWork("maven:3.9.9-eclipse-temurin-17")
                        .result(0, "containerised work", "")
                .build()
                .start();
        
        System.out.println("CLI Output:\n" + result.getCliOutput());
        System.out.println("Executions: " + result.getExecutions());
        assertEquals(PipelineStatus.SUCCESS, result.getPipelineStatus());
        assertTrue(result.hasExecution("maven:3.9.9-eclipse-temurin-17"));
    }

    @Test
    void testE2E_ContainerisedWork_FailurePath() throws Exception {
        PipelineTestRunResult result = PipelineTestRun.builder()
                .pipelineName("containerisedWorkPipeline")
                .containerisedWork("maven:3.9.9-eclipse-temurin-17")
                        .result(1, "", "work failed")
                .build()
                .start();
        
        assertEquals(PipelineStatus.FAILURE, result.getPipelineStatus());
    }

    @Test
    void testE2E_CustomWork_MockedResult() throws Exception {
        CustomWorkDefinition customWork = TestPipelines.getCustomWorkForPipeline();
        
        PipelineTestRunResult result = PipelineTestRun.builder()
                .pipelineName("customWorkPipeline")
                .customWork(customWork)
                        .result(0, "mocked output", "")
                .build()
                .start();
        
        assertEquals(PipelineStatus.SUCCESS, result.getPipelineStatus());
        assertTrue(result.getCustomExecutions().size() > 0);
    }

    @Test
    void testE2E_MultipleContainerisedWork() throws Exception {
        PipelineTestRunResult result = PipelineTestRun.builder()
                .pipelineName("multiBehaviorPipeline")
                .containerisedWork("maven:3.9.9-eclipse-temurin-17")
                        .result(0, "success", "")
                .build()
                .start();
        
        assertEquals(PipelineStatus.SUCCESS, result.getPipelineStatus());
        int execCount = result.getExecutionCount("maven:3.9.9-eclipse-temurin-17");
        assertTrue(execCount >= 2, "Should have at least 2 containerised work executions");
    }

    @Test
    void testE2E_DefaultBehavior() throws Exception {
        PipelineTestRunResult result = PipelineTestRun.builder()
                .pipelineName("containerisedWorkPipeline")
                .defaultContainerisedWork()
                        .result(0, "default response", "")
                .build()
                .start();
        
        assertEquals(PipelineStatus.SUCCESS, result.getPipelineStatus());
    }

    @Test
    void testE2E_MultipleStartCalls_Isolation() throws Exception {
        PipelineTestRun runner = PipelineTestRun.builder()
                .pipelineName("containerisedWorkPipeline")
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
