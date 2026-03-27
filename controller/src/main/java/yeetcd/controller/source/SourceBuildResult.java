package yeetcd.controller.source;

import java.io.File;

public record SourceBuildResult(YeetcdDefinition yeetcdDefinition, File outputDirectoriesParent) {
}
