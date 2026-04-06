package yeetcd.test;

import yeetcd.sdk.CustomWorkDefinition;
import yeetcd.sdk.DynamicWorkGeneratingWorkDefinition;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class BehaviorChain implements Behavior {
    
    private final PipelineTestRun.Builder builder;
    private final List<ContainerisedBehavior> containerisedBehaviors = new ArrayList<>();
    private final Map<String, CustomWorkBehavior> customWorkBehaviors = new HashMap<>();
    private DynamicWorkGeneratingWorkDefinition dynamicWorkInstance;
    private WorkResponse defaultContainerisedResponse = new WorkResponse(0, "", "");
    private WorkResponse defaultCustomResponse = new WorkResponse(0, "", "");

    BehaviorChain(PipelineTestRun.Builder builder) {
        this.builder = builder;
    }

    @Override
    public ContainerisedWorkBehaviorBuilder containerisedWork(String image) {
        return new ContainerisedWorkBehaviorBuilder(image, this);
    }

    @Override
    public CustomWorkBehaviorBuilder customWork(CustomWorkDefinition instance) {
        String executionId = instance.getExecutionId();
        return new CustomWorkBehaviorBuilder(instance, executionId, this);
    }

    @Override
    public DynamicWorkBehaviorBuilder dynamicWork() {
        return new DynamicWorkBehaviorBuilder(this);
    }

    @Override
    public DefaultWorkBehaviorBuilder defaultContainerisedWork() {
        return new DefaultWorkBehaviorBuilder(true, this);
    }

    @Override
    public DefaultWorkBehaviorBuilder defaultCustomWork() {
        return new DefaultWorkBehaviorBuilder(false, this);
    }

    void addContainerisedWorkBehavior(String image, ContainerisedWorkBehaviorBuilder.ReplacementImage replacement, WorkResponse response) {
        containerisedBehaviors.add(new ContainerisedBehavior(image, replacement, response));
    }

    void addCustomWorkBehavior(String executionId, WorkResponse response, CustomWorkDefinition replacementInstance, boolean runOriginal) {
        customWorkBehaviors.put(executionId, new CustomWorkBehavior(response, replacementInstance, runOriginal));
    }

    void setDynamicWorkBehavior(DynamicWorkGeneratingWorkDefinition instance) {
        this.dynamicWorkInstance = instance;
    }

    void setDefaultContainerisedResponse(WorkResponse response) {
        this.defaultContainerisedResponse = response;
    }

    void setDefaultCustomResponse(WorkResponse response) {
        this.defaultCustomResponse = response;
    }

    public List<ContainerisedBehavior> getContainerisedBehaviors() {
        return containerisedBehaviors;
    }

    public Map<String, CustomWorkBehavior> getCustomWorkBehaviors() {
        return customWorkBehaviors;
    }

    public DynamicWorkGeneratingWorkDefinition getDynamicWorkInstance() {
        return dynamicWorkInstance;
    }

    public WorkResponse getDefaultContainerisedResponse() {
        return defaultContainerisedResponse;
    }

    public WorkResponse getDefaultCustomResponse() {
        return defaultCustomResponse;
    }

    public boolean hasDynamicWork() {
        return dynamicWorkInstance != null;
    }

    @Override
    public PipelineTestRun build() {
        return builder.build();
    }

    public void registerWith(MockServer mockServer) {
        for (ContainerisedBehavior cb : containerisedBehaviors) {
            mockServer.registerContainerisedBehavior(cb.image(), cb.response());
        }
        
        for (Map.Entry<String, CustomWorkBehavior> entry : customWorkBehaviors.entrySet()) {
            mockServer.registerCustomWorkBehavior(entry.getKey(), entry.getValue().response());
        }
        
        mockServer.setDefaultContainerisedResponse(defaultContainerisedResponse);
        mockServer.setDefaultCustomResponse(defaultCustomResponse);
    }

    public record ContainerisedBehavior(
        String image,
        ContainerisedWorkBehaviorBuilder.ReplacementImage replacement,
        WorkResponse response
    ) {}

    public record CustomWorkBehavior(
        WorkResponse response,
        CustomWorkDefinition replacementInstance,
        boolean runOriginal
    ) {}
}
