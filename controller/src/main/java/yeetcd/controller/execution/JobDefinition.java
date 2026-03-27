package yeetcd.controller.execution;

import lombok.Builder;

import java.util.Map;

@Builder
public record JobDefinition(
        String image,
        String[] cmd,
        String workingDir,
        Map<String, String> environment,
        Map<String, MountInput> inputFilePaths,
        Map<String, String> outputDirectoryPaths,
        JobStreams jobStreams

) {
}
