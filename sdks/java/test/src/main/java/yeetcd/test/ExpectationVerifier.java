package yeetcd.test;

import java.util.List;
import java.util.ArrayList;

/**
 * Verifies work executions match expected behaviors.
 * Provides methods: assertExecuted(), assertExecutedCount(), assertNotExecuted(), getExecutedWork().
 */
public class ExpectationVerifier {

    private final List<WorkExecution> executions;
    private final List<String> errors = new ArrayList<>();

    public ExpectationVerifier(List<WorkExecution> executions) {
        this.executions = executions;
    }

    /**
     * Asserts that at least one work execution matched the given behavior.
     */
    public ExpectationVerifier assertExecuted(MockBehavior behavior) {
        long count = executions.stream()
                .filter(e -> e.matches(behavior))
                .count();
        
        if (count == 0) {
            String cmdStr = behavior.getCmd() != null ? String.join(" ", behavior.getCmd()) : "any";
            errors.add("Expected at least one execution matching behavior (image=" + behavior.getImage() 
                    + ", cmd=" + cmdStr + ") but found none");
        }
        
        return this;
    }

    /**
     * Asserts that exactly N work executions matched the given behavior.
     */
    public ExpectationVerifier assertExecutedCount(MockBehavior behavior, int expectedCount) {
        long count = executions.stream()
                .filter(e -> e.matches(behavior))
                .count();
        
        if (count != expectedCount) {
            String cmdStr = behavior.getCmd() != null ? String.join(" ", behavior.getCmd()) : "any";
            errors.add("Expected exactly " + expectedCount + " executions matching behavior (image=" + behavior.getImage() 
                    + ", cmd=" + cmdStr + ") but found " + count);
        }
        
        return this;
    }

    /**
     * Asserts that no work execution matched the given behavior.
     */
    public ExpectationVerifier assertNotExecuted(MockBehavior behavior) {
        long count = executions.stream()
                .filter(e -> e.matches(behavior))
                .count();
        
        if (count > 0) {
            String cmdStr = behavior.getCmd() != null ? String.join(" ", behavior.getCmd()) : "any";
            errors.add("Expected no executions matching behavior (image=" + behavior.getImage() 
                    + ", cmd=" + cmdStr + ") but found " + count);
        }
        
        return this;
    }

    /**
     * Gets all work executions.
     */
    public List<WorkExecution> getExecutedWork() {
        return new ArrayList<>(executions);
    }

    /**
     * Gets work executions that match the given behavior.
     */
    public List<WorkExecution> getExecutedWork(MockBehavior behavior) {
        return executions.stream()
                .filter(e -> e.matches(behavior))
                .toList();
    }

    /**
     * Verifies all assertions and throws if any failed.
     */
    public void verify() {
        if (!errors.isEmpty()) {
            throw new AssertionError(String.join("\n", errors));
        }
    }

    /**
     * Checks if there are any errors (for fluent API).
     */
    public boolean hasErrors() {
        return !errors.isEmpty();
    }

    /**
     * Gets the error messages.
     */
    public List<String> getErrors() {
        return new ArrayList<>(errors);
    }
}