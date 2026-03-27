package yeetcd.controller.execution;

import java.io.File;

public record JobResult(int exitCode, File outputDirectoriesParent) {
}
