package yeetcd.test;

import lombok.EqualsAndHashCode;

import java.util.*;
import java.util.stream.Collectors;

@EqualsAndHashCode
public final class FakeWorkExecutionStage {
    private final Set<FakeWorkExecution> workExecutions;

    private FakeWorkExecutionStage(Set<FakeWorkExecution> workExecutions) {
        this.workExecutions = workExecutions;
    }

    public Set<FakeWorkExecution> getWorkExecutions() {
        return workExecutions;
    }

    public static Builder builder(Set<FakeWorkExecution> workExecutions) {
        return new Builder(workExecutions);
    }

    public static class Builder {
        private final Set<FakeWorkExecution> workExecutions;

        private Builder(Set<FakeWorkExecution> workExecutions) {
            this.workExecutions = workExecutions;
        }

        public FakeWorkExecutionStage build() {
            return new FakeWorkExecutionStage(workExecutions);
        }
    }

    @Override
    public String toString() {
        List<String> workExecutionStrings = workExecutions.stream().map(FakeWorkExecution::toString).toList();
        List<List<String>> workExecutionsLines = new ArrayList<>(workExecutionStrings.stream().map(workExecutionString -> Arrays.stream(workExecutionString.split("\n")).map(" %s "::formatted).toList()).toList());
        workExecutionsLines.sort(Comparator.comparing(a -> String.join("", a)));
        int maxLines = workExecutionsLines.stream().map(List::size).max(Comparator.naturalOrder()).orElse(0);
        int maxLineLength = workExecutionsLines.stream().flatMap(List::stream).map(String::length).max(Comparator.naturalOrder()).orElse(0);
        int adjustedMaxLineLength = maxLineLength % 2 == 1 ? maxLineLength + 1 : maxLineLength;
        String workBorder = "@";
        String workSeparator = "%s   %s".formatted(workBorder, workBorder);
        String stageBorder = " ";
        String stageInnerBorder = "    ";
        int totalWidth = stageBorder.length() + stageInnerBorder.length() + workBorder.length() + workExecutions.size() * adjustedMaxLineLength + (workExecutions.size() - 1) * workSeparator.length() + workBorder.length() + stageInnerBorder.length() + stageBorder.length();
        String stageBorderLine = stageBorder.repeat(totalWidth);
        String stageInnerBorderLine = "%s%s%s%s%s".formatted(stageBorder, stageInnerBorder, " ".repeat(totalWidth - (stageInnerBorder.length() * 2 + stageBorder.length() * 2)), stageInnerBorder, stageBorder);
        String workTopBottomBorderLine = "%s%s%s%s%s%s%s".formatted(stageBorder, stageInnerBorder, workBorder, workExecutions.stream().map(it -> workBorder.repeat(adjustedMaxLineLength)).collect(Collectors.joining(workSeparator)), workBorder, stageInnerBorder, stageBorder);

        StringBuilder stringBuilder = new StringBuilder();
        stringBuilder.append(stageBorderLine);
        stringBuilder.append("\n");
        stringBuilder.append(stageInnerBorderLine);
        stringBuilder.append("\n");
        stringBuilder.append(workTopBottomBorderLine);
        stringBuilder.append("\n");

        for (int i = 0; i < maxLines; i++) {
            stringBuilder.append(stageBorder);
            stringBuilder.append(stageInnerBorder);
            stringBuilder.append(workBorder);
            final int index = i;
            stringBuilder.append(workExecutionsLines.stream()
                .map(workExecutionLines -> {
                    if (index < workExecutionLines.size()) {
                        String workExecutionLine = workExecutionLines.get(index);
                        int paddingLength = adjustedMaxLineLength - workExecutionLine.length();
                        return workExecutionLine + " ".repeat(paddingLength);
                    }
                    else {
                        return " ".repeat(adjustedMaxLineLength);
                    }
                })
                .collect(Collectors.joining(workSeparator))
            );
            stringBuilder.append(workBorder);
            stringBuilder.append(stageInnerBorder);
            stringBuilder.append(stageBorder);
            stringBuilder.append("\n");
        }

        stringBuilder.append(workTopBottomBorderLine);
        stringBuilder.append("\n");
        stringBuilder.append(stageInnerBorderLine);
        stringBuilder.append("\n");
        stringBuilder.append(stageBorderLine);

        return stringBuilder.toString();
    }

    public static String toString(Collection<FakeWorkExecutionStage> stages) {
        String rawStagesString = stages.stream().map(FakeWorkExecutionStage::toString).collect(Collectors.joining("\n|\n"));
        String[] allLines = rawStagesString.split("\n");
        int maxLineLength = Arrays.stream(allLines)
            .flatMap(stage -> Arrays.stream(stage.split("\n")))
            .map(String::length)
            .max(Comparator.naturalOrder())
            .orElse(0);
        StringBuilder stringBuilder = new StringBuilder();
        Arrays.stream(allLines).forEach(stageString -> {
            int adjustedMaxLineLength = maxLineLength % 2 == 1 ? maxLineLength + 1 : maxLineLength;
            int paddingLength = (adjustedMaxLineLength - stageString.length()) / 2;
            Arrays.stream(stageString.split("\n")).forEach(line -> {
                stringBuilder.append(" ".repeat(paddingLength));
                stringBuilder.append(line);
                stringBuilder.append(" ".repeat(paddingLength));
                stringBuilder.append("\n");
            });
        });
        return stringBuilder.toString();
    }
}
