package yeetcd.test;

import yeetcd.sdk.Work;
import yeetcd.sdk.WorkContext;
import com.fasterxml.jackson.annotation.JsonAutoDetect;
import com.fasterxml.jackson.annotation.PropertyAccessor;
import com.fasterxml.jackson.databind.*;
import com.fasterxml.jackson.databind.node.ObjectNode;
import lombok.EqualsAndHashCode;
import lombok.SneakyThrows;

import java.util.Collections;
import java.util.Map;

@EqualsAndHashCode
public final class FakeSimpleWorkExecution implements FakeWorkExecution {

    private final Work work;
    private final WorkContext workContext;
    private final Map<String, String> envVars;
    private final Map<String, byte[]> inputFiles;
    private final FakeWorkStatus status;

    private final Map<String, byte[]> exportedFiles;
    private final String stdOut;

    private static final ObjectMapper OBJECT_MAPPER = new ObjectMapper()
        .setVisibility(PropertyAccessor.ALL, JsonAutoDetect.Visibility.NONE)
        .setVisibility(PropertyAccessor.FIELD, JsonAutoDetect.Visibility.ANY)
        .configure(SerializationFeature.ORDER_MAP_ENTRIES_BY_KEYS, true);

    private static final ObjectWriter OBJECT_WRITER = OBJECT_MAPPER.writerWithDefaultPrettyPrinter();

    private FakeSimpleWorkExecution(Work work, WorkContext workContext, Map<String, String> envVars, Map<String, byte[]> inputFiles, Map<String, byte[]> exportedFiles, String stdOut, FakeWorkStatus status) {
        this.work = work;
        this.workContext = workContext;
        this.envVars = envVars;
        this.inputFiles = inputFiles;
        this.exportedFiles = exportedFiles;
        this.stdOut = stdOut;
        this.status = status;
    }

    public FakeWorkStatus getStatus() {
        return status;
    }

    public Map<String, byte[]> getExportedFiles() {
        return exportedFiles;
    }

    public String getStdOut() {
        return stdOut;
    }

    public static Builder builder(Work work) {
        return new Builder(work);
    }

    public static class Builder {
        private final Work work;

        private WorkContext workContext = WorkContext.empty();

        private Map<String, String> envVars = Collections.emptyMap();

        private Map<String, byte[]> inputFiles = Collections.emptyMap();

        private Map<String, byte[]> exportedFiles = Collections.emptyMap();
        private String stdOut = "";

        private FakeWorkStatus status = FakeWorkStatus.SUCCESS;

        private Builder(Work work) {
            this.work = work;
        }

        public Builder workContext(WorkContext workContext) {
            this.workContext = workContext;
            return this;
        }

        public Builder envVars(Map<String, String> envVars) {
            this.envVars = envVars;
            return this;
        }

        public Builder inputFiles(Map<String, byte[]> inputFiles) {
            this.inputFiles = inputFiles;
            return this;
        }

        public Builder exportedFiles(Map<String, byte[]> exportedFiles) {
            this.exportedFiles = exportedFiles;
            return this;
        }

        public Builder stdOut(String stdOut) {
            this.stdOut = stdOut;
            return this;
        }

        public Builder status(FakeWorkStatus status) {
            this.status = status;
            return this;
        }

        public FakeSimpleWorkExecution build() {
            return new FakeSimpleWorkExecution(work, workContext, envVars, inputFiles, exportedFiles, stdOut, status);
        }
    }

    @Override
    @SneakyThrows
    public String toString() {
        try {
            JsonNode value = OBJECT_MAPPER.valueToTree(this);
            ObjectNode workNode = (ObjectNode) value.get("work");
            workNode.remove("previousWork");
            return OBJECT_WRITER.writeValueAsString(value);
        }
        catch (IllegalArgumentException ex) {
            throw ex;
        }
    }
}
