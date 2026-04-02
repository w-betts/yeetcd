package yeetcd.test;

import lombok.EqualsAndHashCode;
import lombok.ToString;

import java.util.Collections;
import java.util.Map;

@EqualsAndHashCode
@ToString
public final class FakeWorkResult {

    private final FakeWorkStatus status;
    private final String stdOut;
    private final Map<String, byte[]> exportedFiles;

    private FakeWorkResult(FakeWorkStatus status, String stdOut, Map<String, byte[]> exportedFiles) {
        this.status = status;
        this.stdOut = stdOut;
        this.exportedFiles = exportedFiles;
    }

    public FakeWorkStatus getStatus() {
        return status;
    }

    public String getStdOut() {
        return stdOut;
    }

    public Map<String, byte[]> getExportedFiles() {
        return exportedFiles;
    }

    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private FakeWorkStatus status = FakeWorkStatus.SUCCESS;
        private String stdOut = "";
        private Map<String, byte[]> exportedFiles = Collections.emptyMap();

        public Builder status(FakeWorkStatus status) {
            this.status = status;
            return this;
        }

        public Builder stdOut(String stdOut) {
            this.stdOut = stdOut;
            return this;
        }

        public Builder exportedFiles(Map<String, byte[]> exportedFiles) {
            this.exportedFiles = exportedFiles;
            return this;
        }

        public FakeWorkResult build() {
            return new FakeWorkResult(status, stdOut, exportedFiles);
        }
    }
}
