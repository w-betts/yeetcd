package yeetcd.controller.source;

import java.io.Closeable;
import java.io.File;
import java.util.Map;

public record SourceExtractionResult(Source source, File directory, Map<String, YeetcdDefinition> yeetcdDefinitions) implements Closeable {
    @Override
    public void close() {
        if (!directory.delete()) {

        }
    }
}
