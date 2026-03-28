package yeetcd.controller.execution;

import software.amazon.awssdk.services.s3.S3Client;

/**
 * Factory for creating S3 clients to connect to RustFS.
 * Used for test verification and file upload/download operations.
 */
public class S3ClientFactory {

    private final String endpoint;
    private final String accessKey;
    private final String secretKey;

    public S3ClientFactory(String endpoint, String accessKey, String secretKey) {
        this.endpoint = endpoint;
        this.accessKey = accessKey;
        this.secretKey = secretKey;
    }

    /**
     * Creates an S3 client configured to connect to RustFS.
     * 
     * @return configured S3Client
     */
    public S3Client createClient() {
        return S3Client.builder()
            .endpointOverride(java.net.URI.create(endpoint))
            .credentialsProvider(() -> software.amazon.awssdk.auth.credentials.AwsBasicCredentials.create(accessKey, secretKey))
            .region(software.amazon.awssdk.regions.Region.US_EAST_1) // RustFS doesn't use regions, but SDK requires one
            .build();
    }

    /**
     * Gets the S3 endpoint URL.
     */
    public String getEndpoint() {
        return endpoint;
    }

    /**
     * Gets the S3 access key.
     */
    public String getAccessKey() {
        return accessKey;
    }

    /**
     * Gets the S3 secret key.
     */
    public String getSecretKey() {
        return secretKey;
    }
}
