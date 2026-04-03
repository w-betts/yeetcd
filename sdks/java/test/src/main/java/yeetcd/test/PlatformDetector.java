package yeetcd.test;

/**
 * Utility class for detecting the operating system and architecture at runtime.
 * Used to select the correct CLI binary from bundled resources.
 */
public class PlatformDetector {

    private PlatformDetector() {
        // Utility class, no instantiation
    }

    /**
     * Gets the operating system name (lowercase).
     * 
     * @return "darwin" or "linux"
     */
    public static String getOS() {
        String osName = System.getProperty("os.name", "").toLowerCase();
        
        if (osName.contains("mac") || osName.contains("darwin")) {
            return "darwin";
        } else if (osName.contains("linux")) {
            return "linux";
        } else if (osName.contains("windows")) {
            return "windows";
        }
        
        throw new IllegalStateException("Unsupported operating system: " + osName);
    }

    /**
     * Gets the CPU architecture (lowercase).
     * 
     * @return "amd64" or "arm64"
     */
    public static String getArch() {
        String osArch = System.getProperty("os.arch", "").toLowerCase();
        
        // x86_64, amd64, i386, i486, i586, i686 -> amd64
        if (osArch.equals("x86_64") || osArch.equals("amd64") || osArch.startsWith("i3") || osArch.startsWith("i4") || osArch.startsWith("i5") || osArch.startsWith("i6")) {
            return "amd64";
        }
        
        // aarch64, arm64 -> arm64
        if (osArch.equals("aarch64") || osArch.equals("arm64")) {
            return "arm64";
        }
        
        throw new IllegalStateException("Unsupported architecture: " + osArch);
    }

    /**
     * Gets the combined platform string (e.g., "darwin-arm64", "linux-amd64").
     * 
     * @return the platform string
     */
    public static String getPlatform() {
        return getOS() + "-" + getArch();
    }

    /**
     * Gets the CLI binary name for this platform.
     * 
     * @return the binary name (e.g., "yeetcd-darwin-arm64")
     */
    public static String getBinaryName() {
        return "yeetcd-" + getPlatform();
    }

    /**
     * Checks if the current platform is supported.
     * 
     * @return true if supported, false otherwise
     */
    public static boolean isSupported() {
        try {
            getPlatform();
            return true;
        } catch (IllegalStateException e) {
            return false;
        }
    }
}