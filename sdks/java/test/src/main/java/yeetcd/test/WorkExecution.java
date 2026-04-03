package yeetcd.test;

/**
 * Record representing a work execution captured from the mock server.
 * Used for expectation verification.
 */
public record WorkExecution(
    String image,
    String[] cmd,
    java.util.Map<String, String> envVars,
    String workingDir,
    int exitCode,
    String stdout,
    String stderr
) {
    /**
     * Creates a WorkExecution with default values.
     */
    public WorkExecution {
        if (envVars == null) envVars = java.util.Collections.emptyMap();
    }

    /**
     * Checks if this execution matches a given behavior.
     */
    public boolean matches(MockBehavior behavior) {
        return behavior.matches(image, cmd, envVars);
    }
}