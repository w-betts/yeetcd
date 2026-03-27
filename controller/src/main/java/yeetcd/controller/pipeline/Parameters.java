package yeetcd.controller.pipeline;

import java.util.Collections;
import java.util.Map;

public record Parameters(Map<String, Parameter> parameters) {

    public static Parameters fromMap(Map<String, Parameter> map) {
        return new Parameters(map);
    }

    public static Parameters empty() {
        return new Parameters(Collections.emptyMap());
    }
}
