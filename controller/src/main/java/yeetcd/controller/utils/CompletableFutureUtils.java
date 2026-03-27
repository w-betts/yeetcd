package yeetcd.controller.utils;

import lombok.SneakyThrows;

import java.util.Collection;
import java.util.List;
import java.util.concurrent.CompletableFuture;

public class CompletableFutureUtils {

    public static <T> CompletableFuture<List<T>> zip(Collection<CompletableFuture<T>> completableFutures) {
        return CompletableFuture
                .allOf(completableFutures.toArray(CompletableFuture[]::new))
                .thenApply(ignored -> completableFutures.stream().map(CompletableFutureUtils::sneakyGet).toList());
    }

    @SneakyThrows
    private static <T> T sneakyGet(CompletableFuture<T> sourceBuildResultCompletableFuture) {
        return sourceBuildResultCompletableFuture.get();
    }
}
