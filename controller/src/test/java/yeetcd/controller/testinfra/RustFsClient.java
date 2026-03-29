package yeetcd.controller.testinfra;

import lombok.extern.slf4j.Slf4j;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.concurrent.TimeUnit;

/**
 * Client for RustFS S3-compatible storage using mc (MinIO Client) CLI.
 * 
 * Used for test verification and file upload/download operations.
 * Executes mc commands via ProcessBuilder for reliability with RustFS.
 * 
 * Prerequisites:
 * - mc CLI must be installed and available on PATH
 * - RustFS alias must be configured (use configureAlias() before operations)
 */
@Slf4j
public class RustFsClient {

    private static final String ALIAS = "rustfs";
    private static final int COMMAND_TIMEOUT_SECONDS = 30;
    
    private final String endpoint;
    private final String accessKey;
    private final String secretKey;
    private final String bucketName;
    
    private boolean aliasConfigured = false;

    /**
     * Creates a RustFsClient.
     * 
     * @param endpoint RustFS endpoint URL (e.g., "http://localhost:9000")
     * @param accessKey RustFS access key
     * @param secretKey RustFS secret key
     * @param bucketName Default bucket name for operations
     */
    public RustFsClient(String endpoint, String accessKey, String secretKey, String bucketName) {
        this.endpoint = endpoint;
        this.accessKey = accessKey;
        this.secretKey = secretKey;
        this.bucketName = bucketName;
    }

    /**
     * Configures the mc alias for RustFS.
     * Must be called before other operations.
     * 
     * @throws TestInfrastructureException if configuration fails
     */
    public void configureAlias() {
        log.info("Configuring mc alias '{}' for endpoint '{}'...", ALIAS, endpoint);
        
        // Remove existing alias if present (ignore errors)
        executeCommand("mc", "alias", "rm", ALIAS, "--force");
        
        // Add new alias
        McResult result = executeCommand(
            "mc", "alias", "set", ALIAS, endpoint, accessKey, secretKey
        );
        
        if (!result.success) {
            throw new TestInfrastructureException(
                "MC_ALIAS_CONFIG_FAILED",
                "Failed to configure mc alias for RustFS",
                "mc alias '" + ALIAS + "'",
                "exit code " + result.exitCode + ", stderr: " + result.stderr,
                "Ensure mc CLI is installed and RustFS is accessible at " + endpoint,
                null
            );
        }
        
        aliasConfigured = true;
        log.info("mc alias '{}' configured successfully", ALIAS);
    }

    /**
     * Ensures the bucket exists, creating it if necessary.
     * 
     * @throws TestInfrastructureException if bucket creation fails
     */
    public void ensureBucketExists() {
        ensureAliasConfigured();
        
        log.debug("Ensuring bucket '{}' exists...", bucketName);
        
        // Try to create bucket (will fail if already exists, which is OK)
        McResult result = executeCommand("mc", "mb", ALIAS + "/" + bucketName, "--ignore-existing");
        
        if (!result.success) {
            throw new TestInfrastructureException(
                "MC_BUCKET_CREATE_FAILED",
                "Failed to create bucket '" + bucketName + "'",
                "bucket '" + bucketName + "'",
                "exit code " + result.exitCode + ", stderr: " + result.stderr,
                "Check RustFS is running and accessible. Verify with 'mc ls " + ALIAS + "'",
                null
            );
        }
        
        log.debug("Bucket '{}' exists", bucketName);
    }

