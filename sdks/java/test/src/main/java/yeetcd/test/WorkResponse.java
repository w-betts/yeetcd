package yeetcd.test;

public record WorkResponse(
    int exitCode,
    String stdout,
    String stderr
) {
    public static WorkResponse success() {
        return new WorkResponse(0, "", "");
    }

    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private int exitCode = 0;
        private String stdout = "";
        private String stderr = "";

        public Builder exitCode(int exitCode) {
            this.exitCode = exitCode;
            return this;
        }

        public Builder stdout(String stdout) {
            this.stdout = stdout != null ? stdout : "";
            return this;
        }

        public Builder stderr(String stderr) {
            this.stderr = stderr != null ? stderr : "";
            return this;
        }

        public WorkResponse build() {
            return new WorkResponse(exitCode, stdout, stderr);
        }
    }
}
