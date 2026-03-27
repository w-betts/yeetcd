package yeetcd.controller.execution;

import java.io.File;

public record OnDiskMountInput(File directory) implements MountInput {
}
