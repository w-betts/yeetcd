package yeetcd.controller.testinfra;

/**
 * Exception thrown by test infrastructure when operations fail.
 * 
 * Defensive coding: Provides full context for debugging - operation, resource, state, and suggested fix.
 * Never fail silently - always throw with actionable information.
 */
public class TestInfrastructureException extends RuntimeException {

    private final String errorCode;
    private final String operation;
    private final String resource;
    private final String state;
    private final String suggestedFix;

    public TestInfrastructureException(String errorCode, String message, String resource, String state, String suggestedFix) {
        super(buildMessage(errorCode, message, resource, state, suggestedFix));
        this.errorCode = errorCode;
        this.operation = message;
        this.resource = resource;
        this.state = state;
        this.suggestedFix = suggestedFix;
    }

    public TestInfrastructureException(String errorCode, String message, String resource, String state, String suggestedFix, Throwable cause) {
        super(buildMessage(errorCode, message, resource, state, suggestedFix), cause);
        this.errorCode = errorCode;
        this.operation = message;
        this.resource = resource;
        this.state = state;
        this.suggestedFix = suggestedFix;
    }

    private static String buildMessage(String errorCode, String message, String resource, String state, String suggestedFix) {
        return String.format(
            "[%s] %s%n  Resource: %s%n  State: %s%n  Suggested fix: %s",
            errorCode, message, resource, state, suggestedFix
        );
    }

    public String getErrorCode() {
        return errorCode;
    }

    public String getOperation() {
        return operation;
    }

    public String getResource() {
        return resource;
    }

    public String getState() {
        return state;
    }

    public String getSuggestedFix() {
        return suggestedFix;
    }
}
