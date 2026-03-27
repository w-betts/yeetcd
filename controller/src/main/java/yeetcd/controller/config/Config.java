package yeetcd.controller.config;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import lombok.Data;
import lombok.SneakyThrows;

import java.io.File;
import java.net.URL;

import static com.fasterxml.jackson.dataformat.yaml.YAMLGenerator.Feature.MINIMIZE_QUOTES;

public class Config {

    private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper(new YAMLFactory().enable(MINIMIZE_QUOTES));

    @SneakyThrows
    public static Kubernetes createKubernetesConfig() {
        return OBJECT_MAPPER.readValue(configUrl(), Kubernetes.class);
    }

    @SneakyThrows
    public static Kubernetes createKubernetesConfig(URL configUrl) {
        return OBJECT_MAPPER.readValue(configUrl, Kubernetes.class);
    }

    @SneakyThrows
    private static URL configUrl() {
        String yeetcdControllerConfigFilePath = System.getenv("AULOS_CONTROLLER_CONFIG_FILE") == null ? "/etc/yeetcd/yeetcd-controller.yaml" : System.getenv("AULOS_CONTROLLER_CONFIG_FILE");
        File yeetcdControllerConfigFile = new File(yeetcdControllerConfigFilePath);
        if (yeetcdControllerConfigFile.exists()) {
            return yeetcdControllerConfigFile.toURI().toURL();
        }

        URL classpathConfig = Config.class.getClassLoader().getResource("yeetcd-controller.yaml");
        if (classpathConfig != null) {
            return classpathConfig;
        }

        throw new IllegalStateException("No config file found on file system or on classpath");
    }

    @Data
    public static class Kubernetes {

        private Registry registry;

        @Data
        public static class Registry {
            private String pushAddress;
            private String pullAddress;
        }
    }
}
