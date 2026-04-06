package yeetcd.test;

import yeetcd.sdk.CustomWorkDefinition;
import yeetcd.sdk.DynamicWorkGeneratingWorkDefinition;

public interface Behavior {
    
    ContainerisedWorkBehaviorBuilder containerisedWork(String image);
    
    CustomWorkBehaviorBuilder customWork(CustomWorkDefinition instance);
    
    DynamicWorkBehaviorBuilder dynamicWork();
    
    DefaultWorkBehaviorBuilder defaultContainerisedWork();
    
    DefaultWorkBehaviorBuilder defaultCustomWork();
    
    PipelineTestRun build();
}
