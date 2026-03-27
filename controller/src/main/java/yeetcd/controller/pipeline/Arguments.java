package yeetcd.controller.pipeline;

import org.apache.commons.lang3.StringUtils;

import java.util.HashMap;
import java.util.Map;

public record Arguments(Map<String, String> arguments) {

    WorkContext asValidatedWorkContext(Parameters parameters) {
        Map<String, String> validArguments = new HashMap<>();
        arguments.forEach((key, value) -> {
            Parameter parameter = parameters.parameters().get(key);
            if (parameter == null) {
                throw new IllegalArgumentException("Unsupported argument '%s'".formatted(key));
            }
            parameter.validateArgument(value);
            validArguments.put(key, value);
        });
        parameters.parameters().forEach((key, value) -> {
            if (value.required() && StringUtils.isBlank(validArguments.get(key))) {
                throw new IllegalArgumentException("Required argument '%s' is missing".formatted((key)));
            }
        });
        return WorkContext.fromMap(validArguments);
    }

    public static Arguments of(String p1, String v1) {
        return new Arguments(Map.of(p1, v1));
    }

    public static Arguments of(String p1, String v1, String p2, String v2) {
        return new Arguments(Map.of(p1, v1, p2, v2));
    }
}
