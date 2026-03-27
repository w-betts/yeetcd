package yeetcd.sdk;

import yeetcd.protocol.pipeline.PipelineOuterClass;
import yeetcd.sdk.condition.Condition;
import yeetcd.sdk.condition.Conditions;
import yeetcd.sdk.condition.PreviousWorkStatusCondition;
import lombok.EqualsAndHashCode;
import lombok.ToString;
import org.apache.commons.codec.digest.DigestUtils;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Stream;

@EqualsAndHashCode
@ToString
public final class Work {

    private static final ConcurrentHashMap<WorkIdKey, String> workIdLookup = new ConcurrentHashMap<>();
    private final String description;
    private final WorkDefinition workDefinition;

    private final WorkContext workContext;

    private final WorkOutputPath[] workOutputPaths;
    private final PreviousWork[] previousWork;
    private final Condition condition;
    private Work(String description, WorkDefinition workDefinition, WorkContext workContext, WorkOutputPath[] workOutputPaths, PreviousWork[] previousWork, Condition condition) {
        this.description = description;
        this.workContext = workContext;
        this.workOutputPaths = workOutputPaths;
        this.workDefinition = workDefinition;
        this.previousWork = previousWork;
        this.condition = condition;
    }

    public static Builder builder(String description, WorkDefinition workDefinition) {
        return new Builder(description, workDefinition);
    }

    public String getDescription() {
        return description;
    }

    public WorkDefinition getWorkDefinition() {
        return workDefinition;
    }

    public WorkContext getWorkContext() {
        return workContext;
    }

    public WorkOutputPath[] getWorkOutputPaths() {
        return workOutputPaths;
    }

    public PreviousWork[] getPreviousWork() {
        return previousWork;
    }

    public Condition getCondition() {
        return condition;
    }

    PipelineOuterClass.Work toProtobuf(WorkContext containingContext) {
        PipelineOuterClass.Work.Builder builder = PipelineOuterClass.Work
                .newBuilder()
                .setId(id(containingContext))
                .setDescription(description)
                .addAllOutputPaths(Arrays.stream(workOutputPaths).map(WorkOutputPath::toProtobuf).toList())
                .putAllWorkContext(workContext.getWorkContextMap())
                .addAllPreviousWork(Arrays.stream(previousWork).map(work -> work.toProtobuf(containingContext)).toList());
        workDefinition.applyTo(mergedContext(containingContext), builder);
        condition.applyTo(builder);
        return builder.build();
    }

    String id(WorkContext containingContext) {
        WorkIdKey key = new WorkIdKey(this, mergedContext(containingContext));
        return workIdLookup.computeIfAbsent(key, ignore -> DigestUtils.sha256Hex(UUID.randomUUID().toString()));
    }

    Stream<NativeWorkDefinition> nativeWorkDefinitions() {
        return Stream.concat(
                Arrays.stream(previousWork).flatMap(work -> work.getWork().nativeWorkDefinitions()),
                workDefinition.nativeWorkDefinitions()
        );
    }

    private WorkContext mergedContext(WorkContext containingContext) {
        return workContext.mergeInto(containingContext);
    }

    public static class Builder {
        private final String description;
        private final WorkDefinition workDefinition;
        private WorkContext workContext = WorkContext.empty();
        private WorkOutputPath[] workOutputPaths = new WorkOutputPath[]{};
        private PreviousWork[] previousWork = new PreviousWork[]{};
        private Condition condition = Conditions.previousWorkStatusCondition(PreviousWorkStatusCondition.Status.SUCCESS);

        public Builder(String description, WorkDefinition workDefinition) {
            this.description = description;
            this.workDefinition = workDefinition;
        }

        public Builder workContext(WorkContext workContext) {
            this.workContext = workContext;
            return this;
        }


        public Builder workOutputPaths(WorkOutputPath... workOutputPaths) {
            this.workOutputPaths = workOutputPaths;
            return this;
        }


        public Builder previousWork(PreviousWork... previousWork) {
            this.previousWork = previousWork;
            return this;
        }

        public Builder condition(Condition condition) {
            this.condition = condition;
            return this;
        }

        public Work build() {
            return new Work(description, workDefinition, workContext, workOutputPaths, previousWork, condition);
        }
    }

    private record WorkIdKey(Work work, WorkContext workContext) {

    }
}
