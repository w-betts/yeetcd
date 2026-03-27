package yeetcd.sdk;

import lombok.EqualsAndHashCode;
import lombok.ToString;

import java.util.Collections;
import java.util.Map;

@EqualsAndHashCode
@ToString
public final class Parameters {

    private final Map<String, Parameter> parametersMap;

    public Parameters(Map<String, Parameter> parametersMap) {
        this.parametersMap = parametersMap;
    }

    public Map<String, Parameter> getParametersMap() {
        return parametersMap;
    }

    public static Parameters of(String p1, Parameter v1) {
        return new Parameters(Map.of(p1, v1));
    }

    public static Parameters of(String p1, Parameter v1, String p2, Parameter v2) {
        return new Parameters(Map.of(p1, v1, p2, v2));
    }

    public static Parameters empty() {
        return new Parameters(Collections.emptyMap());
    }


}
