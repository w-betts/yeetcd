package yeetcd.test;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

import java.io.BufferedReader;
import java.io.File;
import java.io.IOException;
import java.io.InputStreamReader;
import java.nio.file.Path;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Unit tests for Makefile build targets.
 * Tests that various make build targets execute correctly.
 */
class MakefileBuildTest {

    /**
     * Gets the project root directory (where Makefile and golang/ exist).
     * Searches upward from current directory until finding Makefile and golang/.
     */
    private File getProjectRoot() {
        // Start from user.dir and search upward for project root
        File currentDir = new File(System.getProperty("user.dir"));
        
        // Search up to 10 levels to find project root
        for (int i = 0; i < 10; i++) {
            File makefile = new File(currentDir, "Makefile");
            File golangDir = new File(currentDir, "golang");
            if (makefile.exists() && golangDir.exists() && golangDir.isDirectory()) {
                return currentDir;
            }
            File parent = currentDir.getParentFile();
            if (parent == null) break;
            currentDir = parent;
        }
        
        // Fallback: try going up from sdks/java/test/target/test-classes
        // which is where tests run from
        return new File(System.getProperty("user.dir"))
            .getParentFile() // target
            .getParentFile() // test
            .getParentFile() // sdks/java/test
            .getParentFile() // sdks/java
            .getParentFile(); // project root
    }

    /**
     * Test that 'make build' executes successfully.
     * GIVEN: Go toolchain installed and golang/cmd/yeetcd exists
     * WHEN: make build is executed
     * THEN: command succeeds (exit code 0)
     */
    @Test
    void testMakeBuildExecutesSuccessfully() throws IOException, InterruptedException {
        File projectRoot = getProjectRoot();
        
        ProcessBuilder pb = new ProcessBuilder("make", "build");
        pb.directory(projectRoot);
        pb.redirectErrorStream(true);
        
        Process process = pb.start();
        int exitCode = process.waitFor();
        
        assertEquals(0, exitCode, "make build should execute successfully");
    }

    /**
     * Test that 'make build' creates the binary at bin/yeetcd.
     * GIVEN: make build has been executed
     * WHEN: checking for bin/yeetcd
     * THEN: binary file exists
     */
    @Test
    void testMakeBuildCreatesBinary() throws IOException, InterruptedException {
        File projectRoot = getProjectRoot();
        File binary = new File(projectRoot, "bin/yeetcd");
        
        // Run make build
        ProcessBuilder pb = new ProcessBuilder("make", "build");
        pb.directory(projectRoot);
        pb.redirectErrorStream(true);
        
        Process process = pb.start();
        int exitCode = process.waitFor();
        
        assertEquals(0, exitCode, "make build should execute successfully");
        assertTrue(binary.exists(), "bin/yeetcd should exist after make build");
    }

    /**
     * Test that 'make build-all' executes successfully (may fail if target doesn't exist yet).
     * GIVEN: Go toolchain installed and golang/cmd/yeetcd exists
     * WHEN: make build-all is executed
     * THEN: command succeeds and creates all platform binaries in bin/cli/
     */
    @Test
    void testMakeBuildAllExecutesSuccessfully() throws IOException, InterruptedException {
        File projectRoot = getProjectRoot();
        
        ProcessBuilder pb = new ProcessBuilder("make", "build-all");
        pb.directory(projectRoot);
        pb.redirectErrorStream(true);
        
        Process process = pb.start();
        int exitCode = process.waitFor();
        
        // This may fail if target doesn't exist yet - that's expected for stubs
        if (exitCode != 0) {
            fail("make build-all target may not exist yet or failed. Exit code: " + exitCode);
        }
        
        // If it succeeds, verify the binaries exist
        String[] platforms = {"darwin-amd64", "darwin-arm64", "linux-amd64", "linux-arm64"};
        File cliDir = new File(projectRoot, "bin/cli");
        
        for (String platform : platforms) {
            File binary = new File(cliDir, "yeetcd-" + platform);
            assertTrue(binary.exists(), "bin/cli/yeetcd-" + platform + " should exist");
        }
    }

