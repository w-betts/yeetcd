package yeetcd.controller.execution;

import yeetcd.controller.config.Config;
import yeetcd.controller.testinfra.TestClusterFixture;
import io.kubernetes.client.openapi.ApiClient;
import lombok.SneakyThrows;
import org.junit.jupiter.api.extension.ExtendWith;

/**
 * Integration tests for KubernetesExecutionEngine.
 * 
 * Uses TestClusterFixture to manage k3d cluster lifecycle.
 * The cluster is automatically created if missing, and test resources are cleaned up.
 */
@ExtendWith(TestClusterFixture.class)
public class KubernetesExecutionEngineTest extends AbstractExecutionEngineTest {

    private PipelinePvcManager pvcManager;
    private S3ClientFactory s3ClientFactory;

    @Override
    String builtImagePullAddress() {
        return TestClusterFixture.getRegistryPullAddress();
    }

    @Override
    ExecutionEngine executionEngine() {
        Config.Kubernetes config = getKubernetesConfig();
        ApiClient apiClient = testApiClient();
        
        pvcManager = new PipelinePvcManager(apiClient);
        s3ClientFactory = new S3ClientFactory(
            config.getS3().getEndpoint(),
            config.getS3().getAccessKey(),
            config.getS3().getSecretKey()
        );
        
        return new KubernetesExecutionEngine(
            config,
            apiClient,
            true,
            pvcManager,
            s3ClientFactory
        );
    }

    private static Config.Kubernetes getKubernetesConfig() {
        // Build config dynamically from test infrastructure
        Config.Kubernetes config = new Config.Kubernetes();
        
        Config.Kubernetes.Registry registry = new Config.Kubernetes.Registry();
        registry.setPushAddress(TestClusterFixture.getRegistryPushAddress());
        registry.setPullAddress(TestClusterFixture.getRegistryPullAddress());
        config.setRegistry(registry);
        
        Config.Kubernetes.S3 s3 = new Config.Kubernetes.S3();
        s3.setEndpoint("http://rustfs.yeetcd.svc.cluster.local:9000");
        s3.setAccessKey("rustfs");
        s3.setSecretKey("rustfs-secret");
        s3.setBucketName("yeetcd-pipelines");
        config.setS3(s3);
        
        return config;
    }

    @SneakyThrows
    private ApiClient testApiClient() {
        return TestClusterFixture.getApiClient();
    }
}
