package yeetcd.controller.source;

import yeetcd.controller.execution.ImageBase;

public enum SourceLanguage {
    JAVA(ImageBase.JAVA, "yeetcd.sdk.GeneratedPipelineDefinitions") {
        @Override
        public String[] getCustomTaskRunnerCmd(String pipelineName, String taskName) {
            return new String[]{"yeetcd.sdk.GeneratedCustomWorkRunner", pipelineName, taskName};
        }
    };

    private final ImageBase imageBase;

    private final String generatePipelineDefinitionsCmd;

    SourceLanguage(ImageBase imageBase, String generatePipelineDefinitionsCmd) {
        this.imageBase = imageBase;
        this.generatePipelineDefinitionsCmd = generatePipelineDefinitionsCmd;
    }

    public ImageBase getImageBase() {
        return imageBase;
    }

    public String getGeneratePipelineDefinitionsCmd() {
        return generatePipelineDefinitionsCmd;
    }

    public abstract String[] getCustomTaskRunnerCmd(String pipelineName, String taskName);
}
