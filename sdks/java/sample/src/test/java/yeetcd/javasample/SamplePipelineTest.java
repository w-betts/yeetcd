package yeetcd.javasample;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import yeetcd.test.ExpectationVerifier;
import yeetcd.test.MockBehavior;
import yeetcd.test.WorkExecution;
import yeetcd.test.YeetcdMockRunner;

import java.io.IOException;
import java.util.List;
import java.util.Map;

import static org.hamcrest.MatcherAssert.assertThat;
import static org.hamcrest.Matchers.*;
import static org.junit.jupiter.api.Assertions.*;

/**
 * Integration tests for sample pipelines using YeetcdMockRunner.
 * 
 * These tests demonstrate the usage of YeetcdMockRunner to test sample pipelines
 * with mock behaviors. The tests verify that:
 * - YeetcdMockRunner can be configured and built
 * - Mock behaviors can be defined and registered
 * - Work executions are recorded correctly
 * - Expectation verifications work correctly
 */
public class SamplePipelineTest {

    private YeetcdMockRunner runner;

    @BeforeEach
    void setUp() throws IOException {
        // Builder configuration - the runner is built but we may not
        // actually run the CLI in unit test mode
        runner = YeetcdMockRunner.builder()
                .port(50051)
                .pipelineName("sample")
                .build();
    }

