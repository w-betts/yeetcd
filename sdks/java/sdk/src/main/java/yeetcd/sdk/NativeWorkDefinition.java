package yeetcd.sdk;

import com.fasterxml.jackson.annotation.JsonAutoDetect;
import com.fasterxml.jackson.annotation.PropertyAccessor;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import org.apache.commons.codec.digest.DigestUtils;

import java.util.stream.Stream;

public abstract class NativeWorkDefinition implements Runnable, WorkDefinition {

    private final ObjectMapper OBJECT_MAPPER = new ObjectMapper()
            .setVisibility(PropertyAccessor.ALL, JsonAutoDetect.Visibility.NONE)
            .setVisibility(PropertyAccessor.FIELD, JsonAutoDetect.Visibility.ANY)
            .configure(SerializationFeature.ORDER_MAP_ENTRIES_BY_KEYS, true);

    String executionId() {
        try {
            return DigestUtils.sha256Hex(this.getClass().getName() + "%" + OBJECT_MAPPER.writeValueAsString(this));
        } catch (JsonProcessingException e) {
            throw new RuntimeException("Unable to serialise object to executionId", e);
        }
    }

    @Override
    public Stream<NativeWorkDefinition> nativeWorkDefinitions() {
        return Stream.of(this);
    }

    protected String workContextValue(String key) {
        if (System.getenv(key) != null) {
            return System.getenv(key);
        }
        // This might happen in tests
        return System.getProperty(key);
    }
}