    /**
     * Uploads content to RustFS as a file.
     * 
     * @param key Object key (path within bucket)
     * @param content File content
     * @throws TestInfrastructureException if upload fails
     */
    public void uploadFile(String key, String content) {
        ensureAliasConfigured();
        ensureBucketExists();
        
        // Strip leading slash from key (mc doesn't handle keys starting with /)
        key = normalizeKey(key);
        
        log.debug("Uploading file '{}' to bucket '{}'...", key, bucketName);
        
        try {
            // Write content to temp file
            Path tempFile = Files.createTempFile("rustfs-upload-", ".txt");
            try {
                Files.writeString(tempFile, content, StandardCharsets.UTF_8);
                
                // Upload using mc cp
                McResult result = executeCommand(
                    "mc", "cp", tempFile.toString(), ALIAS + "/" + bucketName + "/" + key
                );
                
                if (!result.success) {
                    throw new TestInfrastructureException(
                        "MC_UPLOAD_FAILED",
                        "Failed to upload file '" + key + "' to bucket '" + bucketName + "'",
                        "object '" + key + "'",
                        "exit code " + result.exitCode + ", stderr: " + result.stderr,
                        "Check bucket exists and has write permissions. Try 'mc ls " + ALIAS + "/" + bucketName + "'",
                        null
                    );
                }
                
                log.debug("File '{}' uploaded successfully", key);
            } finally {
                Files.deleteIfExists(tempFile);
            }
        } catch (IOException e) {
            throw new TestInfrastructureException(
                "MC_UPLOAD_TEMP_FILE_FAILED",
                "Failed to create temp file for upload",
                "temp file for '" + key + "'",
                "exception: " + e.getMessage(),
                "Check disk space and permissions for temp directory",
                e
            );
        }
    }

    /**
     * Downloads a file from RustFS.
     * 
     * @param key Object key (path within bucket)
     * @return File content as string
     * @throws TestInfrastructureException if download fails
     */
    public String downloadFile(String key) {
        ensureAliasConfigured();
        
        // Strip leading slash from key
        key = normalizeKey(key);
        
        log.debug("Downloading file '{}' from bucket '{}'...", key, bucketName);
        
        try {
            // Download to temp file
            Path tempFile = Files.createTempFile("rustfs-download-", ".txt");
            try {
                McResult result = executeCommand(
                    "mc", "cp", ALIAS + "/" + bucketName + "/" + key, tempFile.toString()
                );
                
                if (!result.success) {
                    throw new TestInfrastructureException(
                        "MC_DOWNLOAD_FAILED",
                        "Failed to download file '" + key + "' from bucket '" + bucketName + "'",
                        "object '" + key + "'",
                        "exit code " + result.exitCode + ", stderr: " + result.stderr,
                        "Check file exists. Try 'mc ls " + ALIAS + "/" + bucketName + "/" + key + "'",
                        null
                    );
                }
                
                String content = Files.readString(tempFile, StandardCharsets.UTF_8);
                log.debug("File '{}' downloaded successfully ({} bytes)", key, content.length());
                return content;
            } finally {
                Files.deleteIfExists(tempFile);
            }
        } catch (IOException e) {
            throw new TestInfrastructureException(
                "MC_DOWNLOAD_TEMP_FILE_FAILED",
                "Failed to create temp file for download",
                "temp file for '" + key + "'",
                "exception: " + e.getMessage(),
                "Check disk space and permissions for temp directory",
                e
            );
        }
    }

    /**
     * Checks if a file exists in RustFS.
     * 
     * @param key Object key (path within bucket)
     * @return true if file exists, false otherwise
     */
    public boolean fileExists(String key) {
        ensureAliasConfigured();
        
        // Strip leading slash from key
        key = normalizeKey(key);
        
        McResult result = executeCommand("mc", "stat", ALIAS + "/" + bucketName + "/" + key);
        return result.success;
    }

    /**
     * Lists files in a directory/prefix in RustFS.
     * 
     * @param prefix Directory prefix to list (e.g., "outputs/")
     * @return List of object keys (paths) found
     * @throws TestInfrastructureException if listing fails
     */
    public java.util.List<String> listFiles(String prefix) {
        ensureAliasConfigured();
        
        // Strip leading slash from prefix
        prefix = normalizeKey(prefix);
        
        log.debug("Listing files with prefix '{}' in bucket '{}'...", prefix, bucketName);
        
        McResult result = executeCommand(
            "mc", "ls", "--recursive", ALIAS + "/" + bucketName + "/" + prefix
        );
        
        if (!result.success) {
            throw new TestInfrastructureException(
                "MC_LIST_FAILED",
                "Failed to list files with prefix '" + prefix + "' in bucket '" + bucketName + "'",
                "prefix '" + prefix + "'",
                "exit code " + result.exitCode + ", stderr: " + result.stderr,
                "Check bucket exists. Try 'mc ls " + ALIAS + "/" + bucketName + "'",
                null
            );
        }
        
        // Parse mc ls output - format is like:
        // [2026-03-29 00:49:21 GMT]    36B STANDARD 24e03ee5-a611-461d-8249-4c669f549ba7/inputs/...
        // Format: [date time timezone]  size storageClass path
        // Note: mc ls --recursive returns paths relative to the prefix, so we need to prepend the prefix
        java.util.List<String> files = new java.util.ArrayList<>();
        for (String line : result.stdout().split("\n")) {
            if (line.isBlank()) continue;
            // Extract the path part (after the storage class)
            // Format: [date time timezone]  size storageClass path
            int lastBracket = line.lastIndexOf(']');
            if (lastBracket >= 0) {
                String rest = line.substring(lastBracket + 1).trim();
                // Split by whitespace: size, storageClass, path
                String[] parts = rest.split("\\s+", 3);
                if (parts.length >= 3) {
                    String path = parts[2].trim();
                    // Prepend prefix to get full path within bucket
                    String fullPath = prefix.isEmpty() ? path : prefix + path;
                    files.add(fullPath);
                }
            }
        }
        
        log.debug("Found {} files with prefix '{}'", files.size(), prefix);
        return files;
    }