    /**
     * Test that 'make build-darwin-arm64' executes and creates correct binary.
     * GIVEN: Go toolchain installed
     * WHEN: make build-darwin-arm64 is executed
     * THEN: bin/cli/yeetcd-darwin-arm64 binary is created
     */
    @Test
    void testMakeBuildDarwinArm64CreatesBinary() throws IOException, InterruptedException {
        File projectRoot = getProjectRoot();
        
        ProcessBuilder pb = new ProcessBuilder("make", "build-darwin-arm64");
        pb.directory(projectRoot);
        pb.redirectErrorStream(true);
        
        Process process = pb.start();
        int exitCode = process.waitFor();
        
        // This may fail if target doesn't exist yet - that's expected for stubs
        if (exitCode != 0) {
            fail("make build-darwin-arm64 target may not exist yet or failed. Exit code: " + exitCode);
        }
        
        File binary = new File(projectRoot, "bin/cli/yeetcd-darwin-arm64");
        assertTrue(binary.exists(), "bin/cli/yeetcd-darwin-arm64 should exist");
    }

    /**
     * Test that built binary is executable and responds to --version.
     * GIVEN: make build has been executed
     * WHEN: bin/yeetcd --version is executed
     * THEN: version information is printed to stdout
     */
    @Test
    void testBinaryIsExecutableAndShowsVersion() throws IOException, InterruptedException {
        File projectRoot = getProjectRoot();
        
        // First ensure binary exists by running make build
        ProcessBuilder buildPb = new ProcessBuilder("make", "build");
        buildPb.directory(projectRoot);
        buildPb.redirectErrorStream(true);
        Process buildProcess = buildPb.start();
        assertEquals(0, buildProcess.waitFor(), "make build should succeed");
        
        // Now run --version
        ProcessBuilder versionPb = new ProcessBuilder("./bin/yeetcd", "--version");
        versionPb.directory(projectRoot);
        
        Process process = versionPb.start();
        
        StringBuilder output = new StringBuilder();
        try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
            String line;
            while ((line = reader.readLine()) != null) {
                output.append(line).append("\n");
            }
        }
        
        int exitCode = process.waitFor();
        
        assertEquals(0, exitCode, "bin/yeetcd --version should succeed");
        String versionOutput = output.toString();
        assertFalse(versionOutput.isEmpty(), "Version output should not be empty");
        assertTrue(versionOutput.toLowerCase().contains("yeetcd"), "Version should mention 'yeetcd'");
    }

    /**
     * Test that PlatformDetector.getOS() returns correct OS.
     * GIVEN: running on darwin or linux
     * WHEN: PlatformDetector.getOS() is called
     * THEN: returns 'darwin' or 'linux' matching the current OS
     */
    @Test
    void testPlatformDetectorGetOS() {
        String os = PlatformDetector.getOS();
        
        // Should be one of the supported values
        assertTrue(os.equals("darwin") || os.equals("linux") || os.equals("windows"),
                "OS should be darwin, linux, or windows but was: " + os);
        
        // Verify it matches actual OS
        String actualOs = System.getProperty("os.name", "").toLowerCase();
        if (actualOs.contains("mac") || actualOs.contains("darwin")) {
            assertEquals("darwin", os, "Should detect darwin on Mac");
        } else if (actualOs.contains("linux")) {
            assertEquals("linux", os, "Should detect linux on Linux");
        }
    }

    /**
     * Test that PlatformDetector.getArch() returns correct architecture.
     * GIVEN: running on x86_64 or arm64
     * WHEN: PlatformDetector.getArch() is called
     * THEN: returns 'amd64' or 'arm64' matching the current architecture
     */
    @Test
    void testPlatformDetectorGetArch() {
        String arch = PlatformDetector.getArch();
        
        // Should be one of the supported values
        assertTrue(arch.equals("amd64") || arch.equals("arm64"),
                "Arch should be amd64 or arm64 but was: " + arch);
        
        // Verify it matches actual architecture
        String actualArch = System.getProperty("os.arch", "").toLowerCase();
        if (actualArch.equals("x86_64") || actualArch.equals("amd64")) {
            assertEquals("amd64", arch, "Should detect amd64 on x86_64");
        } else if (actualArch.equals("aarch64") || actualArch.equals("arm64")) {
            assertEquals("arm64", arch, "Should detect arm64 on aarch64/arm64");
        }
    }
}