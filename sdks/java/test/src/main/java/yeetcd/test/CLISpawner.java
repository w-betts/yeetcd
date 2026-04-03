package yeetcd.test;

import java.io.*;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

/**
 * Spawns the CLI binary with classpath and mock server address.
 * Detects platform-specific CLI binary from classpath resources,
 * spawns subprocess with YEETCD_MOCK_ADDRESS env var.
 */
public class CLISpawner {

    private final String classpath;
    private final String mockAddress;

    private CLISpawner(Builder builder) {
        this.classpath = builder.classpath;
        this.mockAddress = builder.mockAddress;
    }

    /**
     * Spawns the CLI as a subprocess.
     * 
     * @return the process
     * @throws IOException if CLI binary cannot be found
     * @throws InterruptedException if the process is interrupted
     */
    public Process spawn() throws IOException, InterruptedException {
        // Get the binary name for current platform
        String binaryName = PlatformDetector.getBinaryName();
        
        // Try to find the binary in classpath resources
        String binaryPath = findBinaryInClasspath(binaryName);
        
        if (binaryPath == null) {
            throw new IOException("CLI binary not found in classpath: " + binaryName);
        }
        
        // Build command
        List<String> command = new ArrayList<>();
        command.add(binaryPath);
        
        // Build environment
        Map<String, String> env = System.getenv();
        // Add YEETCD_MOCK_ADDRESS if mock address is set
        // (ProcessBuilder will inherit all env vars by default)
        
        // Create process
        ProcessBuilder pb = new ProcessBuilder(command);
        pb.environment().putAll(env);
        
        if (mockAddress != null) {
            pb.environment().put("YEETCD_MOCK_ADDRESS", mockAddress);
        }
        
        return pb.start();
    }

    /**
     * Finds the binary in classpath resources.
     */
    private String findBinaryInClasspath(String binaryName) {
        // Check if the binary exists as a resource
        String resourcePath = "/cli/" + binaryName;
        
        // Try to load as resource to verify it exists
        if (getClass().getResource(resourcePath) != null) {
            // For resources, we need to extract to a temp file
            // This is a simplified version - in production you'd extract properly
            return resourcePath;
        }
        
        // Fallback: check in working directory (for development)
        File workingDirBinary = new File("bin", binaryName);
        if (workingDirBinary.exists()) {
            return workingDirBinary.getAbsolutePath();
        }
        
        return null;
    }

    /**
     * Builder for CLISpawner.
     */
    public static class Builder {
        private String classpath;
        private String mockAddress;

        public Builder() {}

        public Builder classpath(String classpath) {
            this.classpath = classpath;
            return this;
        }

        public Builder mockAddress(String mockAddress) {
            this.mockAddress = mockAddress;
            return this;
        }

        public CLISpawner build() {
            return new CLISpawner(this);
        }
    }

    /**
     * Creates a new Builder for CLISpawner.
     */
    public static Builder builder() {
        return new Builder();
    }
}