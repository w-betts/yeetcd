package yeetcd.test;

public record MatchedBehavior(
    WorkBehaviorType type,
    String matchKey,
    WorkExecution execution,
    WorkResponse response
) {}
