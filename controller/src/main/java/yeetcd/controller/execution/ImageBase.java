package yeetcd.controller.execution;

import java.util.List;
import java.util.stream.Collectors;

public enum ImageBase {
    // TODO - shouldn't the user control this and have the option of a custom image
    JAVA("maven:3.9.9-eclipse-temurin-17") {
        @Override
        String[] entryPoint(String artifactParentDirectoryPath, List<String> artifactDefinitionNames) {
            String classPathArgs = artifactDefinitionNames
                .stream()
                .map(name -> "%s/%s".formatted(artifactParentDirectoryPath, name))
                .map(path -> "%s:%s/*".formatted(path, path))
                .collect(Collectors.joining(":"));
            return new String[]{"java", "-cp", classPathArgs};
        }
    };

    private final String baseImage;

    ImageBase(String baseImage) {
        this.baseImage = baseImage;
    }

    public String getBaseImage() {
        return baseImage;
    }

    @SuppressWarnings("SameParameterValue")
    abstract String[] entryPoint(String artifactParentDirectoryPath, List<String> artifactDefinitionNames);
}
