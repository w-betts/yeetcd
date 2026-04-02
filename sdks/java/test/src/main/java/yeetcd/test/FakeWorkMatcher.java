package yeetcd.test;

import yeetcd.sdk.Work;
import yeetcd.sdk.WorkContext;
import lombok.EqualsAndHashCode;
import lombok.ToString;

@EqualsAndHashCode
@ToString
public final class FakeWorkMatcher {

    private final Work work;
    private final WorkContext workContext;

    private FakeWorkMatcher(Work work, WorkContext workContext) {
        this.work = work;
        this.workContext = workContext;
    }

    boolean matches(Work work, WorkContext workContext) {
        return workMatches(work) && workContextMatches(workContext);
    }

    private boolean workMatches(Work work) {
        return this.work == null || this.work.equals(work);
    }

    private boolean workContextMatches(WorkContext workContext) {
        return this.workContext == null || this.workContext.equals(workContext);
    }

    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private Work work;
        private WorkContext workContext;

        public Builder work(Work work) {
            this.work = work;
            return this;
        }

        public Builder workContext(WorkContext workContext) {
            this.workContext = workContext;
            return this;
        }

        public FakeWorkMatcher build() {
            return new FakeWorkMatcher(work, workContext);
        }

    }

}