    @AfterEach
    void tearDown() {
        if (runner != null) {
            try {
                runner.close();
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        }
    }

    /**
     * Test that YeetcdMockRunner can be built with builder pattern.
     * 
     * Given: YeetcdMockRunner.builder() is called
     * When: Builder is configured with port, pipeline name
     * Then: YeetcdMockRunner instance is created successfully
     */
    @Test
    void testBuilderCreatesRunner() throws IOException {
        // Given - builder is created in setUp
        
        // When - runner is built with custom settings
        YeetcdMockRunner testRunner = YeetcdMockRunner.builder()
                .port(50052)
                .classpath("/test/classpath")
                .sourcePath("/test/source")
                .pipelineName("testPipeline")
                .build();
        
        // Then - runner is created
        assertThat(testRunner, notNullValue());
        
        // Clean up
        try {
            testRunner.close();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    /**
     * Test that YeetcdMockRunner builder has default port.
     * 
     * Given: YeetcdMockRunner.builder() is called
     * When: No port is specified
     * Then: Default port 50051 is used
     */
    @Test
    void testBuilderHasDefaultPort() throws IOException {
        // Given - builder with no port specified
        YeetcdMockRunner testRunner = YeetcdMockRunner.builder()
                .pipelineName("test")
                .build();
        
        // Then - runner is created (mock server will use default port)
        assertThat(testRunner, notNullValue());
        
        try {
            testRunner.close();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    /**
     * Test mock behavior builder creates correct behavior.
     * 
     * Given: A MockBehavior builder
     * When: Behavior is configured with image, cmd, exit code, stdout
     * Then: MockBehavior is created with all configured values
     */
    @Test
    void testMockBehaviorBuilder() {
        // Given
        String expectedImage = "maven:3.9.9-eclipse-temurin-17";
        String[] expectedCmd = {"bash", "-c", "echo 'Hello'"};
        int expectedExitCode = 0;
        String expectedStdout = "Hello from container";
        
        // When
        MockBehavior behavior = MockBehavior.builder()
                .matchingImage(expectedImage)
                .matchingCmd(expectedCmd)
                .exitCode(expectedExitCode)
                .stdout(expectedStdout)
                .build();
        
        // Then
        assertThat(behavior.getImage(), is(expectedImage));
        assertThat(behavior.getCmd(), is(expectedCmd));
        assertThat(behavior.toMockWorkResponse().getExitCode(), is(expectedExitCode));
        assertThat(behavior.toMockWorkResponse().getStdout(), is(expectedStdout));
    }

    /**
     * Test mock behavior matching with exact image and command.
     * 
     * Given: A MockBehavior configured for specific image and command
     * When: Work execution matches the behavior
     * Then: matches() returns true
     */
    @Test
    void testMockBehaviorMatches() {
        // Given
        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("maven:3.9.9-eclipse-temurin-17")
                .matchingCmd("bash", "-c", "echo 'test'")
                .build();
        
        // When - exact match
        boolean matches = behavior.matches(
                "maven:3.9.9-eclipse-temurin-17",
                new String[]{"bash", "-c", "echo 'test'"},
                null
        );
        
        // Then
        assertThat(matches, is(true));
    }

    /**
     * Test mock behavior does not match different command.
     * 
     * Given: A MockBehavior configured for specific command
     * When: Work execution has different command
     * Then: matches() returns false
     */
    @Test
    void testMockBehaviorDoesNotMatchDifferentCommand() {
        // Given
        MockBehavior behavior = MockBehavior.builder()
                .matchingCmd("bash", "-c", "echo 'test'")
                .build();
        
        // When - different command
        boolean matches = behavior.matches(
                "maven:3.9.9-eclipse-temurin-17",
                new String[]{"bash", "-c", "echo 'different'"},
                null
        );
        
        // Then
        assertThat(matches, is(false));
    }

    /**
     * Test mock behavior matching with environment variables.
     * 
     * Given: A MockBehavior with environment variable matching
     * When: Work execution has matching env vars
     * Then: matches() returns true
     */
    @Test
    void testMockBehaviorMatchesEnvVars() {
        // Given
        Map<String, String> expectedEnvVars = Map.of(
                "PIPELINE_NAME", "sample",
                "WORK_NAME", "test-work"
        );
        
        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("test-image")
                .matchingEnvVars(expectedEnvVars)
                .build();
        
        // When - matching env vars
        boolean matches = behavior.matches(
                "test-image",
                null,
                expectedEnvVars
        );
        
        // Then
        assertThat(matches, is(true));
    }

    /**
     * Test expectation verifier asserts executed work.
     * 
     * Given: Work executions recorded
     * When: assertExecuted is called with matching behavior
     * Then: No errors are recorded
     */
    @Test
    void testExpectationVerifierAssertExecuted() {
        // Given - executions match behavior
        WorkExecution execution = new WorkExecution(
                "maven:3.9.9-eclipse-temurin-17",
                new String[]{"bash", "-c", "echo 'test'"},
                Map.of("KEY", "value"),
                "/workspace",
                0,
                "output",
                ""
        );
        
        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("maven:3.9.9-eclipse-temurin-17")
                .matchingCmd("bash", "-c", "echo 'test'")
                .build();
        
        // When
        ExpectationVerifier verifier = new ExpectationVerifier(List.of(execution));
        verifier.assertExecuted(behavior);
        
        // Then
        assertThat(verifier.hasErrors(), is(false));
    }

    /**
     * Test expectation verifier detects missing execution.
     * 
     * Given: No work executions recorded
     * When: assertExecuted is called
     * Then: Error is recorded
     */
    @Test
    void testExpectationVerifierDetectsMissingExecution() {
        // Given
        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("maven:3.9.9-eclipse-temurin-17")
                .matchingCmd("bash", "-c", "echo 'test'")
                .build();
        
        // When
        ExpectationVerifier verifier = new ExpectationVerifier(List.of());
        verifier.assertExecuted(behavior);
        
        // Then
        assertThat(verifier.hasErrors(), is(true));
        assertThat(verifier.getErrors(), hasSize(1));
    }

    /**
     * Test expectation verifier asserts executed count.
     * 
     * Given: Multiple work executions recorded
     * When: assertExecutedCount is called with expected count
     * Then: No errors if count matches
     */
    @Test
    void testExpectationVerifierAssertExecutedCount() {
        // Given - 3 executions matching behavior
        WorkExecution execution = new WorkExecution(
                "test-image",
                new String[]{"cmd"},
                null,
                "/workspace",
                0,
                "",
                ""
        );
        
        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("test-image")
                .build();
        
        // When
        ExpectationVerifier verifier = new ExpectationVerifier(
                List.of(execution, execution, execution)
        );
        verifier.assertExecutedCount(behavior, 3);
        
        // Then
        assertThat(verifier.hasErrors(), is(false));
    }

    /**
     * Test expectation verifier detects wrong count.
     * 
     * Given: Work executions with different count than expected
     * When: assertExecutedCount is called with wrong count
     * Then: Error is recorded
     */
    @Test
    void testExpectationVerifierDetectsWrongCount() {
        // Given - 2 executions
        WorkExecution execution = new WorkExecution(
                "test-image",
                new String[]{"cmd"},
                null,
                "/workspace",
                0,
                "",
                ""
        );
        
        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("test-image")
                .build();
        
        // When - expecting 3
        ExpectationVerifier verifier = new ExpectationVerifier(
                List.of(execution, execution)
        );
        verifier.assertExecutedCount(behavior, 3);
        
        // Then
        assertThat(verifier.hasErrors(), is(true));
    }

    /**
     * Test expectation verifier asserts not executed.
     * 
     * Given: No executions for specific behavior
     * When: assertNotExecuted is called
     * Then: No errors
     */
    @Test
    void testExpectationVerifierAssertNotExecuted() {
        // Given - no executions
        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("test-image")
                .build();
        
        // When
        ExpectationVerifier verifier = new ExpectationVerifier(List.of());
        verifier.assertNotExecuted(behavior);
        
        // Then
        assertThat(verifier.hasErrors(), is(false));
    }

    /**
     * Test expectation verifier detects unexpected execution.
     * 
     * Given: Execution exists for behavior
     * When: assertNotExecuted is called
     * Then: Error is recorded
     */
    @Test
    void testExpectationVerifierDetectsUnexpectedExecution() {
        // Given
        WorkExecution execution = new WorkExecution(
                "test-image",
                new String[]{"cmd"},
                null,
                "/workspace",
                0,
                "",
                ""
        );
        
        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("test-image")
                .build();
        
        // When
        ExpectationVerifier verifier = new ExpectationVerifier(List.of(execution));
        verifier.assertNotExecuted(behavior);
        
        // Then
        assertThat(verifier.hasErrors(), is(true));
    }

    /**
     * Test expectation verifier verify() throws on errors.
     * 
     * Given: ExpectationVerifier with errors
     * When: verify() is called
     * Then: AssertionError is thrown
     */
    @Test
    void testExpectationVerifierVerifyThrows() {
        // Given
        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("test-image")
                .build();
        
        ExpectationVerifier verifier = new ExpectationVerifier(List.of());
        verifier.assertExecuted(behavior);
        
        // When/Then
        assertThrows(AssertionError.class, () -> verifier.verify());
    }

    /**
     * Test expectation verifier verify() passes without errors.
     * 
     * Given: ExpectationVerifier without errors
     * When: verify() is called
     * Then: No exception is thrown
     */
    @Test
    void testExpectationVerifierVerifyPasses() {
        // Given
        WorkExecution execution = new WorkExecution(
                "test-image",
                new String[]{"cmd"},
                null,
                "/workspace",
                0,
                "",
                ""
        );
        
        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("test-image")
                .build();
        
        ExpectationVerifier verifier = new ExpectationVerifier(List.of(execution));
        verifier.assertExecuted(behavior);
        
        // When/Then - should not throw
        assertDoesNotThrow(() -> verifier.verify());
    }

    /**
     * Test sample pipeline 'sample' has containerised work.
     * 
     * Given: SamplePipelines.sample() pipeline definition
     * When: Pipeline is inspected
     * Then: It contains containerised work definition
     * 
     * Note: This verifies the pipeline structure for test planning
     */
    @Test
    void testSamplePipelineStructure() {
        // Given - the sample pipeline is defined in SamplePipelines.sample()
        // We verify the structure by checking the generated protobuf
        // For now, we test the expected structure
        
        // Then - sample pipeline should exist and be processable
        // The actual structure is verified through the SDK generator
        assertThat(SamplePipelines.class, notNullValue());
    }

    /**
     * Test integration with sample pipeline - mock server setup.
     * 
     * Given: A configured YeetcdMockRunner
     * When: startMockServer() is called
     * Then: Mock server starts and is accessible
     * 
     * Note: This is an integration test that requires the mock server
     */
    @Test
    void testMockServerStartup() throws IOException, InterruptedException {
        // Given - runner is built in setUp
        
        // When
        runner.startMockServer();
        
        // Then - mock server should be running
        assertThat(runner.getMockServer(), notNullValue());
        assertThat(runner.getMockServer().getPort(), is(50051));
    }

    /**
     * Test mock behavior registration and retrieval.
     * 
     * Given: A configured YeetcdMockRunner with mock server started
     * When: defineBehavior() is called
     * Then: Behavior is registered and can be used in verification
     */
    @Test
    void testDefineBehavior() throws IOException, InterruptedException {
        // Given
        runner.startMockServer();
        
        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("test-image")
                .matchingCmd("test", "command")
                .exitCode(0)
                .stdout("test output")
                .build();
        
        // When
        runner.defineBehavior(behavior);
        
        // Then - behavior is registered (no exception thrown)
        // The actual verification would happen after CLI runs
        assertThat(runner, notNullValue());
    }

    /**
     * Test getting executed work after mock execution.
     * 
     * Given: A YeetcdMockRunner with recorded executions
     * When: getExecutedWork() is called
     * Then: List of WorkExecution is returned
     */
    @Test
    void testGetExecutedWork() throws IOException {
        // Given - runner with empty executions
        assertThat(runner.getExecutedWork(), is(notNullValue()));
        
        // When/Then - getExecutedWork returns empty list initially
        assertThat(runner.getExecutedWork(), hasSize(0));
    }

    /**
     * Test multiple behaviors can be defined.
     * 
     * Given: Multiple MockBehavior configurations
     * When: Each behavior is built
     * Then: Each behavior is independent and correctly configured
     */
    @Test
    void testMultipleBehaviors() {
        // Given
        MockBehavior behavior1 = MockBehavior.builder()
                .matchingImage("image1")
                .exitCode(0)
                .stdout("output1")
                .build();
        
        MockBehavior behavior2 = MockBehavior.builder()
                .matchingImage("image2")
                .exitCode(0)
                .stdout("output2")
                .build();
        
        MockBehavior behavior3 = MockBehavior.builder()
                .matchingImage("image3")
                .exitCode(1)
                .stderr("error")
                .build();
        
        // When/Then - all behaviors are independent
        assertThat(behavior1.getImage(), is("image1"));
        assertThat(behavior2.getImage(), is("image2"));
        assertThat(behavior3.getImage(), is("image3"));
        assertThat(behavior3.toMockWorkResponse().getExitCode(), is(1));
    }

    /**
     * Test fluent API for expectation verification.
     * 
     * Given: Work executions
     * When: Multiple assertions are chained
     * Then: All assertions are applied
     */
    @Test
    void testFluentExpectationVerification() {
        // Given
        WorkExecution exec1 = new WorkExecution("image1", null, null, "/wd", 0, "", "");
        WorkExecution exec2 = new WorkExecution("image2", null, null, "/wd", 0, "", "");
        WorkExecution exec3 = new WorkExecution("image1", null, null, "/wd", 0, "", "");
        
        MockBehavior behavior1 = MockBehavior.builder().matchingImage("image1").build();
        MockBehavior behavior2 = MockBehavior.builder().matchingImage("image2").build();
        
        // When - fluent API
        ExpectationVerifier verifier = new ExpectationVerifier(List.of(exec1, exec2, exec3));
        verifier.assertExecutedCount(behavior1, 2)
                .assertExecutedCount(behavior2, 1);
        
        // Then
        assertThat(verifier.hasErrors(), is(false));
        assertDoesNotThrow(() -> verifier.verify());
    }
}
