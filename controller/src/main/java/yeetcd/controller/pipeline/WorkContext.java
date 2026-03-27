package yeetcd.controller.pipeline;

import java.util.Collections;
import java.util.HashMap;
import java.util.Map;

public record WorkContext(Map<String, String> workContextMap) {

    public static WorkContext empty() {
        return new WorkContext(Map.of());
    }

    public static WorkContext fromMap(Map<String, String> map) {
        return new WorkContext(Collections.unmodifiableMap(map));
    }

    public WorkContext mergeInto(WorkContext workContext) {
        Map<String, String> newContextMap = new HashMap<>();
        newContextMap.putAll(workContext.workContextMap);
        newContextMap.putAll(workContextMap);
        return new WorkContext(newContextMap);
    }
}
