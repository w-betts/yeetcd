package yeetcd.controller.pipeline;

import yeetcd.controller.execution.JobStreams;

import java.io.File;

public record WorkResult(WorkStatus workStatus, File outputDirectoriesParent, JobStreams jobStreams) {
}
