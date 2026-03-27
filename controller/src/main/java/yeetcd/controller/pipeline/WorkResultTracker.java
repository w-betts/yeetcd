package yeetcd.controller.pipeline;

import yeetcd.controller.execution.MountInput;
import yeetcd.controller.execution.OnDiskMountInput;
import lombok.SneakyThrows;

import java.util.Collections;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.CompletableFuture;
import java.util.function.Supplier;

public class WorkResultTracker {

    private final Map<Work, CompletableFuture<WorkResult>> workResultMap = new HashMap<>();

    // This is close to what we'd get from a ConcurrentHashMap, but that suffers from issues with recursive update
    public synchronized CompletableFuture<WorkResult> getOrExecute(Work key, Supplier<CompletableFuture<WorkResult>> execute) {
        CompletableFuture<WorkResult> workResultCompletableFuture = workResultMap.get(key);
        if (workResultCompletableFuture == null) {
            workResultCompletableFuture = execute.get();
            workResultMap.put(key, workResultCompletableFuture);
        }
        return workResultCompletableFuture;
    }
    public synchronized Map<Work, CompletableFuture<WorkResult>> getWorkResultMap() {
        return Collections.unmodifiableMap(workResultMap);
    }

    @SneakyThrows
    MountInput outputDirectoriesMountInput(PreviousWork previousWork) {
        return new OnDiskMountInput(getWorkResultMap().get(previousWork.work()).get().outputDirectoriesParent());
    }

    @SneakyThrows
    byte[] stdOut(PreviousWork previousWork) {
        return getWorkResultMap().get(previousWork.work()).get().jobStreams().getStdOut();
    }
}