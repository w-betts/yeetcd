package yeetcd.controller.source;

import yeetcd.controller.ArchiveUtils;
import lombok.SneakyThrows;
import org.junit.jupiter.api.Test;

import java.io.File;
import java.io.IOException;
import java.util.Arrays;
import java.util.List;
import java.util.Map;

import static org.hamcrest.CoreMatchers.*;
import static org.hamcrest.MatcherAssert.assertThat;

public class SourceExtractorTest {

    @Test
    public void shouldExtractTheJavaSampleApplication() throws IOException {
        // given
        SourceExtractor sourceExtractor = new SourceExtractor();

        // when
        try (SourceExtractionResult sourceExtractionResult = sourceExtractor.extract(new Source("test", ArchiveUtils.projectZip()))) {
            // then
            File[] files = sourceExtractionResult.directory().listFiles();
            assertThat(files, notNullValue());
            assertThat(files.length, equalTo(1));
            assertThat(files[0].getName(), equalTo("yeetcd"));

            File[] childFiles = files[0].listFiles();
            assertThat(childFiles, notNullValue());
            assertThat(Arrays.stream(childFiles).map(File::getName).toList(), hasItems("yeetcd.yaml", "java-sample", "controller"));

            // and
            assertThat(sourceExtractionResult.yeetcdDefinitions(), equalTo(Map.of(
                "yeetcd/yeetcd.yaml", new YeetcdDefinition(
                            "java-sample",
                            SourceLanguage.JAVA,
                            "maven:3.9.9-eclipse-temurin-17",
                            "mvn -am -pl java-sample clean test package dependency:copy-dependencies",
                            List.of(
                                    new ArtifactDefinition("classes", "java-sample/target/classes"),
                                    new ArtifactDefinition("dependencies", "java-sample/target/dependency")
                            )
                    )
            )));
        }

    }

    @SneakyThrows
    public static SourceExtractionResult givenLocalProjectExtraction() {
        SourceExtractor sourceExtractor = new SourceExtractor();

        // when
        SourceExtractionResult extract = sourceExtractor.extract(new Source("test", ArchiveUtils.projectZip()));
        extract.directory().deleteOnExit();
        return extract;
    }
}
