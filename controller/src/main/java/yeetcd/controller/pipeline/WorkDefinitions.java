package yeetcd.controller.pipeline;

import yeetcd.protocol.pipeline.PipelineOuterClass;

import java.util.stream.Collectors;

public class WorkDefinitions {

    static WorkDefinition fromWorkProtobuf(PipelineOuterClass.Work work) {
        if (work.hasContainerisedWorkDefinition()) {
            return new ContainerisedWorkDefinition(
                    work.getContainerisedWorkDefinition().getImage(),
                    work.getContainerisedWorkDefinition().getCmdList().stream().toList()
            );
        } else if (work.hasCustomWorkDefinition()) {
            return new CustomWorkDefinition(work.getCustomWorkDefinition().getExecutionId());
        } else if (work.hasCompoundWorkDefinition()) {
            return new CompoundWorkDefinition(
                    work.getCompoundWorkDefinition().getFinalWorkList().stream().map(Work::fromProtobuf).collect(Collectors.toList())
            );
        } else if (work.hasDynamicWorkGeneratingWorkDefinition()) {
            return new DynamicWorkGeneratingWorkDefinition(work.getDynamicWorkGeneratingWorkDefinition().getExecutionId());
        } else {
            throw new IllegalArgumentException("unsupported task protobuf object");
        }
    }
}
