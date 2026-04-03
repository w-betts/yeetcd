package yeetcd.test;

import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

/**
 * Unit tests for ExpectationVerifier.
 */
class ExpectationVerifierTest {

    @Test
    void testAssertExecutedPassesWhenMatched() {
        WorkExecution exec = new WorkExecution(
            "nginx",
            new String[]{"echo", "hello"},
            java.util.Map.of("PORT", "8080"),
            "/app",
            0,
            "Hello!",
            ""
        );

        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("nginx")
                .matchingCmd("echo", "hello")
                .build();

        ExpectationVerifier verifier = new ExpectationVerifier(java.util.List.of(exec));
        
        // Should not throw
        verifier.assertExecuted(behavior).verify();
    }

    @Test
    void testAssertExecutedFailsWhenNotMatched() {
        WorkExecution exec = new WorkExecution(
            "alpine",
            new String[]{"echo", "hello"},
            null,
            "/app",
            0,
            "Hello!",
            ""
        );

        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("nginx")
                .build();

        ExpectationVerifier verifier = new ExpectationVerifier(java.util.List.of(exec));
        
        assertThrows(AssertionError.class, () -> 
            verifier.assertExecuted(behavior).verify()
        );
    }

    @Test
    void testAssertExecutedCountPassesWhenExactCount() {
        WorkExecution exec1 = new WorkExecution("nginx", new String[]{"echo", "hello"}, null, "/app", 0, "", "");
        WorkExecution exec2 = new WorkExecution("nginx", new String[]{"echo", "hello"}, null, "/app", 0, "", "");

        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("nginx")
                .matchingCmd("echo", "hello")
                .build();

        ExpectationVerifier verifier = new ExpectationVerifier(java.util.List.of(exec1, exec2));
        
        // Should not throw
        verifier.assertExecutedCount(behavior, 2).verify();
    }

    @Test
    void testAssertExecutedCountFailsWhenCountMismatch() {
        WorkExecution exec = new WorkExecution("nginx", new String[]{"echo", "hello"}, null, "/app", 0, "", "");

        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("nginx")
                .build();

        ExpectationVerifier verifier = new ExpectationVerifier(java.util.List.of(exec));
        
        assertThrows(AssertionError.class, () -> 
            verifier.assertExecutedCount(behavior, 5).verify()
        );
    }

    @Test
    void testAssertNotExecutedPassesWhenNoMatch() {
        WorkExecution exec = new WorkExecution("alpine", new String[]{"echo", "hello"}, null, "/app", 0, "", "");

        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("nginx")
                .build();

        ExpectationVerifier verifier = new ExpectationVerifier(java.util.List.of(exec));
        
        // Should not throw
        verifier.assertNotExecuted(behavior).verify();
    }

    @Test
    void testAssertNotExecutedFailsWhenMatchFound() {
        WorkExecution exec = new WorkExecution("nginx", new String[]{"echo", "hello"}, null, "/app", 0, "", "");

        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("nginx")
                .build();

        ExpectationVerifier verifier = new ExpectationVerifier(java.util.List.of(exec));
        
        assertThrows(AssertionError.class, () -> 
            verifier.assertNotExecuted(behavior).verify()
        );
    }

    @Test
    void testGetExecutedWorkReturnsAll() {
        WorkExecution exec1 = new WorkExecution("nginx", null, null, "/app", 0, "", "");
        WorkExecution exec2 = new WorkExecution("alpine", null, null, "/app", 0, "", "");

        ExpectationVerifier verifier = new ExpectationVerifier(java.util.List.of(exec1, exec2));
        
        assertEquals(2, verifier.getExecutedWork().size());
    }

    @Test
    void testGetExecutedWorkFiltersByBehavior() {
        WorkExecution exec1 = new WorkExecution("nginx", null, null, "/app", 0, "", "");
        WorkExecution exec2 = new WorkExecution("alpine", null, null, "/app", 0, "", "");

        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("nginx")
                .build();

        ExpectationVerifier verifier = new ExpectationVerifier(java.util.List.of(exec1, exec2));
        
        assertEquals(1, verifier.getExecutedWork(behavior).size());
        assertEquals("nginx", verifier.getExecutedWork(behavior).get(0).image());
    }

    @Test
    void testFluentApiChaining() {
        WorkExecution exec1 = new WorkExecution("nginx", new String[]{"echo", "hi"}, null, "/app", 0, "", "");
        WorkExecution exec2 = new WorkExecution("alpine", new String[]{"echo", "bye"}, null, "/app", 0, "", "");

        MockBehavior nginx = MockBehavior.builder().matchingImage("nginx").build();
        MockBehavior alpine = MockBehavior.builder().matchingImage("alpine").build();

        ExpectationVerifier verifier = new ExpectationVerifier(java.util.List.of(exec1, exec2));
        
        // Should not throw - chaining multiple assertions
        verifier
            .assertExecuted(nginx)
            .assertExecuted(alpine)
            .verify();
    }
}