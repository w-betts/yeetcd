package yeetcd.controller.pipeline;

import yeetcd.controller.source.SourceLanguage;
import yeetcd.protocol.pipeline.PipelineOuterClass;

import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

public record Pipeline(String name, Parameters parameters, WorkContext workContext, List<Work> finalWork, PipelineMetadata pipelineMetadata) {

    public Pipeline(String name, List<Work> finalWork, PipelineMetadata pipelineMetadata) {
        this(name, Parameters.empty(), WorkContext.empty(), finalWork, pipelineMetadata);
    }

    public Pipeline withArguments(Arguments arguments) {
        return new Pipeline(
                name,
                parameters,
                arguments.asValidatedWorkContext(parameters).mergeInto(workContext),
                finalWork,
                pipelineMetadata
        );
    }

    public static Pipeline fromProtobuf(PipelineOuterClass.Pipeline pipelineProtobuf, String builtSourceImage, SourceLanguage sourceLanguage) {
        WorkContext pipelineWorkContext = new WorkContext(pipelineProtobuf.getWorkContextMap());
        return new Pipeline(
                pipelineProtobuf.getName(),
                new Parameters(pipelineProtobuf
                        .getParametersMap()
                        .entrySet().stream()
                        .collect(Collectors.toMap(Map.Entry::getKey, entry -> Parameter.fromProtobuf(entry.getValue())))
                ),
                pipelineWorkContext,
                pipelineProtobuf.getFinalWorkList().stream().map(Work::fromProtobuf).toList(),
                new PipelineMetadata(pipelineProtobuf.getName(), builtSourceImage, sourceLanguage)
        );
    }
}
