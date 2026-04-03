package yeetcd.test;

import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

/**
 * Unit tests for MockBehavior DSL.
 */
class MockBehaviorTest {

    @Test
    void testBuilderCreatesBehaviorWithDefaults() {
        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("alpine")
                .matchingCmd("echo", "hello")
                .build();

        assertEquals("alpine", behavior.getImage());
        assertArrayEquals(new String[]{"echo", "hello"}, behavior.getCmd());
    }

    @Test
    void testBuilderSetsExitCodeAndStdout() {
        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("nginx")
                .exitCode(0)
                .stdout("Hello, World!")
                .stderr("")
                .build();

        assertEquals(0, behavior.toMockWorkResponse().getExitCode());
        assertEquals("Hello, World!", behavior.toMockWorkResponse().getStdout());
        assertEquals("", behavior.toMockWorkResponse().getStderr());
    }

    @Test
    void testMatchesExactImageAndCmd() {
        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("nginx")
                .matchingCmd("echo", "hello")
                .build();

        assertTrue(behavior.matches("nginx", new String[]{"echo", "hello"}, null));
        assertFalse(behavior.matches("alpine", new String[]{"echo", "hello"}, null));
    }

    @Test
    void testMatchesDifferentCmd() {
        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("nginx")
                .build();

        assertTrue(behavior.matches("nginx", new String[]{"echo", "hello"}, null));
        assertTrue(behavior.matches("nginx", new String[]{"sh"}, null));
    }

    @Test
    void testMatchesWithEnvVars() {
        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("nginx")
                .matchingEnvVars(java.util.Map.of("PORT", "8080"))
                .build();

        assertTrue(behavior.matches("nginx", null, java.util.Map.of("PORT", "8080")));
        assertFalse(behavior.matches("nginx", null, java.util.Map.of("PORT", "3000")));
    }

    @Test
    void testMatchesNullImageMatchesAny() {
        MockBehavior behavior = MockBehavior.builder()
                .matchingCmd("echo", "hello")
                .build();

        assertTrue(behavior.matches("any-image", new String[]{"echo", "hello"}, null));
    }

    @Test
    void testMatchesNullCmdMatchesAny() {
        MockBehavior behavior = MockBehavior.builder()
                .matchingImage("nginx")
                .build();

        assertTrue(behavior.matches("nginx", new String[]{"any", "cmd"}, null));
    }
}