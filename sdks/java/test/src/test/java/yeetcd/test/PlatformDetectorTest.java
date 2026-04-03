package yeetcd.test;

import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

/**
 * Unit tests for PlatformDetector.
 */
class PlatformDetectorTest {

    @Test
    void testGetOSReturnsKnownValue() {
        String os = PlatformDetector.getOS();
        
        // Should be one of the supported values
        assertTrue(os.equals("darwin") || os.equals("linux") || os.equals("windows"),
                "OS should be darwin, linux, or windows but was: " + os);
    }

    @Test
    void testGetArchReturnsKnownValue() {
        String arch = PlatformDetector.getArch();
        
        // Should be one of the supported values
        assertTrue(arch.equals("amd64") || arch.equals("arm64"),
                "Arch should be amd64 or arm64 but was: " + arch);
    }

    @Test
    void testGetPlatformCombinesOSAndArch() {
        String platform = PlatformDetector.getPlatform();
        
        // Should be os-arch format
        assertTrue(platform.matches("^darwin-(amd64|arm64)$") || 
                   platform.matches("^linux-(amd64|arm64)$") ||
                   platform.matches("^windows-(amd64|arm64)$"),
                "Platform should match os-arch pattern but was: " + platform);
    }

    @Test
    void testGetBinaryNameIncludesPlatform() {
        String binaryName = PlatformDetector.getBinaryName();
        
        // Should start with "yeetcd-"
        assertTrue(binaryName.startsWith("yeetcd-"), 
                "Binary name should start with 'yeetcd-' but was: " + binaryName);
        
        // Should contain platform
        String platform = PlatformDetector.getPlatform();
        assertTrue(binaryName.endsWith(platform),
                "Binary name should end with platform but was: " + binaryName);
    }

    @Test
    void testIsSupportedReturnsTrueOnSupportedPlatform() {
        // On any正常运行 system, should return true
        assertTrue(PlatformDetector.isSupported());
    }
}