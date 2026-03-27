package yeetcd.test;

import lombok.EqualsAndHashCode;
import lombok.ToString;

@EqualsAndHashCode
@ToString
public final class FakeWorkMatcherResult {
    private final FakeWorkMatcher workMatcher;

    private final FakeWorkResult workResult;
    private FakeWorkMatcherResult(FakeWorkMatcher workMatcher, FakeWorkResult workResult) {
        this.workMatcher = workMatcher;
        this.workResult = workResult;
    }

    public FakeWorkMatcher getWorkMatcher() {
        return workMatcher;
    }

    public FakeWorkResult getWorkResult() {
        return workResult;
    }

    public static Builder builder(FakeWorkMatcher workMatcher, FakeWorkResult workResult) {
        return new Builder(workMatcher, workResult);
    }

    public static class Builder {
        private final FakeWorkMatcher workMatcher;
        private final FakeWorkResult workResult;

        public Builder(FakeWorkMatcher workMatcher, FakeWorkResult workResult) {
            this.workMatcher = workMatcher;
            this.workResult = workResult;
        }

        public FakeWorkMatcherResult build() {
            return new FakeWorkMatcherResult(workMatcher, workResult);
        }
    }

}
