package yeetcd.controller.source;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import lombok.SneakyThrows;

import java.io.ByteArrayInputStream;
import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.util.HashMap;
import java.util.Map;

import static com.fasterxml.jackson.dataformat.yaml.YAMLGenerator.Feature.MINIMIZE_QUOTES;

public class SourceExtractor {

    private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper(new YAMLFactory().enable(MINIMIZE_QUOTES)).setSerializationInclusion(JsonInclude.Include.NON_NULL);

    public SourceExtractionResult extract(Source source) throws IOException {
        File destDir = Files.createTempDirectory("extraction").toFile();
        destDir.deleteOnExit();
        Map<String, YeetcdDefinition> yeetcdDefinitions = new HashMap<>();
        ZipExtractor.extract(new ByteArrayInputStream(source.zip()), destDir, new ZipExtractor.FileHandler(
                (fileName) -> fileName.equals("yeetcd.yaml"),
                (handledFile -> yeetcdDefinitions.put(handledFile.parent(), getYeetcdDefinition(handledFile))
        )));
        return new SourceExtractionResult(source, destDir, yeetcdDefinitions);
    }

    @SneakyThrows
    private static YeetcdDefinition getYeetcdDefinition(ZipExtractor.HandledFile handledFile) {
        return OBJECT_MAPPER.readValue(handledFile.contents(), YeetcdDefinition.class);
    }

}
