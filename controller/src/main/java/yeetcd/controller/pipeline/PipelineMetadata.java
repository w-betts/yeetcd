package yeetcd.controller.pipeline;

import yeetcd.controller.source.SourceLanguage;

public record PipelineMetadata(String pipelineName, String builtSourceImage, SourceLanguage sourceLanguage) {
}
