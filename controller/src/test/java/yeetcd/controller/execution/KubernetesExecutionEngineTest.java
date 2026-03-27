package yeetcd.controller.execution;

import yeetcd.controller.config.Config;
import io.kubernetes.client.openapi.ApiClient;
import io.kubernetes.client.util.ClientBuilder;
import io.kubernetes.client.util.KubeConfig;
import lombok.SneakyThrows;
import org.junit.jupiter.api.Disabled;

import java.io.InputStreamReader;
import java.net.URL;
import java.util.Objects;

@Disabled
public class KubernetesExecutionEngineTest extends AbstractExecutionEngineTest {

    Config.Kubernetes kubernetesConfig = getKubernetesConfig();

    @Override
    String builtImagePullAddress() {
        return kubernetesConfig.getRegistry().getPullAddress();
    }

    @Override
    ExecutionEngine executionEngine() {
        return new KubernetesExecutionEngine(
            kubernetesConfig,
            testApiClient(),
            true
        );
    }

    private static Config.Kubernetes getKubernetesConfig() {
        String configFile = "yeetcd-controller.yaml";
        URL yeetcdControllerConfig = AbstractExecutionEngineTest.class.getClassLoader().getResource(configFile);
        if (yeetcdControllerConfig == null) {
            throw new RuntimeException("Config file '%s' not found on classpath. Make sure you have run `./local-k8s.sh start`".formatted(configFile));
        }
        return Config.createKubernetesConfig(yeetcdControllerConfig);
    }

    @SneakyThrows
    private ApiClient testApiClient() {
        try (InputStreamReader kubeconfigStreamReader = new InputStreamReader(Objects.requireNonNull(
            AbstractExecutionEngineTest.class.getClassLoader().getResourceAsStream("kubeconfig")
        ))
        ) {
            return ClientBuilder.kubeconfig(KubeConfig.loadKubeConfig(kubeconfigStreamReader)).build();
        }
    }
}
