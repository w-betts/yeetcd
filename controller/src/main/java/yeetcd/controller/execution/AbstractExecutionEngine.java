package yeetcd.controller.execution;

import java.time.Duration;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.function.Predicate;
import java.util.function.Supplier;

public abstract class AbstractExecutionEngine implements ExecutionEngine {
    private static final AtomicInteger threadCount = new AtomicInteger();
    protected final ScheduledExecutorService executor = new ScheduledThreadPoolExecutor(
        1,
        runnable -> new Thread(runnable, this.getClass().getSimpleName() + "-" + threadCount.incrementAndGet())
    );

    protected <T> CompletableFuture<T> doAsync(Supplier<T> work) {
        CompletableFuture<T> completableFuture = new CompletableFuture<>();
        executor.submit(() -> {
            try {
                completableFuture.complete(work.get());
            }
            catch (Throwable throwable) {
                completableFuture.completeExceptionally(throwable);
            }

        });
        return completableFuture;
    }

    protected <T> CompletableFuture<T> doAsyncUntil(Supplier<T> work, Predicate<T> workCheck, Duration timeLimit) {
        CompletableFuture<T> completableFuture = new CompletableFuture<>();
        doAsyncWithRetries(completableFuture, () -> {
            T t = work.get();
            if (!workCheck.test(t)) {
                throw new IllegalStateException("Work result condition not met yet");
            }
            return t;
        }, System.currentTimeMillis(), timeLimit, 100, 15000, TimeUnit.MILLISECONDS, true);
        return completableFuture;
    }

    private <T> void doAsyncWithRetries(CompletableFuture<T> completableFuture, Supplier<T> work, long start, Duration timeLimit, long nextBackOff, long maxBackOff, TimeUnit backoffTimeUnit, boolean firstAttempt) {
        executor.schedule(() -> {
            try {
                completableFuture.complete(work.get());
            }
            catch (Throwable throwable) {
                if (System.currentTimeMillis() - start > timeLimit.toMillis()) {
                    completableFuture.completeExceptionally(throwable);
                }
                else {
                    doAsyncWithRetries(completableFuture, work, start, timeLimit, Math.min(firstAttempt ? nextBackOff : nextBackOff * 2, maxBackOff), maxBackOff, backoffTimeUnit, false);
                }
            }

        }, firstAttempt ? 0 : nextBackOff, backoffTimeUnit);
    }
}
