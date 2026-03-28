package yeetcd.controller.execution;

import yeetcd.controller.config.Config;
import yeetcd.controller.testinfra.TestClusterFixture;
import io.kubernetes.client.openapi.ApiClient;
import lombok.SneakyThrows;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.extension.ExtendWith;

/**
 * Integration tests for KubernetesExecutionEngine.
 * 
 * Uses TestClusterFixture to manage k3d cluster lifecycle.
 * The cluster is automatically created if missing, and test resources are cleaned up.
 */
@Disabled("Pending Phase 3: PVC support implementation")
@ExtendWith(TestClusterFixture.class)
public class KubernetesExecutionEngineTest extends AbstractExecutionEngineTest {

    @Override
    String builtImagePullAddress() {
        return TestClusterFixture.getRegistryPullAddress();
    }

    @Override
    ExecutionEngine executionEngine() {
        return new KubernetesExecutionEngine(
            getKubernetesConfig(),
            testApiClient(),
            true
        );
    }

    private static Config.Kubernetes getKubernetesConfig() {
        // Build config dynamically from test infrastructure
        Config.Kubernetes config = new Config.Kubernetes();
        
        Config.Kubernetes.Registry registry = new Config.Kubernetes.Registry();
        registry.setPushAddress(TestClusterFixture.getRegistryPushAddress());
        registry.setPullAddress(TestClusterFixture.getRegistryPullAddress());
        config.setRegistry(registry);
        
        return config;
    }

    @SneakyThrows
    private ApiClient testApiClient() {
        return TestClusterFixture.getApiClient();
    }
}
