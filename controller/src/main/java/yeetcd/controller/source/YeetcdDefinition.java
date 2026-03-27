package yeetcd.controller.source;

import java.util.List;

public record YeetcdDefinition(String name, SourceLanguage language, String buildImage, String buildCmd, List<ArtifactDefinition> artifacts) {
}
