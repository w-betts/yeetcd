package yeetcd.test;

import yeetcd.protocol.mock.Mock.MockImageBuildResponse;
import yeetcd.protocol.mock.Mock.MockWorkResponse;

import java.util.Map;

/**
 * DSL for defining mock behaviors in tests.
 * Allows matching work by image, command, and environment variables.
 */
public class MockBehavior {

    private String image;
    private String[] cmd;
    private Map<String, String> envVars;
    private int exitCode;
    private String stdout;
    private String stderr;

    private MockBehavior(Builder builder) {
        this.image = builder.image;
        this.cmd = builder.cmd;
        this.envVars = builder.envVars;
        this.exitCode = builder.exitCode;
        this.stdout = builder.stdout;
        this.stderr = builder.stderr;
    }

    public String getImage() {
        return image;
    }

    public String[] getCmd() {
        return cmd;
    }

    public Map<String, String> getEnvVars() {
        return envVars;
    }

    /**
     * Converts this behavior to a MockWorkResponse for use with MockServer.
     */
    public MockWorkResponse toMockWorkResponse() {
        return MockWorkResponse.newBuilder()
                .setExitCode(exitCode)
                .setStdout(stdout)
                .setStderr(stderr)
                .build();
    }

    /**
     * Checks if this behavior matches the given work parameters.
     */
    public boolean matches(String image, String[] cmd, Map<String, String> envVars) {
        // Check image
        if (this.image != null && !this.image.equals(image)) {
            return false;
        }
        
        // Check command
        if (this.cmd != null && this.cmd.length > 0) {
            if (cmd == null) {
                return false;
            }
            if (this.cmd.length != cmd.length) {
                return false;
            }
            for (int i = 0; i < this.cmd.length; i++) {
                if (!this.cmd[i].equals(cmd[i])) {
                    return false;
                }
            }
        }
        
        // Check env vars (if specified in behavior)
        if (this.envVars != null && !this.envVars.isEmpty()) {
            if (envVars == null) {
                return false;
            }
            for (Map.Entry<String, String> entry : this.envVars.entrySet()) {
                if (!entry.getValue().equals(envVars.get(entry.getKey()))) {
                    return false;
                }
            }
        }
        
        return true;
    }

    /**
     * Builder for MockBehavior.
     */
    public static class Builder {
        private String image;
        private String[] cmd;
        private Map<String, String> envVars;
        private int exitCode = 0;
        private String stdout = "";
        private String stderr = "";

        public Builder() {}

        /**
         * Sets the image to match.
         */
        public Builder matchingImage(String image) {
            this.image = image;
            return this;
        }

        /**
         * Sets the command to match.
         */
        public Builder matchingCmd(String... cmd) {
            this.cmd = cmd;
            return this;
        }

        /**
         * Sets environment variables to match.
         */
        public Builder matchingEnvVars(Map<String, String> envVars) {
            this.envVars = envVars;
            return this;
        }

        /**
         * Sets the exit code to return.
         */
        public Builder exitCode(int exitCode) {
            this.exitCode = exitCode;
            return this;
        }

        /**
         * Sets the stdout to return.
         */
        public Builder stdout(String stdout) {
            this.stdout = stdout;
            return this;
        }

        /**
         * Sets the stderr to return.
         */
        public Builder stderr(String stderr) {
            this.stderr = stderr;
            return this;
        }

        /**
         * Builds the MockBehavior.
         */
        public MockBehavior build() {
            return new MockBehavior(this);
        }
    }

    /**
     * Creates a new Builder for MockBehavior.
     */
    public static Builder builder() {
        return new Builder();
    }
}