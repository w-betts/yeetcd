package yeetcd.sdk;

import lombok.EqualsAndHashCode;
import lombok.ToString;

import java.util.Collections;
import java.util.HashMap;
import java.util.Map;

@EqualsAndHashCode
@ToString
public final class WorkContext {
    private final Map<String, String> workContextMap;

    public WorkContext(Map<String, String> environment) {
        this.workContextMap = environment;
    }

    public Map<String, String> getWorkContextMap() {
        return workContextMap;
    }

    public WorkContext mergeInto(WorkContext workContext) {
        Map<String, String> newContextMap = new HashMap<>();
        newContextMap.putAll(workContext.workContextMap);
        newContextMap.putAll(workContextMap);
        return new WorkContext(newContextMap);
    }

    public static WorkContext empty() {
        return new WorkContext(Collections.emptyMap());
    }

    public static WorkContext of(String k1, String v1) {
        return new WorkContext(Map.of(k1, v1));
    }

    public static WorkContext of(String k1, String v1, String k2, String v2) {
        return new WorkContext(Map.of(k1, v1, k2, v2));
    }

    public static WorkContext of(String k1, String v1, String k2, String v2, String k3, String v3) {
        return new WorkContext(Map.of(k1, v1, k2, v2, k3, v3));
    }

    public static WorkContext of(Map<String, String> map) {
        return new WorkContext(map);
    }
}