    /**
     * Deletes a file from RustFS.
     * 
     * @param key Object key (path within bucket)
     * @throws TestInfrastructureException if deletion fails
     */
    public void deleteFile(String key) {
        ensureAliasConfigured();
        
        // Strip leading slash from key
        key = normalizeKey(key);
        
        log.debug("Deleting file '{}' from bucket '{}'...", key, bucketName);
        
        McResult result = executeCommand("mc", "rm", ALIAS + "/" + bucketName + "/" + key);
        
        if (!result.success) {
            throw new TestInfrastructureException(
                "MC_DELETE_FAILED",
                "Failed to delete file '" + key + "' from bucket '" + bucketName + "'",
                "object '" + key + "'",
                "exit code " + result.exitCode + ", stderr: " + result.stderr,
                "Check file exists. Try 'mc ls " + ALIAS + "/" + bucketName + "'",
                null
            );
        }
        
        log.debug("File '{}' deleted successfully", key);
    }

    private void ensureAliasConfigured() {
        if (!aliasConfigured) {
            configureAlias();
        }
    }
    
    /**
     * Normalizes a key by stripping leading slashes.
     * mc CLI doesn't handle keys starting with / correctly.
     */
    private String normalizeKey(String key) {
        while (key.startsWith("/")) {
            key = key.substring(1);
        }
        return key;
    }

    /**
     * Executes a command and returns the result.
     */
    private McResult executeCommand(String... command) {
        try {
            ProcessBuilder pb = new ProcessBuilder(command);
            pb.redirectErrorStream(false);
            
            log.debug("Executing: {}", String.join(" ", command));
            
            Process process = pb.start();
            
            // Capture stdout
            String stdout = readStream(process.getInputStream());
            String stderr = readStream(process.getErrorStream());
            
            boolean completed = process.waitFor(COMMAND_TIMEOUT_SECONDS, TimeUnit.SECONDS);
            
            if (!completed) {
                process.destroyForcibly();
                throw new TestInfrastructureException(
                    "MC_COMMAND_TIMEOUT",
                    "Command timed out after " + COMMAND_TIMEOUT_SECONDS + "s",
                    "command: " + String.join(" ", command),
                    "timeout",
                    "Check RustFS is responsive. Try 'mc admin info " + ALIAS + "'",
                    null
                );
            }
            
            int exitCode = process.exitValue();
            boolean success = exitCode == 0;
            
            if (!success) {
                log.debug("Command failed with exit code {}: {}", exitCode, stderr);
            }
            
            return new McResult(success, exitCode, stdout, stderr);
            
        } catch (IOException | InterruptedException e) {
            throw new TestInfrastructureException(
                "MC_COMMAND_EXECUTION_FAILED",
                "Failed to execute mc command",
                "command: " + String.join(" ", command),
                "exception: " + e.getMessage(),
                "Ensure mc CLI is installed and available on PATH",
                e
            );
        }
    }

    private String readStream(InputStream stream) throws IOException {
        StringBuilder sb = new StringBuilder();
        try (BufferedReader reader = new BufferedReader(new InputStreamReader(stream, StandardCharsets.UTF_8))) {
            String line;
            while ((line = reader.readLine()) != null) {
                sb.append(line).append("\n");
            }
        }
        return sb.toString().trim();
    }

    private record McResult(boolean success, int exitCode, String stdout, String stderr) {}
}
